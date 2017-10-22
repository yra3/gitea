package shared


import	"github.com/hashicorp/go-plugin"

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "GITEA_PLUGIN",
	MagicCookieValue: "hello",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"init": &Plugin{}, //Mandatory to define what the plugin is capable
	"router": &Router{},
  "method": &Method{},
}


// Plugin is the interface that we're exposing that setup the plugin and inform the main process of capabilities
type Plugin interface {
	Details() PluginDetails
	Capabilities() []string
}

type PluginDetails struct {
	Name string
	Description string
	State PluginState // running,
	Capabilities []string // router, method //TODO define a list
}

type PluginState string  // running //TODO define a list
