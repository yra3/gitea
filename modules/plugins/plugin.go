// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package plugins

import (
	"code.gitea.io/gitea/modules/plugins/shared"
	plugin "github.com/hashicorp/go-plugin"
)

// Plugin represents a plugin and the process inherit from Gitea.
type Plugin struct {
	ID   int64
	Path string
	//TODO use as cache in GetDetails Details shared.PluginDetails
	Client *plugin.Client
}

//GetDetails start the ProtocolClient of a plugin to get details on it.
func (p *Plugin) GetDetails() (*shared.PluginDetails, error) {

	// Connect via RPC
	rpcClient, err := p.Client.Client()
	if err != nil {
		return nil, err
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("plugin")
	if err != nil {
		return nil, err
	}

	// We should have a Greeter now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	plugin := raw.(shared.Plugin)

	details := plugin.Details()
	return &details, nil
}
