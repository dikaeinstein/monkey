package object

// Environment associates values with variable names.
// It is a stack of frames, and each frame is a list of variable bindings.
type Environment struct {
	frames []*Frame
}

// NewEnvironment creates an environment initialized with the `main` stack frame.
func NewEnvironment() *Environment {
	main := &Frame{store: make(map[string]Object), prev: nil}
	return &Environment{[]*Frame{main}}
}

// AddFrame creates a new Frame and push it to the top of the stack.
func (env *Environment) AddFrame() *Environment {
	f := &Frame{store: make(map[string]Object), prev: env.TopFrame()}
	env.frames = append(env.frames, f)
	return env
}

// DestroyFrame pops the current top frame from stack.
func (env *Environment) DestroyFrame() {
	// pop from stack
	env.frames = env.frames[:len(env.frames)-1]
}

func (env *Environment) TopFrame() *Frame {
	return env.frames[len(env.frames)-1]
}

// Get returns the value associated to the variable from the environment .
func (env *Environment) Get(name string) (Object, bool) {
	for i := len(env.frames) - 1; i >= 0; i-- {
		if obj, ok := env.frames[i].Get(name); ok {
			return obj, ok
		}
	}

	return nil, false
}

// Set binds a value to the variable name and adds it to the
// frame on the top of the stack.
func (env *Environment) Set(name string, value Object) Object {
	return env.TopFrame().Set(name, value)
}

// Frame stores variable bindings.
type Frame struct {
	store map[string]Object
	prev  *Frame
}

// Get returns the value associated to the variable name in this stack frame.
func (f *Frame) Get(name string) (Object, bool) {
	value, ok := f.store[name]
	return value, ok
}

// Set binds a value to the variable name in the stack frame.
func (f *Frame) Set(name string, value Object) Object {
	f.store[name] = value
	return value
}
