package main

import (
    "bitbucket.org/listboss/go-alfred"
    "bufio"
    "encoding/xml"
    "fmt"
    "os"
    "path"
    "strings"
    "time"
)

var data = `<?xml version="1.0" encoding="UTF-8" ?>
<posts user="hamid">
<post href="http://vimeo.com/37562944#" time="2013-03-15T07:07:48Z" description="See Through 3D Desktop, on Vimeo" extended="" tag="3d ui screen monitor desktop leapmotion" hash="62e78977500a21754caa741fcafa3695" meta="ddf8de5b3a6d2cd6d7562c2ae27902d4"  shared="yes"  />
<post href="http://www.itefix.no/phpws/index.php?module=phpwsbb&amp;PHPWSBB_MAN_OP=view&amp;PHPWS_MAN_ITEMS[]=343" time="2013-03-09T11:38:25Z" description="Icon Sets and Designers Directory - Iconify.info - A comprehensive curated gallery of icons and icon-related resources" extended="" tag="icon font free vector-graphic svg" hash="5b3f7ff0257a582531c8f33e703f76a5" meta="b38c248e539cf132ce8c09207931f33a"  shared="no"  />
<post href="http://durak.org/cvswebsites/howto-cvs/node37.html" time="2006-03-03T20:55:18Z" description="CVS Server Setup" extended="CVS Server Setup" tag="Learn Linux CVS" hash="f99a793264f1a3ef8e8ecdc7d1f7c154" meta="7da1d359a76955fcbd2cbce622483f63"  shared="no"  />
</posts>`

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

var (
    hamid string
)

func update_tags_cache(ga *Alfred.GoAlfred) (posts *Posts, err error) {
    posts = new(Posts)
    rd := strings.NewReader(data)
    brd := bufio.NewReaderSize(rd, rd.Len())
    p := make([]byte, rd.Len())
    fmt.Println(rd.Len())
    brd.Read(p)
    xml.Unmarshal(p, posts)

    tags_map := make(map[string]uint)

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

    tags_cache_fn := path.Join(ga.CacheDir,
        strings.Join([]string{TagsCacheFN, AccountName}, "_"))
    file, err := os.OpenFile(tags_cache_fn, os.O_CREATE|os.O_WRONLY, 0666)
    defer file.Close()
    if err != nil {
        return nil, err
    }
    for tag, count := range tags_map {
        file.WriteString(fmt.Sprintf("%s %d\n", tag, count))
    }
    return posts, nil
}
