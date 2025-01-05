package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"code.icb4dc0.de/prskr/supabase-operator/internal/errx"
)

func latestReleaseVersion(ctx context.Context, owner, repo string) (tagName string, err error) {
	releaseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
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

	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}
