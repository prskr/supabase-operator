/*
Copyright 2025 Peter Kurfer.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package oidc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var ErrUnexpectedStatusCode = errors.New("unexpected status code")

type DiscoveryDocument struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
}

func IssuerConfiguration(ctx context.Context, issuerUrl string) (dd DiscoveryDocument, err error) {
	const oidcDiscoveryEndpoint = "/.well-known/openid-configuration"
	if !strings.HasSuffix(issuerUrl, oidcDiscoveryEndpoint) {
		issuerUrl = strings.TrimSuffix(issuerUrl, "/") + oidcDiscoveryEndpoint
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, issuerUrl, nil)
	if err != nil {
		return dd, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return dd, err
	}

	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode != http.StatusOK {
		return dd, fmt.Errorf("%w: %d - %s", ErrUnexpectedStatusCode, resp.StatusCode, resp.Status)
	}

	return dd, json.NewDecoder(resp.Body).Decode(&dd)
}
