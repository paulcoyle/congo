package congo

type Context interface {
  Content() string
  Write([]byte) (int, error)
}

type BaseContext struct {
  // content holds intermediate template results.  In Congo, the convention
  // (as in many large web frameworks) is to render a template specific to the
  // current action then inject it's result into a layout template that is
  // generally common to many actions.  This value is used to store the
  // rendered action template then passed to the layout template which will
  // call {{.Content}} to insert the action template result into the layout.
  // This is similar to Rails' <%= yield %>.
  content string
}

func NewBaseContext() *BaseContext {
  return &BaseContext{}
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
