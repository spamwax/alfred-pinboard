package main

import (
    "bitbucket.org/listboss/go-alfred"
    // "fmt"
    // "io/ioutil"
    "os"
    "path"
    "strconv"
    "strings"
    // "regexp"
)

var (
    AccountName  string = "hamid"
    TagsCacheFN  string = "tags_cache"
    MaxNoResults int    = 10
)

func main() {
    ga := Alfred.NewAlfred("go-pinboard")
    res, _ := ga.XML()
    _ = res // remove this after debugging
    // fmt.Println(string(res), os.Args)
    args := os.Args[1:]

    // args = []string{"p"}
    if len(args) == 0 {
        return
    }

    showTags := true
    for _, arg := range args {
        if len(arg) > 0 {
            if arg[0] == '#' && (len(arg) == 1 || arg[1] != '#') {
                showTags = false
                break
            }
        }
    }
    if showTags {
        err := generateTagSuggestions(args, ga)
        if err != nil {
            ga.MakeError(err)
            ga.WriteToAlfred()
            os.Exit(1)
        }
        ga.WriteToAlfred()
    } else {
        err := postToCloud(args, ga)
        if err != nil {
            ga.MakeError(err)
            ga.WriteToAlfred()
            os.Exit(1)
        }
    }
}

func generateTagSuggestions(args []string, ga *Alfred.GoAlfred) (err error) {
    tags_cache_fn := path.Join(ga.CacheDir,
        strings.Join([]string{TagsCacheFN, AccountName}, "_"))
    noTagQ := len(args)
    last_arg := args[noTagQ-1]
    // TODO: Add setting so that user can toggle showing lower/upper case tags
    strings.ToLower(last_arg)
    tags, err := getTagsFor(last_arg, tags_cache_fn)
    if err != nil {
        return err
    }

    ic := 0
    for tag, freq := range tags {
        auto_complete := tag
        // Add the current selected tags to the auto_complete
        if noTagQ >= 2 {
            auto_complete = strings.Join(args[:noTagQ-1], " ")
            auto_complete += " " + tag
        }
        // TODO: generate UUID for the tags so Alfred can learn about them.
        ga.AddItem("", tag, strconv.Itoa(int(freq)), "no", auto_complete, "",
            "", Alfred.NewIcon("tag_icon.icns", ""), true)
        ic++
        if ic > MaxNoResults {
            break
        }
    }
    return nil
}

func getTagsFor(q string, tags_cache_fn string) (m map[string]uint, err error) {
    tags_map, err := load_tags_cache(tags_cache_fn)
    if err != nil {
        return nil, err
    }

    m = make(map[string]uint)
    for tag, count := range tags_map {
        if count == 0 {
            continue
        }
        if strings.Contains(tag, q) {
            m[tag] = count
        }
    }
    return m, nil
}

func postToCloud(args []string, ga *Alfred.GoAlfred) (err error) {
    return nil
}
