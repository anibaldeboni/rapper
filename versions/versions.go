package versions

import (
	"encoding/json"

	"github.com/anibaldeboni/rapper/cli/ui"
	"github.com/anibaldeboni/rapper/web"

	version "github.com/hashicorp/go-version"
)

const (
	releaseUrl = "https://api.github.com/repos/anibaldeboni/rapper/releases?per_page=1?page=1"
	NoUpdates  = ""
)

type release struct {
	TagName string `json:"tag_name"`
	HtmlUrl string `json:"html_url"`
}

func CheckForUpdate(hc web.HttpClient, currentVersion string) string {
	headers := map[string]string{
		"Accept":               "application/vnd.github+json",
		"User-Agent":           "rapper",
		"X-GitHub-Api-Version": "2022-11-28",
	}
	res, err := hc.Get(releaseUrl, headers)
	if err != nil {
		return NoUpdates
	}
	var releases []release
	err = json.Unmarshal(res.Body, &releases)
	if err != nil {
		return NoUpdates
	}

	current, err := version.NewVersion(currentVersion)
	if err != nil {
		return NoUpdates
	}
	latest, err := version.NewVersion(releases[0].TagName)
	if err != nil {
		return NoUpdates
	}

	if latest.GreaterThan(current) {
		str := ui.IconInformation + "  New version available: " + ui.Bold(releases[0].TagName) + " (" + releases[0].HtmlUrl + ")"
		return str
	}
	return NoUpdates
}
