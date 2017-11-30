package main

import (
	"bitbucket.org/listboss/go-alfred"
	"fmt"
	"os"
	"testing"
)

func TestArgs(t *testing.T) {
	ga := Alfred.NewAlfred("go-pinboard")
	args := []string{"p"}
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

func TestUpdateCache(t *testing.T) {
	ga := Alfred.NewAlfred("go-pinboard")
	err := update_tags_cache(ga)
	if err != nil {
		t.Error("Couldn't update tags cache")
	}

	Pins := new(Posts)
	if Pins, err = readPostsCache(ga); err != nil {
		t.Error("Couldn't read posts cache")
	}

	for _, p := range Pins.Pins {
		fmt.Printf("url: %v\nhash: %v\nshared: %v\ntags: %v\n", p.Url, p.Hash, p.Shared, p.Tags)
		fmt.Printf("ext: %v\ntime: %v\nmeta:%v\n", p.Notes, p.Time, p.Meta)
		fmt.Println()
		//fmt.Println(strings.Fields(p.Tags))
	}
}
