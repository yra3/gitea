package shared

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "GITEA_PLUGIN",
	MagicCookieValue: "hello",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"plugin": &Plugin{}, //Mandatory to define what the plugin is capable
	"router": &Router{},
	"method": &Method{},
}

// Plugin is the interface that we're exposing that setup the plugin and inform the main process of capabilities
type Plugin interface {
	Details() PluginDetails
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
