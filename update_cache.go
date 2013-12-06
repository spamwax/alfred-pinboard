package main

import (
    "bitbucket.org/listboss/go-alfred"
    "bufio"
    "bytes"
    "encoding/gob"
    "encoding/xml"
    "errors"
    "io/ioutil"
    "net/http"
    "net/url"
    // "fmt"
    "os"
    "path"
    "strings"
    "time"
)

var data = `<?xml version="1.0" encoding="UTF-8" ?>
<posts user="hamid">
<post href="http://vimeo.com/37562944#" time="2013-03-15T07:07:48Z" description="See Through 3D Desktop, on Vimeo" extended="" tag="3d ui screen monitor desktop leapmotion" hash="62e78977500a21754caa741fcafa3695" meta="ddf8de5b3a6d2cd6d7562c2ae27902d4"  shared="yes"  />
<post href="http://www.itefix.no/phpws/index.php?module=phpwsbb&amp;PHPWSBB_MAN_OP=view&amp;PHPWS_MAN_ITEMS[]=343" time="2013-03-09T11:38:25Z" description="Icon Sets and Designers Directory - Iconify.info - A comprehensive curated gallery of icons and icon-related resources" extended="" tag="icon font free vector-graphic svg" hash="5b3f7ff0257a582531c8f33e703f76a5" meta="b38c248e539cf132ce8c09207931f33a"  shared="no"  />
<post href="http://durak.org/cvswebsites/howto-cvs/node37.html" time="2006-03-03T20:55:18Z" description="CVS Server Setup" extended="CVS Server Setup" tag="Learn screen Linux CVS" hash="f99a793264f1a3ef8e8ecdc7d1f7c154" meta="7da1d359a76955fcbd2cbce622483f63"  shared="no"  />
</posts>`

type pinboardResponse struct {
    Result     pResult `xml:"result"`
    Posts      Posts   `xml:"posts"`
    UpdateTime pUpdate `xml:"update"`
}

type pResult struct {
    Code string `xml:"code,attr"`
}

type pUpdate struct {
    Update time.Time `xml:"time,attr"`
}

type Posts struct {
    XMLName xml.Name `xml:"posts"`
    Pins    []Link   `xml:"post"`
}

type Link struct {
    XMLName xml.Name  `xml:"post"`
    Url     string    `xml:"href,attr"`
    Desc    string    `xml:"description,attr"`
    Notes   string    `xml:"extended,attr"`
    Time    time.Time `xml:"time,attr"`
    Hash    string    `xml:"hash,attr"`
    Shared  bool      `xml:"shared,attr"`
    Tags    string    `xml:"tag,attr"`
    Meta    string    `xml:"meta,attr"`
}

type Tags map[string]uint

var (
    hamid string
)

func update_tags_cache(ga *Alfred.GoAlfred) (err error) {
    posts := new(Posts)

    needed, err := updateNeeded(ga)
    if err != nil {
        return err
    }
    if !needed {
        return err
    }

    if err = updatePostsCache(ga); err != nil {
        return err
    }
    if posts, err = readPostsCache(ga); err != nil {
        return err
    }

    tags_map := make(Tags)

    for _, pin := range posts.Pins {
        tags := strings.Fields(pin.Tags)
        if len(tags) != 0 {
            for _, tag := range tags {
                count := tags_map[tag]
                count++
                tags_map[tag] = count
            }
        }
    }

    tags_cache_fn, err := ga.Get("tags_cache_pn")
    if err != nil {
        return err
    }
    tags_map.store_tags_cache(tags_cache_fn)
    if err != nil {
        return err
    }

    return nil
}

func updatePostsCache(ga *Alfred.GoAlfred) error {
    u, err := makeURLWithAuth(ga, "/v1/posts/all")
    if err != nil {
        return err
    }

    res, err := fetchDataFromHttp(u)
    if err != nil {
        return err
    }

    fn, err := ga.Get("posts_cache_fn")
    if err != nil {
        return err
    }
    file, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0666)
    defer file.Close()
    if err != nil {
        return err
    }
    n, err := file.WriteString(string(res))
    if n != len(res) {
        return errors.New("Wrote different # of bytes than expected.")
    }
    return nil
}

func readPostsCache(ga *Alfred.GoAlfred) (posts *Posts, err error) {
    posts = new(Posts)

    fn, err := ga.Get("posts_cache_fn")
    if err != nil {
        return nil, err
    }
    file, err := os.Open(fn)
    defer file.Close()
    if err != nil {
        return nil, err
    }
    result, err := ioutil.ReadAll(file)
    if err != nil {
        return nil, err
    }
    rd := strings.NewReader(string(result))
    brd := bufio.NewReaderSize(rd, rd.Len())
    p := make([]byte, rd.Len())
    brd.Read(p)
    if err = xml.Unmarshal(p, posts); err != nil {
        return nil, err
    }
    return posts, nil
}

func updateNeeded(ga *Alfred.GoAlfred) (flag bool, err error) {
    u, err := makeURLWithAuth(ga, "/v1/posts/update")
    if err != nil {
        return false, err
    }

    status, err := fetchDataFromHttp(u)
    if err != nil {
        return false, err
    }

    var pinRes pinboardResponse
    if err = xml.Unmarshal(status, &pinRes); err != nil {
        return false, err
    }
    if pinRes.UpdateTime.Update.Before(time.Now()) { // Always update!!!
        return true, nil
    } else {
        return false, nil
    }
    return
}

func fetchDataFromHttp(u url.URL) ([]byte, error) {
    res, err := http.Get(u.String())
    if err != nil {
        return nil, err
    }
    if res.StatusCode != http.StatusOK {
        return nil, errors.New(res.Status)
    }
    result, err := ioutil.ReadAll(res.Body)
    res.Body.Close()
    if err != nil {
        return nil, err
    }
    return result, nil
}

func makeURLWithAuth(ga *Alfred.GoAlfred, pathURL string) (url.URL, error) {
    u := url.URL{}
    u.Scheme = hostURLScheme
    u.Host = path.Join(hostURLPinboard, pathURL)
    q := u.Query()
    auth_token, err := ga.Get("oauth")
    if err != nil {
        return url.URL{}, err
    } else if auth_token == "" {
        return url.URL{}, errors.New("Set your authorization token first!")
    }
    q.Set("auth_token", auth_token)
    u.RawQuery = q.Encode()
    return u, nil
}

func (tm *Tags) store_tags_cache(fn string) error {
    file, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0666)
    defer file.Close()
    if err != nil {
        return err
    }
    var b bytes.Buffer
    enc := gob.NewEncoder(&b)
    err = enc.Encode(tm)
    if err != nil {
        return err
    }
    file.Write(b.Bytes())
    return nil
}

func load_tags_cache(fn string) (Tags, error) {
    file, err := os.Open(fn)
    defer file.Close()
    if err != nil {
        return nil, err
    }

    tags_map := make(Tags)
    decoder := gob.NewDecoder(file)
    err = decoder.Decode(&tags_map)
    if err != nil {
        return nil, err
    }
    return tags_map, nil
}
