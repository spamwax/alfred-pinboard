package main

import (
    "bitbucket.org/listboss/go-alfred"
    "net/http"
    "net/url"
    // "fmt"
    // "io/ioutil"
    "errors"
    "os"
    "os/exec"
    "path"
    "strconv"
    "strings"
    // "regexp"
)

var (
    AccountName     string = "hamid"
    TagsCacheFN     string = "tags_cache"
    MaxNoResults    int    = 10
    hostURLPinboard string = "api.pinboard.in"
    hostURLScheme   string = "https"
)

type pinboardPayload struct {
    url         string
    description string
    extended    string
    tags        string
    replace     string
    shared      string
}

type pinboardResponse struct {
    Result string
}

func main() {
    ga := Alfred.NewAlfred("go-pinboard")
    res, _ := ga.XML()
    _ = res // remove this after debugging
    // fmt.Println(string(res), os.Args)
    args := os.Args[1:]

    // args = []string{"p"}
    if len(args) == 0 {
        return
    }

    showTags := true
    for _, arg := range args {
        if len(arg) > 0 {
            if arg[0] == '#' && (len(arg) == 1 || arg[1] != '#') {
                showTags = false
                break
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
        err := getBrowserInfo(ga)
        err = postToCloud(args, ga)
        if err != nil {
            ga.MakeError(err)
            ga.WriteToAlfred()
            os.Exit(1)
        }
    }
}

// args: list of tags that user has input before entering 'note'
func generateTagSuggestions(args []string, ga *Alfred.GoAlfred) (err error) {
    tags_cache_fn := path.Join(ga.CacheDir,
        strings.Join([]string{TagsCacheFN, AccountName}, "_"))
    noTagQ := len(args)
    last_arg := args[noTagQ-1]
    // TODO: Add setting so that user can toggle showing lower/upper case tags
    strings.ToLower(last_arg)
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
        if strings.Contains(tag, q) {
            m[tag] = count
        }
    }
    return m, nil
}

func postToCloud(args []string, ga *Alfred.GoAlfred) (err error) {
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
    payload.shared = ga.Get("shared") // TODO: set this setting in init()
    urlReq := encodeURL(payload)
    err = postToPinboard(urlReq)
    return err
}

func postToPinBoard(req url.URL) (err error) {
    res, err := http.Get(req.String())
    if err != nil {
        return err
    }
    status, err := ioutil.ReadAll(res.Body)
    res.Body.Close()
    if err != nil {
        return err
    }

}
func getBrowserInfo(ga *Alfred.GoAlfred) (pinInfo []string, err error) {
    browser, err = ga.Get("browser")
    if err != nil {
        return err
    }
    if len(browser) == 0 {
        browser = "safarai"
    }
    appleScript := appleScriptDetectBrowser[browser]
    b, err := exec.Command("osascript", "-s", "s", "-s", "o", "-e",
        appleScript).Output()
    if err != nil {
        return err
    }
    out := string(b)
    foo_ := strings.Split(strings.Trim(out, "{} "), ",")
    pinURL := strings.Trim(foo_[0], "\" ")
    pinDesc := ""
    if len(foo_) > 1 {
        pinDesc = strings.Trim(foo_[1], "\"")
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
