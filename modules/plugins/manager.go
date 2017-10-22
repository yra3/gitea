// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package plugins

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/hashicorp/go-plugin"
	"code.gitea.io/gitea/modules/log"
)

var (
	manager        *Manager
)

// Process represents a working process inherit from Gogs.
type Plugin struct {
	Description string
	Process         *process.Process
}

// Manager knows about all processes and counts PIDs.
type Manager struct {
	mutex sync.Mutex
	counter   int64
	Plugins map[int64]*Plugin
}

// GetManager returns a Manager and initializes one as singleton if there's none yet
func GetManager() *Manager {
	if manager == nil {
		client := plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig: shared.Handshake,
			Plugins:         shared.PluginMap,
			Cmd:             exec.Command("sh", "-c", os.Getenv("KV_PLUGIN")),
			AllowedProtocols: []plugin.Protocol{
				plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
		})
		manager = &Manager{
			Processes: make(map[int64]*Plugin),
		}
	}
	return manager
}

func (pm *Manager) Add() {
}

func (pm *Manager) Remove() {
}

func (pm *Manager) Start() {
}

func (pm *Manager) Stop() {
}
