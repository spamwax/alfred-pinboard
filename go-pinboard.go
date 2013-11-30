package main

import (
    "bitbucket.org/listboss/go-alfred"
    "bufio"
    "fmt"
    "io"
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
    // fmt.Println(string(res), os.Args)
    args := os.Args[1:]

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
    tags_cache := path.Join(ga.CacheDir,
        strings.Join([]string{TagsCacheFN, AccountName}, "_"))
    fileReader, err := os.Open(tags_cache)
    defer fileReader.Close()
    if err != nil {
        return err
    }

    tags, err := getTagsFor(args[len(args)-1], fileReader)
    if err != nil {
        return err
    }
    for tag, freq := range tags {
        ga.AddItem("", tag, strconv.Itoa(freq), "no", "yes", "", "",
            Alfred.NewIcon("tag_icon.png", ""), true)
    }
    return nil
}

func getTagsFor(q string, inFile io.Reader) (m map[string]int, err error) {
    m = make(map[string]int)
    scanner := bufio.NewScanner(inFile)
    ic := 0
    eof := false
    for !eof {
        valid := scanner.Scan()
        if !valid {
            break
        }
        err = scanner.Err()
        line := scanner.Text()
        fields := strings.Fields(line)
        if len(fields) == 1 {
            fields = append(fields, "1")
        }
        if len(fields) == 2 {
            if strings.Contains(fields[0], q) {
                foo, _ := strconv.ParseInt(fields[1], 10, 32)
                m[fields[0]] = int(foo)
                ic++
            }
        }
        if ic > MaxNoResults || err == io.EOF || err == io.ErrNoProgress {
            err = nil
            eof = true
        } else if err != nil {
            return nil, err
        }
    }
    return
}

func postToCloud(args []string, ga *Alfred.GoAlfred) (err error) {
    return nil
}
