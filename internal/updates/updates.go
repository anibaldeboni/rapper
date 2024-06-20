package updates

import (
	"context"
	"encoding/json"

	"github.com/anibaldeboni/rapper/internal/web"

	gv "github.com/hashicorp/go-version"
)

var (
	releaseUrl = "https://api.github.com/repos/anibaldeboni/rapper/releases?per_page=1"
	client     = web.NewHttpClient()
	headers    = map[string]string{
		"Accept":               "application/vnd.github+json",
		"User-Agent":           "rapper",
		"X-GitHub-Api-Version": "2022-11-28",
	}
)

type releaseInfo struct {
	TagName string `json:"tag_name"`
	HtmlUrl string `json:"html_url"`
}

type updateDetails struct {
	Version string
	Url     string
}

func CheckFor(version string) (details updateDetails, hasUpdate bool) {
	res, err := client.Get(context.Background(), releaseUrl, headers)
	if err != nil {
		return
	}

	var releases []releaseInfo
	err = json.Unmarshal(res.Body, &releases)
	if err != nil {
		return
	}

	current, err := gv.NewVersion(version)
	if err != nil {
		return
	}

	latest, err := gv.NewVersion(releases[0].TagName)
	if err != nil {
		return
	}

	if latest.GreaterThan(current) {
		details.Version = releases[0].TagName
		details.Url = releases[0].HtmlUrl
		hasUpdate = true
	}

	return
}
