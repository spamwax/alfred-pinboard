package main

import (
    "net/url"
    "os"
    "os/exec"
    "path"
    "regexp"
    "strconv"
    "strings"

    Alfred "bitbucket.org/listboss/go-alfred"
    cli "github.com/codegangsta/cli"
)

var (
    AccountName            string = "accountName"
    TagsCacheFN            string = "tags_cache"
    PostsCachFn            string = "posts_cache"
    MaxNoResults_Tags      int    = 10
    MaxNoResults_Bookmarks int    = 10
    hostURLPinboard        string = "api.pinboard.in"
    hostURLScheme          string = "https"
    commentCharacter       string = ";"
    AutoUpdate             string = "no"
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
    // ga.Set("shared", "no")
    ga.Set("replace", "yes")
    AccountName, err = ga.Get("username")
    if err != nil {
        ga.MakeError(err)
        ga.WriteToAlfred()
        os.Exit(1)
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
    v, _ := ga.Get("max_tags")
    if v == "" {
        v = "10"
    }
    tmp, err := strconv.ParseInt(v, 10, 32)
    if err != nil {
        MaxNoResults_Tags = 10
    } else {
        MaxNoResults_Tags = int(tmp)
    }

    v, _ = ga.Get("max_bookmarks")
    if v == "" {
        v = "10"
    }
    tmp, err = strconv.ParseInt(v, 10, 32)
    if err != nil {
        MaxNoResults_Bookmarks = 10
    } else {
        MaxNoResults_Bookmarks = int(tmp)
    }

    v, err = ga.Get("auto_update")
    if err == nil && v != "" {
        AutoUpdate = v
    }

    return ga
}

func main() {
    ga := Init()
    app := cli.NewApp()
    app.Name = "alfred_pinboard"
    app.Usage = "Alfred Workflow to manage Pinboard pins."
    app.Action = func(ctx *cli.Context) {
        foo := `NAME:
   alfred_pinboard - Alfred Workflow helper to manage Pinboard pins.

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
            } else {
                os.Stdout.WriteString("Successfully Updated Local Cache.")
            }
        },
    }
    setOptions := cli.Command{
        Name:        "setoptions",
        Usage:       "Sets token and browser options",
        Description: "set Workflow related options.",
        Flags: []cli.Flag{
            cli.StringFlag{Name: "browser", Value: "chrome", Usage: "Browser to fetch the webpage from"},
            cli.StringFlag{Name: "auth", Value: "", Usage: "Set authorization token in form of username:token"},
            cli.StringFlag{Name: "fuzzy,f", Value: "", Usage: "Enable fuzzy search"},
            cli.StringFlag{Name: "shared", Value: "", Usage: "Set sharing/private status for posted bookmarks"},
            cli.StringFlag{Name: "tag-only-search", Value: "", Usage: "Only search through tags when looking up bookmarks"},
            cli.StringFlag{Name: "auto-update", Value: "no", Usage: "Automatically update bookmarks cache after posting a bookmark."},
            cli.IntFlag{Name: "max-tags", Value: -1, Usage: "Set max. number of tags to show."},
            cli.IntFlag{Name: "max-bookmarks", Value: -1, Usage: "Set max. number of bookmarks to show."},
        },
        Action: func(c *cli.Context) {
            // Set max number of tags/bookmarks to show
            if mt := c.Int("max-tags"); mt != -1 {
                mtags := strconv.Itoa(mt)
                ga.Set("max_tags", mtags)
                os.Stdout.WriteString("Max no. tags to show: " + mtags)
            }
            if mb := c.Int("max-bookmarks"); mb != -1 {
                mbook := strconv.Itoa(mb)
                ga.Set("max_bookmarks", mbook)
                os.Stdout.WriteString("Max no. bookmarks to show: " + mbook)
            }
            // Set browser
            if b := c.String("browser"); b != "" {
                ga.Set("browser", b)
            }
            // Set sharing/private status for bookmarks
            if s := c.String("shared"); s != "" {
                ga.Set("shared", s)
                os.Stdout.WriteString("Sharing bookmarks: " + s)
            }
            // Set search option for using tags only
            if ts := c.String("tag-only-search"); ts != "" {
                ga.Set("tag_only_search", ts)
                os.Stdout.WriteString("Tag-only search: " + ts)
            }
            // Set auto-update option
            if au := c.String("auto-update"); au != "" {
                ga.Set("auto_update", au)
            }
            // Set authorization tokens
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
                if err := update_tags_cache(ga); err != nil {
                    os.Stdout.WriteString("Err Setting Auth. " + err.Error())
                } else {
                    os.Stdout.WriteString("Successfully Set Auth. Token.")
                }
            }
            // Enable/Disable fuzzy search
            if value := c.String("fuzzy"); value != "" {
                if value == "yes" || value == "no" {
                    if err := ga.Set("fuzzy_search", value); err != nil {
                        os.Stdout.WriteString("Err setting fuzzy. " + err.Error())
                    } else {
                        os.Stdout.WriteString("Successfully changed\nfuzzy search setting.")
                    }
                }
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

            au := strings.ToLower(AutoUpdate)
            if au == "yes" || au == "1" || au == "on" {
                err := update_tags_cache(ga)
                if err != nil {
                    os.Stdout.WriteString(err.Error())
                }
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
    showBookmarksCommand := cli.Command{
        Name:  "showbookmarks",
        Usage: "Show list of bookmarks that contain 'all' the search keywords",
        Action: func(c *cli.Context) {
            args := []string(c.Args())
            showBookmarks(args, ga)
        },
    }
    showSettingsCommand := cli.Command{
        Name:  "showsettings",
        Usage: "Show Workflow's settings",
        Action: func(c *cli.Context) {
            showSettings(ga)
        },
    }
    app.Commands = []cli.Command{updateBookmarksCache, setOptions,
        postBookmark, showTagsCommand, showBookmarksCommand,
        showSettingsCommand}
    app.Run(os.Args)
}

func showSettings(ga *Alfred.GoAlfred) {
    browser, _ := ga.Get("browser")
    max_tags, _ := ga.Get("max_tags")
    max_bookmarks, _ := ga.Get("max_bookmarks")
    fuzzy_search, _ := ga.Get("fuzzy_search")
    tag_search, _ := ga.Get("tag_only_search")
    shared, _ := ga.Get("shared")

    // ga.AddItem(uid, title, subtitle, valid, auto, rtype, arg, icon, check_valid)
    ga.AddItem("", "Browser: "+browser, "Browser to use.", "yes", "", "",
        "pset browser",
        Alfred.NewIcon("ACD33B7C-7C31-47F4-B8AC-E15E09EC31DD.png", ""), false)

    ga.AddItem("", "No. Tags: "+max_tags, "No. of tags to show.", "yes", "",
        "", "pset tags", Alfred.NewIcon("tag_icon.icns", ""), false)

    ga.AddItem("", "No. Bookmarks: "+max_bookmarks, "No. of bookmarks to show.",
        "yes", "", "", "pset bmarks",
        Alfred.NewIcon("pin.png", ""),
        false)

    ga.AddItem("", "Fuzzy search: "+fuzzy_search, "Use fuzzy search.", "yes",
        "", "", "pset fuzzy", Alfred.NewIcon("fuzzy_search.icns", ""), false)

    ga.AddItem("", "Sharing boookmarks: "+shared,
        "Private or Shared bookmarking", "yes", "", "", "pset shared",
        Alfred.NewIcon("shared_bookmarking.png", ""), false)

    ga.AddItem("", "Auto update bookmarks cache: "+AutoUpdate,
        "Download all bookmarks after posting a new bookmark.",
        "yes", "", "", "pset auto", Alfred.NewIcon("auto_update.png", ""),
        false)

    ga.AddItem("", "Tag-only search: "+tag_search, "Search only the tags?",
        "yes", "", "", "pset tagonly",
        Alfred.NewIcon("tag_only.png", ""), false)
    ga.WriteToAlfred()
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

func buildRegExp(s string) (re *regexp.Regexp) {
    regexp_exp := ""
    for _, v := range s {
        regexp_exp += string(v) + "+.*"
    }
    re = regexp.MustCompile(regexp_exp)
    return
}

func getBrowserInfo(ga *Alfred.GoAlfred) (pinInfo []string, err error) {
    browser, err := ga.Get("browser")
    if err != nil {
        return nil, err
    }
    browser = strings.ToLower(browser)
    if len(browser) == 0 ||
        (browser != "chrome" && browser != "safari" && browser != "chromium" &&
            browser != "firefox") {
        browser = "chrome"
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
    // If the current page doesn't have title set it to the URL
    if pinDesc == "" {
        pinDesc = pinURL
    }

    return []string{pinURL, pinDesc}, err
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
    "chromium": `on run
            tell application "Chromium"
                set theURL to URL of active tab of first window
                set theDesc to title of active tab of first window
            end tell
            return {theURL, theDesc}
            end run`,
    "firefox": `on run
            tell application "Firefox"
                activate
                set w to item 1 of window 1
                set theDesc to name of w
            end tell
            tell application "System Events"
                set myApp to name of first application process whose frontmost is true
                if myApp is "Firefox" then
                    tell application "System Events"
                        keystroke "l" using command down
                        delay 0.5
                        keystroke "c" using command down
                    end tell
                    delay 0.5
                end if
                delay 0.5
            end tell
            set theURL to get the clipboard
            return {theURL, theDesc}
            end run`,
}

// for firefox we could try to only raise the window using:
// tell application "System Events"
//     perform action "AXRaise" of window 1 of process "Firefox"
// end tell
