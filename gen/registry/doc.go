// Package registry provides a global type registry for the goconfygen CLI tool.
//
// Target projects register their config types by implementing the [Provider]
// interface and calling [Register] in an init() function. The CLI then looks
// up providers by ID to generate templates, validate configs, and dump output.
//
// The registry is safe for concurrent use.
//
// Example registration in a target project:
//
//	package config
//
//	import "github.com/keksclan/goConfy/gen/registry"
//
//	type configProvider struct{}
//
//	func (configProvider) ID() string { return "myservice" }
//	func (configProvider) New() any   { return &Config{} }
//
//	func init() {
//	    registry.Register(configProvider{})
//	}
package registry
