package main

import (
    Alfred "bitbucket.org/listboss/go-alfred"
)

type BookmarkAPI interface {
    postToCloud(args string, ga *Alfred.GoAlfred) (info []string, err error)
    // update_tags_cache(ga *Alfred.GoAlfred) (err error)
    // updatePostsCache(ga *Alfred.GoAlfred) (err error)
    // readPostsCache(ga *Alfred.GoAlfred) (posts *Posts, err error)
    showBookmarks(query []string, ga *Alfred.GoAlfred)
    showtags(args []string, ga *Alfred.GoAlfred)
}
