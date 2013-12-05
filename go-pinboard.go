package main

import (
    "bitbucket.org/listboss/go-alfred"
    "encoding/xml"
    "errors"
    "github.com/codegangsta/cli"
    // "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "net/url"
    "os"
    "os/exec"
    "path"
    "strconv"
    "strings"
    "time"
    // "regexp"
)

var (
    AccountName      string = "hamid"
    TagsCacheFN      string = "tags_cache"
    MaxNoResults     int    = 10
    hostURLPinboard  string = "api.pinboard.in"
    hostURLScheme    string = "https"
    commentCharacter string = ";"
)

type pinboardPayload struct {
    url         string
    description string
    extended    string
    tags        string
    replace     string
    shared      string
    auth_token  string
}

func main() {
    ga := Alfred.NewAlfred("go-pinboard")
    ga.Set("shared", "no")
    ga.Set("replace", "yes")
    ga.Set("oauth", "username:token")
    ga.Set("browser", "chrome")

    app := cli.NewApp()
    app.Name = "alfred_pinboard"
    app.Usage = "Alfred Workflow helper to manage Pinboard pins using Alfred."
    app.Action = func(ctx *cli.Context) {
        os.Stdout.Write([]byte(app.Usage))
    }
    updateBookmarksCache := cli.Command{
        Name:  "update",
        Usage: "Fetches all the bookmarks and updates the tags cache.",
        Action: func(c *cli.Context) {
            update_tags_cache(ga)
        },
    }
    setOptions := cli.Command{
        Name:  "setopions",
        Usage: "Sets token and browser otions",
        Flags: []cli.Flag{
            cli.StringFlag{"browser", "safari", "Browser to fetch the webpage from"},
            cli.StringFlag{"auth", "", "Set authorization token in form of username:token"},
        },
        Action: func(c *cli.Context) {
            if b := c.String("browser"); b != "" {
                ga.Set("browser", b)
            }
            if t := c.String("auth"); t != "" {
                ga.Set("oauth", t)
            }
        },
    }
    postBookmark := cli.Command{
        Name: "post",
        Action: func(c *cli.Context) {
            query := strings.Join(c.Args(), " ")
            err := postToCloud(query, ga)
            if err != nil {
                os.Stdout.WriteString(err.Error())
            }
        },
    }
    showTagsCommand := cli.Command{
        Name: "showtags",
    }
    app.Commands = []cli.Command{updateBookmarksCache, setOptions,
        postBookmark, showTagsCommand}

    _fn := ga.DataDir
    logfile, _ := os.OpenFile(_fn+"/log.txt", os.O_APPEND|os.O_WRONLY, 0666)
    var L = log.New(logfile, time.Now().String()+": ", 0)
    // fmt.Println(len(os.Args))
    args := os.Args[1:]

    // args := []string{"hami:d", commentCharacter, "testing"}
    if len(args) == 0 {
        // TODO: show the bookmark if it has already bin pinned
        return
    }
    L.Println(args)

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
        // b, _ := ga.XML()
        // L.Println(string(b))
    } else {
        ga.AddItem("", "Hit Enter to save the bookmark.", "Pinboard", "yes",
            "", "", "{query}", Alfred.NewIcon("bookmark.icns", ""), true)
        ga.WriteToAlfred()
        // err := postToCloud(query, ga)
        // if err != nil {
        //     os.Stdout.WriteString(err.Error())
        // }
    }
}

// args: list of tags that user has input before entering 'note'
func generateTagSuggestions(args []string, ga *Alfred.GoAlfred) (err error) {
    tags_cache_fn := path.Join(ga.CacheDir,
        strings.Join([]string{TagsCacheFN, AccountName}, "_"))
    noTagQ := len(args)
    last_arg := args[noTagQ-1]
    // TODO: Add setting so that user can toggle showing lower/upper case tags
    last_arg = strings.ToLower(last_arg)
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
            "{query}", Alfred.NewIcon("tag_icon.icns", ""), true)
        // ga.AddItem(uid, title, subtitle, valid, auto, rtype, arg, icon, check_valid)
        ic++
        if ic > MaxNoResults {
            break
        }
    }
    return nil
}

