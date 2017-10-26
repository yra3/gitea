package shared

import (
	"net/rpc"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/ngaut/log"
)

//TODO everythings ^^

// Router is the interface that we're exposing as a router plugin.
type Router interface {
	Routes() []string
	Handle(string) error
}

type RouterImpl struct {
	// Impl Injection
	Impl Router
}

func (p *RouterImpl) Server(*plugin.MuxBroker) (interface{}, error) {
	return &RouterRPCServer{Impl: p.Impl}, nil
}

func (RouterImpl) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RouterRPC{client: c}, nil
}

/* TODO
func (p *RouterImpl) GRPCServer(s *grpc.Server) error {
	proto.RegisterKVServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *RouterImpl) GRPCClient(c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewKVClient(c)}, nil
}
*/

// Here is the RPC server that GreeterRPC talks to, conforming to
// the requirements of net/rpc
type RouterRPCServer struct {
	// This is the real implementation
	Impl Router
}

func (s *RouterRPCServer) Routes(args interface{}, resp *[]string) error {
	*resp = s.Impl.Routes()
	return nil
}

//TODO improve with interface
func (s *RouterRPCServer) Handle(args string, err *error) error {
	*err = s.Impl.Handle(args)
	return nil
}

// Here is an implementation that talks over RPC
type RouterRPC struct{ client *rpc.Client }

func (g *RouterRPC) Routes() []string {
	var resp []string
	err := g.client.Call("Plugin.Routes", new(interface{}), &resp)
	if err != nil {
		log.Error("Plugin RPC error: %v", err)
	}
	return resp
}

func (g *RouterRPC) Handle(m string) error {
	var resp error
	err := g.client.Call("Plugin.Handle", m, &resp)
	if err != nil {
		log.Error("Plugin RPC error: %v", err)
	}
	return resp
}
