package main

import (
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	Alfred "bitbucket.org/listboss/go-alfred"
)

type tagpair struct {
	name  string
	count uint
}

type sortedTags []tagpair

func (b sortedTags) Len() int           { return len(b) }
func (b sortedTags) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b sortedTags) Less(i, j int) bool { return b[i].count > b[j].count }

func showtags(args []string, ga *Alfred.GoAlfred) {
	// _fn := ga.DataDir
	// logfile, _ := os.OpenFile(_fn+"/log.txt", os.O_APPEND|os.O_WRONLY, 0666)
	// var L = log.New(logfile, time.Now().String()+": ", 0)

	if len(args) == 0 {
		// TODO: show the bookmark if it has already bin pinned
		return
	}

	// Show tags autocomplete?
	query := strings.Join(args, " ")
	showTags := true
	if strings.Contains(query, commentCharacter) {
		for _, arg := range args {
			if len(arg) > 0 {
				if strings.Contains(arg, commentCharacter) {
					if string(arg[0]) == commentCharacter &&
						(len(arg) == 1 || string(arg[1]) != commentCharacter) {
						showTags = false
						break
					}
				}
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
		// ga.AddItem(uid, title, subtitle, valid, auto, rtype, arg, icon, check_valid)
		ga.AddItem("", "Hit Enter to save the bookmark.", query, "yes",
			"", "", query, Alfred.NewIcon("bookmark.icns", ""), false)
		ga.WriteToAlfred()
	}
}

// args: list of tags that user has input before entering 'note'
func generateTagSuggestions(args []string, ga *Alfred.GoAlfred) (err error) {
	noTagQ := len(args)
	last_arg := args[noTagQ-1]
	// TODO: Add setting so that user can toggle showing lower/upper case tags
	last_arg = strings.ToLower(last_arg)
	tags, err := getTagsFor(last_arg, ga)
	if err != nil {
		return err
	}

	ic := 0
	for _, tp := range tags {
		tag := tp.name
		freq := tp.count

		uid := strconv.Itoa(int(freq))
		subtitle := "freq: " + uid
		auto_complete := tag
		// Add the current selected tags to the auto_complete
		if noTagQ >= 2 {
			auto_complete = strings.Join(args[:noTagQ-1], " ")
			auto_complete += " " + tag
		}
		if freq == 0 {
			subtitle = "NEW TAG"
		}
		ga.AddItem("", tag, subtitle, "yes", auto_complete+" ", "", auto_complete,
			Alfred.NewIcon("tag_icon.icns", ""), false)
		// ga.AddItem(uid, title, subtitle, valid, auto, rtype, arg, icon, check_valid)
		ic++
		if ic == MaxNoResults_Tags {
			break
		}
	}
	return nil
}

func getTagsFor(q string, ga *Alfred.GoAlfred) (m sortedTags, err error) {
	// TODO: Use Pinboard's api to get list of popular tags and add to this
	tags_cache_fn := path.Join(ga.CacheDir,
		strings.Join([]string{TagsCacheFN, AccountName}, "_"))
	tags_map, err := load_tags_cache(tags_cache_fn)
	if err != nil {
		return nil, err
	}

	// If fuzzy search is search, compile the corresponding regular expression.
	var fuzzy string
	var re *regexp.Regexp
	if fuzzy, err = ga.Get("fuzzy_search"); err != nil {
		return nil, err
	}
	regexp_exp := ""
	if fuzzy == "yes" {
		letters := strings.Split(q, "")
		for _, v := range letters {
			regexp_exp += v + "+.*"
		}
		re = regexp.MustCompile(regexp_exp)
	}

	exact_match := false
	// Iterate over all tags and search for the input query
	for tag, count := range tags_map {
		if count == 0 {
			continue
		}
		if fuzzy == "yes" {
			if re.MatchString(strings.ToLower(tag)) {
				m = append(m, tagpair{tag, count})
			}
		} else {
			if strings.Contains(strings.ToLower(tag), q) {
				// tagsMap[tag] := count
				m = append(m, tagpair{tag, count})
			}
		}
		if tag == q {
			exact_match = true
		}
	}

	// Sort based on descending order of tag freq
	sort.Sort(m)
	if !exact_match {
		m = append(m, tagpair{name: q, count: 0})
	}
	return m, nil
}

func parseTags(args string) (tags, desc string) {
	foo_ := strings.Split(args, commentCharacter)
	tags = strings.Trim(foo_[0], " ")
	desc = ""
	if len(foo_) > 1 {
		desc = strings.Trim(foo_[1], " ")
	}
	return tags, desc
}
