package Core

// base app struct.
// any app should extend this base struct.
// Context struct contains this struct as property Context.App .
type Base struct {
	Root   string
	Config *Config
	Router *Router
	Listener *Listener
	Logger LoggerInterface
	View *View
}
