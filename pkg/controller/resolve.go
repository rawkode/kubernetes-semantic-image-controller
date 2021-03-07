package controller

type Resolver interface {
	// Resolve takes in the image input as a string as specified by the
	// user from K8s and returns a string with the qualified and resolved
	// image name (or an error if it fails to resolve)
	Resolve(input string) (string, error)

	// ShouldResolve takes in the image input as a string as specified by the
	// user from K8s and returns a bool to indicate whether this is an image
	// we should consider resolving
	ShouldResolve(input string) bool
}