func getTagsFor(q string, tags_cache_fn string) (m map[string]uint, err error) {
    // TODO: Use Pinboard's api to get list of popular tags and add to this
    tags_map, err := load_tags_cache(tags_cache_fn)
    if err != nil {
        return nil, err
    }

    m = make(map[string]uint)
    for tag, count := range tags_map {
        if count == 0 {
            continue
        }
        if strings.Contains(strings.ToLower(tag), q) {
            m[tag] = count
        }
    }
    return m, nil
}

func postToCloud(args string, ga *Alfred.GoAlfred) (err error) {
    pinInfo, err := getBrowserInfo(ga)
    if err != nil {
        return err
    }
    oauth, err := ga.Get("oauth")
    if err != nil {
        return err
    }
    if oauth == "" {
        return errors.New("Set your authorization token first!")
    }
    var payload pinboardPayload
    payload.tags, payload.extended = parseTags(args)
    payload.url = pinInfo[0]
    payload.description = pinInfo[1]
    payload.replace = "yes"

    if payload.shared, err = ga.Get("shared"); err != nil {
        payload.shared = "no"
    }
    payload.auth_token = oauth

    urlReq := encodeURL(payload, "/v1/posts/add")
    // fmt.Println(urlReq.String())
    err = postToPinboard(urlReq)
    return err
}

func postToPinboard(req url.URL) (err error) {
    // fmt.Printf("req: %v\n", req.String())
    res, err := http.Get(req.String())
    // fmt.Printf("res: %v\n", res.StatusCode)
    if err != nil {
        return err
    }
    if res.StatusCode != 200 {
        return errors.New(res.Status)
    }
    status, err := ioutil.ReadAll(res.Body)
    res.Body.Close()
    if err != nil {
        return err
    }
    var pinRes pinboardResponse
    if err = xml.Unmarshal(status, &pinRes); err != nil {
        return err
    }
    if pinRes.Result.Code != "done" {
        return errors.New(pinRes.Result.Code)
    }
    return nil

}

func parseTags(args string) (tags, desc string) {
    foo_ := strings.Split(args, commentCharacter)
    tags = strings.Trim(foo_[0], " ")
    desc = ""
    if len(foo_) > 1 {
        desc = strings.Trim(foo_[1], " ")
    }
    // fmt.Printf("tags: %v\ndesc: %v\n", tags, desc)
    return tags, desc
}

func getBrowserInfo(ga *Alfred.GoAlfred) (pinInfo []string, err error) {
    browser, err := ga.Get("browser")
    if err != nil {
        return nil, err
    }
    if len(browser) == 0 {
        browser = "safarai"
    }
    appleScript := appleScriptDetectBrowser[browser]
    b, err := exec.Command("osascript", "-s", "s", "-s", "o", "-e",
        appleScript).Output()
    if err != nil {
        return nil, err
    }
    out := string(b)
    // fmt.Println(out)
    foo0 := strings.Trim(out, "{}\n")
    foo1 := strings.Split(foo0, ",")
    // fmt.Printf("foo0: '%v'\np1: '%v'\np2: '%v'\n", foo0, foo1[0], foo1[1])
    pinURL := strings.Trim(foo1[0], "\" ")
    pinDesc := ""
    if len(foo1) > 1 {
        pinDesc = strings.Trim(foo1[1], "\" ")
    }
    // fmt.Printf("%v\n%v\n", pinURL, pinDesc)
    return []string{pinURL, pinDesc}, err
}

func encodeURL(payload pinboardPayload, pathURL string) (req url.URL) {
    u := url.URL{}
    u.Scheme = hostURLScheme
    u.Host = path.Join(hostURLPinboard, pathURL)
    q := u.Query()
    q.Set("url", payload.url)
    q.Set("description", payload.description)
    q.Set("extended", payload.extended)
    q.Set("replace", payload.replace)
    q.Set("shared", payload.shared)
    q.Set("tags", payload.tags)
    q.Set("auth_token", payload.auth_token)
    u.RawQuery = q.Encode()
    return u
}

var appleScriptDetectBrowser = map[string]string{
    "safari": `on run
        tell application "Safari"
            set theURL to URL of current tab of window 1
            set theDesc to name of current tab of window 1
        end tell
        return {theUrl, theDesc}
        end run`,
    "chrome": `on run
            tell application "Google Chrome"
                set theURL to URL of active tab of first window
                set theDesc to title of active tab of first window
            end tell
            return {theURL, theDesc}
            end run`,
}
