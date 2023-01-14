package main

import "strings"

func (app *application) getBaseUrl(url string) string {
	url_split := strings.SplitAfter(url, "/")
	base_url := url_split[1] + url_split[2]
	base_url_trim := base_url[1 : len(base_url)-1]

	return base_url_trim
}
