package main

import "github.com/fuxiaohei/GoInk"

var (
	app *GoInk.App
)

func init() {
	// init new application.
	// it loads default config file "config.json". if not exist, use pre-defined config.
	app = GoInk.New()
}

// default handler, implement GoInk.Handler
func homeHandler(ctx *GoInk.Context) {
	ctx.Body = []byte("Hello GoInk !")
}

func main() {
	// only bind GET handler by homeHandler.
	// if other http method, return 404.
	app.Get("/", homeHandler)

	// run application.
	// it listens localhost:9000 in pre-defined config.
	app.Run()
}
