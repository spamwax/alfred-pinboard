package main

import (
    "bitbucket.org/listboss/go-alfred"
    "fmt"
    "os"
    "strings"
    // "regexp"
)

func main() {
    ga := Alfred.NewAlfred("go-pinboard")
    res, _ := ga.XML()
    fmt.Println(string(res), os.Args)
    args := os.Args[1:]

    showTags := true
    for _, arg := range args {
        if len(arg) > 0 {
            if arg[0] == '#' && (len(arg) == 1 || arg[1] != '#') {
                showTags == false
                break
            }
        }
    }
    if showTags {
        err := generateTags(args, ga)
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

func generateTags(args []string, ga *Alfred.GoAlfred) (err error) {
    return nil
}

func postToCloud(args []string, ga *Alfred.GoAlfred) (err error) {
    return nil
}
