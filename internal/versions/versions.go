package versions

import (
	"encoding/json"

	"github.com/anibaldeboni/rapper/internal/web"

	version "github.com/hashicorp/go-version"
)

const (
	releaseUrl = "https://api.github.com/repos/anibaldeboni/rapper/releases?per_page=1"
)

var headers = map[string]string{
	"Accept":               "application/vnd.github+json",
	"User-Agent":           "rapper",
	"X-GitHub-Api-Version": "2022-11-28",
}

type release struct {
	TagName string `json:"tag_name"`
	HtmlUrl string `json:"html_url"`
}

type Update struct {
	Available bool
	Version   string
	Url       string
}

func CheckForUpdate(hc web.HttpClient, currentVersion string) (Update, bool) {
	var update Update
	var ok bool

	res, err := hc.Get(releaseUrl, headers)
	if err != nil {
		return update, ok
	}
	var releases []release
	err = json.Unmarshal(res.Body, &releases)
	if err != nil {
		return update, ok
	}

	current, err := version.NewVersion(currentVersion)
	if err != nil {
		return update, ok
	}
	latest, err := version.NewVersion(releases[0].TagName)
	if err != nil {
		return update, ok
	}

	if latest.GreaterThan(current) {
		update.Version = releases[0].TagName
		update.Url = releases[0].HtmlUrl
		update.Available = true
		ok = true
	}
	return update, ok
}
