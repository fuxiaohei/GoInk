package Core

// Base app struct.
type Base struct {
	Root   string
	Config *Config
	Router *Router
	Listener *Listener
	Logger LoggerInterface
	View *View
}
