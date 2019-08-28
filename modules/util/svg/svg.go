// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package svg

import "github.com/microcosm-cc/bluemonday"

// SanitizeSVG remove potential malicious dom elements
func SanitizeSVG(svg string) string {
	p := bluemonday.UGCPolicy()
	return p.Sanitize(svg)
}
