# Alfred Workflow for Pinboard

There are probably similar [Alfred](http://www.alfredapp.com/) Workflows out there but my main motivation for writing this was to learn the Go language.

(Alfred forum's post for this workflow: [http://goo.gl/zVgTFW](http://goo.gl/zVgTFW))

## Features
Pinboard is a great and reliable bookmarking service. Its [front page](https://pinboard.in) sums it all:
"***Social Bookmarking for Introverts. Pinboard is a fast, no-nonsense bookmarking site.***"

This plugin will let you:

- _**post**_ a bookmark to Pinboard right from Alfred. It has 'tag' autocomplete feature that will help you in selecting proper tags for the bookmark.
- _**search**_ your current bookmarks

For posting you just need to enter the Workflow's keyword ( <kbd>p</kbd> ) into Alfred's window and follow it with couple of tags and an optional description. The workflow will then post a bookmark for the current active window of your favorite browser to Pinboard.

For searching, use ( <kbd>ps</kbd> ) and then type the search keywords.

## Installation
After [downloading](https://bitbucket.org/listboss/go-pinboard/downloads) the latest version of the workflow and installing it in Alfred, you need to do a one-time setup to authenticate the Workflow. This Workflow only uses username/token method so you won't need to enter your password. (This is the *suggested* way of using Pinboard's API).
If you don't have a token, get one from Pinbaord's [settings page](https://pinboard.in/settings/password).

Then invoke Alfred and enter your username:token after the ***"pa"*** keyword:

![image](https://bitbucket.org/listboss/go-pinboard/wiki/authentication.png)

This workflow will keep a local cache of the tags that you have in Pinboard. To update the cache, you need to issue the ***"pu"*** command:

![image](https://bitbucket.org/listboss/go-pinboard/wiki/update.png)

Use this command once in while to keep the local cache up-to-date.

## Usage (post a bookmark):
The syntax to post a bookmark to Pinboard is :

```
p tag1 tag2 tag3 ; some optional note
```

The workflow will show a list of your current tags as you enter the command:

![image](https://bitbucket.org/listboss/go-pinboard/wiki/tag-suggestion-1.png)

The number below each tag shows how many times you have used it in your Pinboard service.
You can move Alfred's highlighter to the desired tag and hit '**Tab**' to autocomplete it.

![image](https://bitbucket.org/listboss/go-pinboard/wiki/tag-suggestion-2.png)

To finish the process just press Enter.
If you want to add extra description to the bookmark you can add it after the semi-colon:

![image](https://bitbucket.org/listboss/go-pinboard/wiki/adding-notes.png)

## Usage (search bookmarks):
Searching your bookmarks is easy.

```
ps search1 search2 search3
```

Workflow will use the text you enter in Alfred and show list of bookmarks that contain all of the search keywords in any of the bookmarks information (Desrciption of bookmark, its tags and url and its extended notes). So **the more** search keywords you enter **the less** results will be displayed as it tries to find the bookmarks that contain ***all*** of the keywords.

The search result is ordered in descending order of dates they were posted to your Pinboard account.

![image](https://bitbucket.org/listboss/go-pinboard/wiki/bookmarks-search-results.png)

## Additional Settings
- The default browser that this workflow uses is Chrome, you can change it to Safari using
```pset browser``` and then select one of the supported browsers.

![image](https://bitbucket.org/listboss/go-pinboard/wiki/set-browser.png)

- To enable/disable fuzzy search of tags/bookmarks, use:
```pset fuzzy``` and then select one of the options.

![image](https://bitbucket.org/listboss/go-pinboard/wiki/set-fuzzy.png)

When fuzzy search is enabled, the tags/bookmarks that contain the query letters in the given order are displayed:

![image](https://bitbucket.org/listboss/go-pinboard/wiki/fuzzy-search-tags.png)

- If you only want to search tags (using ```ps tag_name``` command), set the options like this:
	- ```pset tagonly```

	![image](https://bitbucket.org/listboss/go-pinboard/wiki/tag_only.png)

- To enable/disable sharing the bookmarks, use:
	- ```pset shared```

	![image](https://bitbucket.org/listboss/go-pinboard/wiki/shared.png)

- You can set the max. number of tags or bookmarks that you want to be displayed using these two keywords:
	- ```pset tags```
	- ```pset bmarks```

	![image](https://bitbucket.org/listboss/go-pinboard/wiki/set-max-tags.png)
	![image](https://bitbucket.org/listboss/go-pinboard/wiki/set-max-bmarks.png)

- You can show the current settings value using:
```pset show```

	Hitting enter on each item will let you set that option.

![image](https://bitbucket.org/listboss/go-pinboard/wiki/set-show.png)

## Alfred Helper for Go
In the process of writing this workflow, I have written a small but full feature Go package that helps with the development of the Alfred workflows in Go. It's very similar to [other Alfred Helpers](http://dferg.us/workflows-class/) just written in Go language, you can check it out at:

[https://bitbucket.org/listboss/go-alfred](https://bitbucket.org/listboss/go-alfred)

## TODO / Missing features

I wish to add the following in the coming releases:

- ~~Add capability to search the bookmark within Alfred~~
- Add popular (public) tags to the tag suggestion list.
- Add support for [Delicious](https://delicious.com/) API.


## Feedback / Bugs
This is my first non-trivial project using Go language so so your [feedback or bug](https://bitbucket.org/listboss/go-pinboard/issues?status=new&status=open) reports are greatly appreciated.

