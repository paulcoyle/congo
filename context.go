package congo

import (
  "net/http"
)

type Context interface {
  ResponseWriter() http.ResponseWriter
  Request() *http.Request
  Content() string
  Write([]byte) (int, error)
}

type BaseContext struct {
  responseWriter http.ResponseWriter
  request        *http.Request
  // content holds intermediate template results.  In Congo, the convention
  // (as in many large web frameworks) is to render a template specific to the
  // current action then inject it's result into a layout template that is
  // generally common to many actions.  This value is used to store the
  // rendered action template then passed to the layout template which will
  // call {{.Content}} to insert the action template result into the layout.
  // This is similar to Rails' <%= yield %>.
  content string
}

func NewBaseContext(w http.ResponseWriter, r *http.Request) *BaseContext {
  return &BaseContext{
    responseWriter: w,
    request:        r,
  }
}

// Returns the current response writer associated with the context.
func (c *BaseContext) ResponseWriter() http.ResponseWriter {
  return c.responseWriter
}

// Returns the current request associated with the context.
func (c *BaseContext) Request() *http.Request {
  return c.request
}

// Returns the currently defined content.  In action templates this will be
// not be set and is really only of interest when injecting action template
// results into layouts.
func (c *BaseContext) Content() string {
  return c.content
}

// This is used to implement the Writer interface so that we can render the
// action template directly into the context for later use in layout
// templates.
func (c *BaseContext) Write(data []byte) (int, error) {
  c.content = c.content + string(data)
  return len(data), nil
}
