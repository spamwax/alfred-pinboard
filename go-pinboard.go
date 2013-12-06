package main

import (
    "bitbucket.org/listboss/go-alfred"
    "encoding/xml"
    "errors"
    "github.com/codegangsta/cli"
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
)

var (
    AccountName      string = "hamid"
    TagsCacheFN      string = "tags_cache"
    PostsCachFn      string = "posts_cache"
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

func Init() (ga *Alfred.GoAlfred) {
    var err error
    ga = Alfred.NewAlfred("go-pinboard")
    ga.Set("shared", "no")
    ga.Set("replace", "yes")
    ga.Set("browser", "chrome")
    AccountName, err = ga.Get("username")
    if err != nil {
        os.Stdout.Write([]byte("Can't get username!"))
    }
    if AccountName == "" {
        AccountName = "deleteMe"
    }
    tags_cache_fn := path.Join(ga.CacheDir,
        strings.Join([]string{TagsCacheFN, AccountName}, "_"))
    posts_cache_fn := path.Join(ga.CacheDir,
        strings.Join([]string{PostsCachFn, AccountName}, "_"))
    ga.Set("tags_cache_fn", tags_cache_fn)
    ga.Set("posts_cache_fn", posts_cache_fn)
    return ga
}

func main() {
    ga := Init()
    app := cli.NewApp()
    app.Name = "alfred_pinboard"
    app.Usage = "Alfred Workflow helper to manage Pinboard pins using Alfred"
    app.Action = func(ctx *cli.Context) {
        foo := `NAME:
   alfred_pinboard - Alfred Workflow helper to manage Pinboard pins using Alfred.

   enter "alfred_pinboard help" for more information`
        os.Stdout.Write([]byte(foo))
    }
    updateBookmarksCache := cli.Command{
        Name:  "update",
        Usage: "Fetches all the bookmarks and updates the tags cache.",
        Action: func(c *cli.Context) {
            err := update_tags_cache(ga)
            if err != nil {
                os.Stdout.WriteString(err.Error())
            }
        },
    }
    setOptions := cli.Command{
        Name:        "setoptions",
        Usage:       "Sets token and browser options",
        Description: "set Workflow related options.",
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
                _username := strings.Split(t, ":")[0]
                ga.Set("username", _username)
                tags_cache_fn := path.Join(ga.CacheDir,
                    strings.Join([]string{TagsCacheFN, _username}, "_"))
                posts_cache_fn := path.Join(ga.CacheDir,
                    strings.Join([]string{PostsCachFn, _username}, "_"))
                ga.Set("tags_cache_fn", tags_cache_fn)
                ga.Set("posts_cache_fn", posts_cache_fn)
            }
        },
    }
    postBookmark := cli.Command{
        Name:        "post",
        Usage:       "post tag1 tag2 ; extra notes",
        Description: "Posts a bookmark to cloud for the current page of the browser.",
        Action: func(c *cli.Context) {
            query := strings.Join(c.Args(), " ")
            binfo, err := postToCloud(query, ga)
            if err == nil {
                os.Stdout.WriteString(binfo[1])
            } else {
                os.Stdout.WriteString(err.Error())
            }
        },
    }
    showTagsCommand := cli.Command{
        Name:  "showtags",
        Usage: "Shows list of available tags based on intut arguments.",
        Action: func(c *cli.Context) {
            args := []string(c.Args())
            showtags(args, ga)
        },
    }
    app.Commands = []cli.Command{updateBookmarksCache, setOptions,
        postBookmark, showTagsCommand}
    app.Run(os.Args)
}

func showtags(args []string, ga *Alfred.GoAlfred) {
    _fn := ga.DataDir
    logfile, _ := os.OpenFile(_fn+"/log.txt", os.O_APPEND|os.O_WRONLY, 0666)
    var L = log.New(logfile, time.Now().String()+": ", 0)

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
    } else {
        // ga.AddItem(uid, title, subtitle, valid, auto, rtype, arg, icon, check_valid)
        ga.AddItem("chichichi", "Hit Enter to save the bookmark.", query, "yes",
            "", "", query, Alfred.NewIcon("bookmark.icns", ""), false)
        ga.WriteToAlfred()
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
            "", Alfred.NewIcon("tag_icon.icns", ""), true)
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

func postToCloud(args string, ga *Alfred.GoAlfred) (info []string, err error) {
    pinInfo, err := getBrowserInfo(ga)
    if err != nil {
        return pinInfo, err
    }
    oauth, err := ga.Get("oauth")
    if err != nil {
        return pinInfo, err
    }
    if oauth == "" {
        return pinInfo, errors.New("Set your authorization token first!")
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
    err = postToPinboard(urlReq)

    return pinInfo, err
}

func postToPinboard(req url.URL) (err error) {
    res, err := http.Get(req.String())
    if err != nil {
        return err
    }
    if res.StatusCode != http.StatusOK {
        return errors.New(res.Status)
    }
    status, err := ioutil.ReadAll(res.Body)
    res.Body.Close()
    if err != nil {
        return err
    }
    var pinRes pinboardResultResponse
    if err = xml.Unmarshal(status, &pinRes); err != nil {
        return err
    }
    if pinRes.Code != "done" {
        return errors.New(pinRes.Code)
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
    foo0 := strings.Trim(out, "{}\n")
    foo1 := strings.Split(foo0, ",")
    pinURL := strings.Trim(foo1[0], "\" ")
    pinDesc := ""
    if len(foo1) > 1 {
        pinDesc = strings.Trim(foo1[1], "\" ")
    }
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
