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

type update struct {
	available bool
	version   string
	url       string
}

type Update interface {
	HasUpdate() bool
	Version() string
	Url() string
}

func (u update) HasUpdate() bool {
	return u.available
}

func (u update) Version() string {
	return u.version
}

func (u update) Url() string {
	return u.url
}

type UpdateChecker interface {
	CheckForUpdate() Update
}

type updateChecker struct {
	hc             web.HttpClient
	currentVersion string
}

func NewUpdateChecker(hc web.HttpClient, currentVersion string) UpdateChecker {
	return &updateChecker{hc: hc, currentVersion: currentVersion}
}

func (this updateChecker) CheckForUpdate() Update {
	var update update

	res, err := this.hc.Get(releaseUrl, headers)
	if err != nil {
		return update
	}
	var releases []release
	err = json.Unmarshal(res.Body, &releases)
	if err != nil {
		return update
	}

	current, err := version.NewVersion(this.currentVersion)
	if err != nil {
		return update
	}
	latest, err := version.NewVersion(releases[0].TagName)
	if err != nil {
		return update
	}

	if latest.GreaterThan(current) {
		update.version = releases[0].TagName
		update.url = releases[0].HtmlUrl
		update.available = true
	}
	return update
}
