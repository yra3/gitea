// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package setting

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_isCookieSameDomain(t *testing.T) {
	type args struct {
		app *url.URL
		raw *url.URL
	}
	tests := []struct {
		name string
		app  string
		raw  string
		want bool
	}{
		{"IPv4-1", "http://127.0.0.1", "http://127.0.0.1", true},
		{"Domain-1", "http://gitea.com", "http://gitea.com", true},
		{"Domain-2", "http://gitea.com", "http://www.gitea.com", true},
		{"Domain-3", "http://gitea.com", "http://raw.gitea.com", true},
		{"IPv4-2", "http://127.0.0.1", "http://127.0.0.10", false},
		{"Domain-4", "http://gitea.com", "http://raw.giteausercontent.com", false},
		{"Domain-5", "http://www.gitea.com", "http://raw.giteausercontent.com", false},
		{"Domain-6", "http://www.gitea.com", "http://gitea.com.giteausercontent.com", false},
		{"IPv6-1", "http://[2001:db8::1]", "http://[2001:db8::1]", true},
		{"IPv6-2", "http://[2001:db8::1]", "http://[2001:db8::2]", false},
		{"Undefined domain", "http://[2001:db8::1]", "http://", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appURL, err := url.Parse(tt.app)
			assert.NoError(t, err)
			rawURL, err := url.Parse(tt.raw)
			assert.NoError(t, err)
			if got := isCookieSameDomain(appURL, rawURL); got != tt.want {
				t.Errorf("isCookieSameDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}
