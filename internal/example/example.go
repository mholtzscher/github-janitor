// Package example provides example internal functionality.
package example

import "fmt"

// Greeter provides greeting functionality.
type Greeter struct {
	Name string
}

// NewGreeter creates a new Greeter.
func NewGreeter(name string) *Greeter {
	return &Greeter{Name: name}
}

// Greet returns a greeting message.
func (g *Greeter) Greet() string {
	return fmt.Sprintf("Hello, %s!", g.Name)
}
