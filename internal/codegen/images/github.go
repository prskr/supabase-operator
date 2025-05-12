package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"code.icb4dc0.de/prskr/supabase-operator/internal/errx"
)

type releaseMatcher interface {
	matchesRelease(r release) bool
}

var (
	excludeDrafts releaseMatcher = releaseMatcherFunc(func(r release) bool {
		return !r.Draft
	})
	excludePreReleases releaseMatcher = releaseMatcherFunc(func(r release) bool {
		return !r.PreRelease
	})
)

func matchesTagPattern(pattern string) releaseMatcher {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}
	return releaseMatcherFunc(func(r release) bool {
		return compiled.MatchString(r.TagName)
	})
}

type release struct {
	TagName    string `json:"tag_name"`
	Draft      bool   `json:"draft"`
	PreRelease bool   `json:"prerelease"`
}

//nolint:err113
func latestReleaseVersion(ctx context.Context, owner, repo string, matchers ...releaseMatcher) (tagName string, err error) {
	releaseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releaseURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer errx.Close(resp.Body, &err)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("failed to retrieve latest release: %s", resp.Status)
	}

	var (
		releases []release
		matcher  = multiMatcher(matchers)
	)

	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", err
	}

	for _, release := range releases {
		if matcher.matchesRelease(release) {
			return release.TagName, nil
		}
	}

	return "", fmt.Errorf("no release found matching the criteria: %s/%s", owner, repo)
}

type multiMatcher []releaseMatcher

func (m multiMatcher) matchesRelease(r release) bool {
	for _, matcher := range m {
		if !matcher.matchesRelease(r) {
			return false
		}
	}
	return true
}

type releaseMatcherFunc func(r release) bool

func (f releaseMatcherFunc) matchesRelease(r release) bool {
	return f(r)
}
