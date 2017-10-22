package shared

// Router is the interface that we're exposing as a router plugin.
type Router interface {
	Routes() []string
	Handle(key string) (interface, error) //TODO
}
