// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package plugins

import (
	"fmt"
	"os/exec"
	"sync"

	plugin "github.com/hashicorp/go-plugin"

	"code.gitea.io/gitea/modules/log"
)

var (
	manager *Manager
)

// Manager knows about all processes and counts PIDs.
type Manager struct {
	mutex   sync.Mutex
	counter int64
	Plugins map[int64]*Plugin
}

// GetManager returns a Manager and initializes one as singleton if there's none yet
func GetManager() *Manager {
	if manager == nil {
		manager = &Manager{
			Plugins: make(map[int64]*Plugin),
		}
	}
	return manager
}

//Add a plugin to the list
func (pm *Manager) Add(path string) {
	pm.mutex.Lock()
	id := pm.counter
	pm.Plugins[id] = &Plugin{
		ID:     id,
		Path:   path,
		Client: nil,
	}
	pm.counter = id + 1
	pm.mutex.Unlock()
	return id
}

//Remove stop and remove a plugin from the list
func (pm *Manager) Remove(id int64) error {
	//TODO force to end/remove plugin in any cases
	err := pm.Stop(id)
	if err != nil {
		return err
	}
	pm.mutex.Lock()
	delete(pm.Plugins, id)
	pm.mutex.Unlock()
}

//Start a plugin in the list
func (pm *Manager) Start(id int64) error { //TODO don't use defer on lock
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	if _, ok := pm.Plugins[id]; !ok {
		return fmt.Errorf("plugin(%d) not found", id)
	}
	p := pm.Plugins[id]
	p.Client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         shared.PluginMap,
		Cmd:             exec.Command(p.Path), //TODO check/clean name
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	})

	//TODO Testing
	details, err := p.GetDetails()
	if err != nil {
		log.Error(2, "plugin(%d:%s) GetDetails failed (%v): %v", id, p.Path, err, details)
		return err
	}

	log.Debug("plugin(%d:%s) GetDetails: %v", id, p.Path, details)
	return nil
}

//Stop stop and reset the selected plugin
func (pm *Manager) Stop(id int64) {
	pm.mutex.Lock()
	if _, ok := pm.Plugins[id]; !ok {
		pm.mutex.Unlock()
		return fmt.Errorf("plugin(%d) not found", id)
	}
	p := pm.Plugins[id]
	if p.Client != nil {
		p.Client.Kill()
		p.Client = nil
	}
	pm.mutex.Unlock()
}

/*
func (pm *Manager) Fetch() {
}

func (pm *Manager) ImportLocalPlugins() {
}
*/
