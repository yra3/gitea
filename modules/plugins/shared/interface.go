package shared

import (
	"net/rpc"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/ngaut/log"
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "GITEA_PLUGIN",
	MagicCookieValue: "hello",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"plugin": &PluginImpl{}, //Mandatory to define what the plugin is capable
	"router": &RouterImpl{},
	"method": &MethodImpl{},
}

// Plugin is the interface that we're exposing that setup the plugin and inform the main process of capabilities
type Plugin interface {
	Details() PluginDetails
}

type PluginImpl struct {
	// Impl Injection
	Impl Plugin
}

//PluginDetails give details about the plugins and what it can do.
type PluginDetails struct {
	Name         string
	Version      string
	Description  string
	State        PluginState
	Capabilities []PluginCapabilities
}

//PluginState reprensente the current state of a plugin
type PluginState int

const (
	//PluginStateBoot plugin is starting and not ready to do some work
	PluginStateBoot = PluginState(iota)
	//PluginStateReady plugin is ready to handle work
	PluginStateReady
	//PluginStateWorking plugin is working and maybe need some time to handle the next request
	PluginStateWorking //Please wait
	//PluginStateError plugin is in a failed state and should not execute task.
	PluginStateError
)

//PluginCapabilities reprensente the capabilities of a plugin
type PluginCapabilities int // running //TODO define a list

const (
	//PluginHandleRoute is a router
	PluginHandleRoute = PluginCapabilities(iota)
	//PluginHandleMethods can override some methods of gitea
	PluginHandleMethods
)

func (p *PluginImpl) Server(*plugin.MuxBroker) (interface{}, error) {
	return &PluginRPCServer{Impl: p.Impl}, nil
}

func (PluginImpl) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PluginRPC{client: c}, nil
}

// Here is the RPC server that GreeterRPC talks to, conforming to
// the requirements of net/rpc
type PluginRPCServer struct {
	// This is the real implementation
	Impl Plugin
}

func (s *PluginRPCServer) Details(args interface{}, resp *PluginDetails) error {
	*resp = s.Impl.Details()
	return nil
}

//TODO GRPC

// Here is an implementation that talks over RPC
type PluginRPC struct{ client *rpc.Client }

func (g *PluginRPC) Details() PluginDetails {
	var resp PluginDetails
	err := g.client.Call("Plugin.Details", new(interface{}), &resp)
	if err != nil {
		log.Error("Plugin RPC error: %v", err)
	}
	return resp
}
