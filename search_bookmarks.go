package main

import (
    "os"
    "reflect"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "time"

    Alfred "bitbucket.org/listboss/go-alfred"
)

type bookmarkPair struct {
    bookmark Link
    date     time.Time
}

type sortedBookmarks []bookmarkPair

func (sb sortedBookmarks) Len() int           { return len(sb) }
func (sb sortedBookmarks) Swap(i, j int)      { sb[i], sb[j] = sb[j], sb[i] }
func (sb sortedBookmarks) Less(i, j int) bool { return sb[i].date.After(sb[j].date) }

func showBookmarks(query []string, ga *Alfred.GoAlfred) {
    err := getBookmarksContaining(query, ga)
    if err != nil {
        ga.MakeError(err)
        ga.WriteToAlfred()
        os.Exit(1)
    }
    ga.WriteToAlfred()
}

func getBookmarksContaining(query []string, ga *Alfred.GoAlfred) (err error) {
    var sPins sortedBookmarks
    sPins, err = bookmarksContain(query, ga)
    if err != nil {
        return err
    }

    ic := 0
    for idx, v := range sPins {
        pin := v.bookmark
        icon := Alfred.NewIcon("pinboard-pin.icns", "")
        ga.AddItem(strconv.Itoa(idx), pin.Desc, pin.Url, "yes", "no", "", pin.Url, icon, false)
        // ga.AddItem(uid, title, subtitle, valid, auto, rtype, arg, icon, check_valid)

        ic++
        if ic == MaxNoResults_Bookmarks {
            break
        }
    }
    return
}

var searchFields = []string{"Desc", "Tags", "Notes", "Url"}

func bookmarksContain(query []string, ga *Alfred.GoAlfred) (sb sortedBookmarks, err error) {

    posts := new(Posts)
    if posts, err = readPostsCache(ga); err != nil {
        return nil, err
    }

    // If fuzzy search is search, compile the corresponding regular expression.
    var fuzzy string
    if fuzzy, err = ga.Get("fuzzy_search"); err != nil {
        return nil, err
    }
    var re *regexp.Regexp

    for _, pin := range posts.Pins {
        matchedPin := true
        v := reflect.ValueOf(pin)
        for _, q := range query {
            found_query := false
            if fuzzy == "yes" {
                re = buildRegExp(q)
            }
            for _, field_ := range searchFields {
                f := v.FieldByName(field_)
                fcontent := f.Interface().(string)
                fcontent = strings.ToLower(fcontent)
                var m bool
                if fuzzy == "yes" {
                    m = re.MatchString(fcontent)
                } else {
                    m = strings.Contains(fcontent, q)
                }
                found_query = found_query || m
                if found_query {
                    break
                }
            }
            if !found_query {
                matchedPin = false
                break
            }
        }
        if matchedPin {
            sb = append(sb, bookmarkPair{bookmark: pin, date: pin.Time})
        }
    }
    sort.Sort(sb)
    return sb, err
}
