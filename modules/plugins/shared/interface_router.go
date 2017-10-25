package shared

//TODO everythings ^^

// Router is the interface that we're exposing as a router plugin.
type Router interface {
	Routes() []string
	Handle(string) error
}
