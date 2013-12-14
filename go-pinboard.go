package main

import (
    "net/url"
    "os"
    "os/exec"
    "path"
    "strings"

    Alfred "bitbucket.org/listboss/go-alfred"
    cli "github.com/codegangsta/cli"
)

var (
    AccountName      string = "accountName"
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
    // ga.Set("browser", "chrome")
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
    app.Usage = "Alfred Workflow to manage Pinboard pins using Alfred"
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

func getBrowserInfo(ga *Alfred.GoAlfred) (pinInfo []string, err error) {
    browser, err := ga.Get("browser")
    if err != nil {
        return nil, err
    }
    if len(browser) == 0 {
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
}
