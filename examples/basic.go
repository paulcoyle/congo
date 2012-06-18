package main

// Run this (go run basic.go) and then point your browser to:
//
//     http://localhost:8080/
//     http://localhost:8080/lion
//
// Ensure that the library is in your GOPATH with:
//
//     go get github.com/paulcoyle/congo

import (
	"github.com/paulcoyle/congo"
	"html/template"
	"net/http"
)

// Wraps a context so that we can add some information about the secure status
// of the current request.
type SecureContext struct {
	congo.Context
}

// Checks if the request is using TLS or, if you're behind a reverse proxy,
// the X-Forwarded-Proto header is set to https.
func (c *SecureContext) IsSecure() bool {
	https, ok := c.Request().Header["X-Forwarded-Proto"]
	return (ok && len(https) > 0 && https[0] == "https") || (c.Request().TLS != nil)
}

// Wraps the secure context so that we know what a lion sounds like
// (terrifyingly awesome).
type LionContext struct {
	*SecureContext
	sound string
}

// Emits the sound of a ferocious lion.
func (c *LionContext) Sound() string {
	return c.sound
}

// Augments the base context with our custom SecureContext.  The methods in
// the base context are promoted within the SecureContext so it automatically
// implements the Context interface.
func augmentContext(c congo.Context) (congo.Context, interface{}) {
	return &SecureContext{c}, nil
}

// Simply renders the home template within the layout with the default context.
func homeHandler(c congo.Context) (congo.Context, interface{}) {
	return c, &congo.RenderResponse{"home", "layout"}
}

// The lion handler augments the secure context for a lion template specific
// method.  When doing things like this it is always good to be mindful of
// what actions have been run beforehand and what you should be expecting
// with respect to the inbound context.
func lionHandler(c congo.Context) (congo.Context, interface{}) {
	return &LionContext{c.(*SecureContext), "RAWR"},
		&congo.RenderResponse{"lion", "layout"}
}

// This just creates some very simple templates to illustrate how augmented
// contexts can be useful when working in templates.
func createTemplates() *template.Template {
	root, _ := template.New("layout").
		Parse("LAYOUT START (SECURE: {{.IsSecure}})\n{{.Content}}\nLAYOUT END")

	root.New("home").Parse("----HOME----")
	root.New("lion").Parse("----LION----\n----{{.Sound}}----")

	return root
}

func main() {
	// This sets the given template as the default template store for all
	// handlers created hereafter.
	congo.SetDefaultTemplateStore(createTemplates())

	base := congo.NewHandler().Actions(augmentContext)
	// The following handlers copy from base handler so they will automatically
	// contain augmentContext as their first action.
	home := base.Copy().Actions(homeHandler)
	lion := base.Copy().Actions(lionHandler)

	// Handlers can be adapted to standard http.HandleFunc using
	// congo.MuxHandler.  This also works with the gorilla web toolkit.
	http.HandleFunc("/", congo.MuxHandler(home))
	http.HandleFunc("/lion", congo.MuxHandler(lion))
	http.ListenAndServe(":8080", nil)
}
