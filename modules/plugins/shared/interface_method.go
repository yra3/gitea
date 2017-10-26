package shared

import (
	"net/rpc"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/ngaut/log"
)

//TODO everythings ^^

// Method is the interface that we're exposing as a method plugin that override gitea method.
type Method interface {
	Methods() []string
	Get(string) error
}

type MethodImpl struct {
	// Impl Injection
	Impl Method
}

func (p *MethodImpl) Server(*plugin.MuxBroker) (interface{}, error) {
	return &MethodRPCServer{Impl: p.Impl}, nil
}

func (MethodImpl) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &MethodRPC{client: c}, nil
}

// Here is the RPC server that GreeterRPC talks to, conforming to
// the requirements of net/rpc
type MethodRPCServer struct {
	// This is the real implementation
	Impl Method
}

func (s *MethodRPCServer) Methods(args interface{}, resp *[]string) error {
	*resp = s.Impl.Methods()
	return nil
}

//TODO improve with interface
func (s *MethodRPCServer) Get(arg string, err *error) error {
	*err = s.Impl.Get(arg)
	return nil
}

// Here is an implementation that talks over RPC
type MethodRPC struct{ client *rpc.Client }

func (g *MethodRPC) Methods() []string {
	var resp []string
	err := g.client.Call("Plugin.Methods", new(interface{}), &resp)
	if err != nil {
		log.Error("Plugin RPC error: %v", err)
	}
	return resp
}

func (g *MethodRPC) Get(m string) error {
	var resp error
	err := g.client.Call("Plugin.Get", m, &resp)
	if err != nil {
		log.Error("Plugin RPC error: %v", err)
	}
	return resp
}
