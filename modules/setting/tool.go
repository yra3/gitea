// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package setting

import (
	"net/url"
	"strings"
)

//validate https://tools.ietf.org/html/rfc6265#section-5.1.3
func isCookieSameDomain(app *url.URL, raw *url.URL) bool {
	appHost := strings.ToLower(app.Hostname())
	rawHost := strings.ToLower(raw.Hostname())

	if rawHost == "" || appHost == "" {
		return true
	}
	if rawHost == appHost {
		return true
	}

	return strings.HasSuffix(rawHost, "."+appHost)
}
