package main

import (
    "bufio"
    "bytes"
    "encoding/gob"
    "encoding/xml"
    "errors"
    "io/ioutil"
    "net/http"
    "net/url"
    "os"
    "path"
    "strings"
    "time"

    Alfred "bitbucket.org/listboss/go-alfred"
)

type pinboardUpdateResponse struct {
    XMLName  xml.Name  `xml:"update"`
    Datetime time.Time `xml:"time,attr"`
}

type pinboardResultResponse struct {
    XMLName xml.Name `xml:"result"`
    Code    string   `xml:"code,attr"`
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

func update_tags_cache(ga *Alfred.GoAlfred) (err error) {
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

    posts := new(Posts)
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

    tags_cache_fn, err := ga.Get("tags_cache_fn")
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
    file, err := os.Create(fn)
    defer file.Close()
    if err != nil {
        return err
    }
    n, err := file.WriteString(string(res))
    if n != len(res) {
        return errors.New("Wrote different # of bytes than expected.")
    }
    ga.Set("update_time", time.Now().Format(time.RFC3339Nano))
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

    var pinRes pinboardUpdateResponse
    if err = xml.Unmarshal(status, &pinRes); err != nil {
        return false, err
    }

    last_update, err := ga.Get("update_time")
    if err != nil {
        return false, err
    }

    var lastTime time.Time
    if last_update == "" { // No update has yet been made
        last_update = "0001-01-01T23:00:00Z"
    }
    lastTime, err = time.Parse(time.RFC3339Nano, last_update)

    if err != nil {
        return false, err
    }
    if pinRes.Datetime.After(lastTime) {
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
    file, err := os.Create(fn)
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
