package main

import (
    "os"
    "reflect"
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
    posts := new(Posts)
    if posts, err = readPostsCache(ga); err != nil {
        return err
    }

    var sPins sortedBookmarks
    sPins = bookmarksContain(query, posts)

    ic := 0
    for idx, v := range sPins {
        pin := v.bookmark
        icon := Alfred.NewIcon("pinboard-pin.icns", "")
        ga.AddItem(strconv.Itoa(idx), pin.Desc, pin.Url, "yes", "no", "", pin.Url, icon, false)
        // ga.AddItem(uid, title, subtitle, valid, auto, rtype, arg, icon, check_valid)

        ic++
        if ic > MaxNoResults {
            break
        }
    }
    return
}

var searchFields = []string{"Desc", "Tags", "Notes"}

func bookmarksContain(query []string, posts *Posts) (sb sortedBookmarks) {
    for _, pin := range posts.Pins {
        matchedPin := true
        v := reflect.ValueOf(pin)
        for _, q := range query {
            found_query := false
            for _, field_ := range searchFields {
                f := v.FieldByName(field_)
                fcontent := f.Interface().(string)
                fcontent = strings.ToLower(fcontent)
                found_query = found_query || strings.Contains(fcontent, q)
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
    return sb
}
