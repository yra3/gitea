package shared

//TODO everythings ^^

// Method is the interface that we're exposing as a method plugin that override gitea method.
type Method interface {
	Methods() []string
	Get(string) error
}
