package congo

import (
  "html/template"
  "io"
  "net/http"
  "sync"
)

// HandlerActions accept a Context interface and optionally return that
// same Context back or another implementation of Context paired with
// any of the responses available in responses.go as the second return value
// (optionally nil).  HandlerActions that return a response will cause the
// chain of actions to halt for its particular Handler when invoked.
type HandlerAction (func(Context) (Context, interface{}))

type Handler struct {
  templateStore *template.Template
  actions       []HandlerAction
  actionMutex   sync.RWMutex
}

// Creates a new handler and returns a pointer to it.
func NewHandler() *Handler {
  return &Handler{
    templateStore: defaultTemplateStore,
    actions:       make([]HandlerAction, 0),
  }
}

// Returns the currently defined template store for this Handler.
func (h *Handler) TemplateStore() *template.Template {
  return h.templateStore
}

// Sets the template store for the given handler.  Note that when rendering
// templates within other templates you will be restricted to those registered
// with the store of top-level template.  Therefore it is best to keep only
// one template store with all of your templates registered in it.  There is a
// method SetDefaultTemplateStore which need only be set once to have the
// provided template store applied to all Handlers.
func (h *Handler) SetTemplateStore(store *template.Template) *Handler {
  h.templateStore = store
  return h
}

// Appends a HandlerActions to the action chain for this Handler.  See the
// type description of HandlerAction for details.
func (h *Handler) Actions(a ...HandlerAction) *Handler {
  h.actionMutex.Lock()
  defer h.actionMutex.Unlock()
  for _, action := range a {
    h.actions = append(h.actions, action)
  }
  return h
}

// Copies the HandlerActions and template store for a Handler to a new
// instance and returns it.
func (h *Handler) Copy() *Handler {
  copy := NewHandler()
  copy.Actions(h.actions...)
  copy.templateStore = h.templateStore
  return copy
}

// Runs the chain of HandlerActions beginning with the given context.  The
// return values from this represent the final context and response given
// by the last HandlerAction to return a response.
func (h *Handler) applyActions(context Context) (Context, interface{}) {
  var response interface{} = nil

  h.actionMutex.RLock()
  defer h.actionMutex.RUnlock()

  for _, action := range h.actions {
    newContext, newResponse := action(context)
    // If the context return value is non-nil, check that it still implements
    // the Context interface.  If not, the method will panic.
    if newContext != nil {
      if assertContext, ok := newContext.(Context); ok {
        context = assertContext
      } else {
        panic("Resulting context does not implement congo.Context")
      }
    }

    // If an HandlerAction has returned a response then halt the chain.
    if newResponse != nil {
      response = newResponse
      break
    }
  }

  return context, response
}

// Returns a function that is compatible with the standard http Handler.
func MuxHandler(h *Handler) func(http.ResponseWriter, *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    context, response := h.applyActions(NewBaseContext(w, r))

    // We must have a response in order to do anything.
    if response == nil {
      panic("No response given after handler action chain executed")
    }

    h.responseRenderStep(context, response)
    h.responseFinalizeStep(context, response)
  }
}

// Performs any rendering needed for a given response.  This method is only
// concerned with any template rendering or related activity.  Any
// response types that do not deal with these activities are ignored here.
func (h *Handler) responseRenderStep(context Context, response interface{}) {
  switch response.(type) {
  case *RenderResponse:
    // In this step, the inner template is rendered and written to the context
    // where it's output is available from context.Content()
    r := response.(*RenderResponse)
    h.execTemplate(r.Template, context, context)
  case *NotFoundResponse:
    r := response.(*NotFoundResponse)
    h.execTemplate(r.Template, context, context)
  }
}

// Performs the finalizing step(s) for a given response.  This includes
// sending appropriate headers etc. as well as rendering the layouts with the
// contents of responses handled by responseRenderStep.
func (h *Handler) responseFinalizeStep(context Context, response interface{}) {
  switch response.(type) {
  default:
    panic("Unknown response type")
  case *NullResponse:
  case *RenderResponse:
    // In this step, the layout template is rendered which should make use of
    // the content from the render step by using {{.Content}} in the template.
    r := response.(*RenderResponse)
    h.execTemplate(r.Layout, context.ResponseWriter(), context)
  case *RedirectResponse:
    r := response.(*RedirectResponse)
    http.Redirect(context.ResponseWriter(), context.Request(),
      r.Path, http.StatusFound)
  case *NotFoundResponse:
    r := response.(*NotFoundResponse)
    context.ResponseWriter().WriteHeader(http.StatusNotFound)
    h.execTemplate(r.Layout, context.ResponseWriter(), context)
  }
}

// Executes a template by name to be written to the writer with the context
// supplied.  This is mainly a convenience method.
func (h *Handler) execTemplate(name string, writer io.Writer, context Context) {
  if h.templateStore == nil {
    panic("No template store associated with handler")
  }

  err := h.templateStore.ExecuteTemplate(writer, name, context)
  if err != nil {
    panic("Template " + name + " could not be executed")
  }
}
