package main

import (
    "encoding/xml"
    "errors"
    "io/ioutil"
    "net/http"
    "net/url"

    Alfred "bitbucket.org/listboss/go-alfred"
)

var hostURLPinboard string = "api.pinboard.in"
var hostSchemePinboard string = "https"

type PinboardAPI struct {
    AccountName     string
    hostURL         string
    hostURLScheme   string
    postURL         string
    updateURL       string
    allBookmarksURL string
    // payload
    // auth_token
}

func (parpi PinboardAPI) PostToCloud(args string, ga *Alfred.GoAlfred) (info []string, err error) {
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

func (parpi PinboardAPI) postToPinboard(req url.URL) (err error) {
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
