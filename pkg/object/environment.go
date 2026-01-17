package object

import "sync"

// Environment stores variable bindings
type Environment struct {
	store  map[string]Object
	outer  *Environment
	consts map[string]bool // Track constant variables
	mu     sync.RWMutex    // For concurrent access safety
}

// NewEnvironment creates a new environment
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	c := make(map[string]bool)
	return &Environment{store: s, outer: nil, consts: c}
}

// NewEnclosedEnvironment creates a new enclosed environment
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Get retrieves a variable from the environment
func (e *Environment) Get(name string) (Object, bool) {
	e.mu.RLock()
	obj, ok := e.store[name]
	e.mu.RUnlock()
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set sets a variable in the environment
func (e *Environment) Set(name string, val Object) Object {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.store[name] = val
	return val
}

// SetConst sets a constant in the environment
func (e *Environment) SetConst(name string, val Object) Object {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.store[name] = val
	e.consts[name] = true
	return val
}

// IsConst checks if a variable is a constant
func (e *Environment) IsConst(name string) bool {
	e.mu.RLock()
	isConst, ok := e.consts[name]
	e.mu.RUnlock()
	if !ok && e.outer != nil {
		return e.outer.IsConst(name)
	}
	return isConst
}

// Update updates an existing variable
func (e *Environment) Update(name string, val Object) (Object, bool) {
	e.mu.Lock()
	_, ok := e.store[name]
	e.mu.Unlock()
	
	if ok {
		if e.IsConst(name) {
			return nil, false // Cannot modify constants
		}
		e.mu.Lock()
		e.store[name] = val
		e.mu.Unlock()
		return val, true
	}
	
	if e.outer != nil {
		return e.outer.Update(name, val)
	}
	
	return nil, false
}

// Exists checks if a variable exists in current scope
func (e *Environment) Exists(name string) bool {
	e.mu.RLock()
	_, ok := e.store[name]
	e.mu.RUnlock()
	return ok
}
