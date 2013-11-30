package main

import (
    "bitbucket.org/listboss/go-alfred"
    "fmt"
    "os"
    "testing"
)

func TestArgs(t *testing.T) {
    ga := Alfred.NewAlfred("go-pinboard")
    args := []string{"x"}
    err := generateTagSuggestions(args, ga)
    if err != nil {
        ga.MakeError(err)
        ga.WriteToAlfred()
        os.Exit(1)
    }
    // ga.WriteToAlfred()
    res, _ := ga.XML()
    fmt.Println(string(res))
}
