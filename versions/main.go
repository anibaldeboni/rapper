package versions

import (
	"encoding/json"
	"rapper/ui"
	"rapper/web"

	version "github.com/hashicorp/go-version"
)

const (
	releaseUrl = "https://api.github.com/repos/anibaldeboni/rapper/releases"
	NoUpdates  = ""
)

type release struct {
	TagName string `json:"tag_name"`
	HtmlUrl string `json:"html_url"`
}

func CheckForUpdate(hc web.HttpClient, currentVersion string) string {
	res, err := hc.Get(releaseUrl, nil)
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
