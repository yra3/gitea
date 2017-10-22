package shared

// Method is the interface that we're exposing as a method plugin that override gitea method.
type Method interface {
	Methods() []string
	Get(key string) (interface, error) //TODO return func that override gitea methods
}
