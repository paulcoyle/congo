package congo

import (
  "html/template"
  "log"
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

// TODO: doc
func (h *Handler) applyActions(context Context) (Context, interface{}) {
  var response interface{} = nil

  h.actionMutex.RLock()
  defer h.actionMutex.RUnlock()

  for _, action := range h.actions {
    newContext, newResponse := action(context)
    // If the return value is non-nil, check that it still implements the
    // HandlerAction interface.  If not, the method will panic.
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

// TODO: doc
func Handle(h *Handler) func(http.ResponseWriter, *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    var baseContext Context = NewBaseContext(r)

    //newCtx, response := h.applyActions(ctx)
    context, response := h.applyActions(baseContext)

    // We must have a response in order to do anything.
    if response == nil {
      panic("No response given after handler action chain executed")
    }

    // Perform any rendering needed for a given response.  This method is only
    // concerned with any template rendering or related activity.  Any
    // response types that do not deal with these activities are ignored here.
    responseRenderStep(context, response)

    // final response step
    // TODO: doc
    responseFinalizeStep(context, response)
  }
}

// TODO: doc
func responseRenderStep(context Context, response interface{}) {
  switch response.(type) {
  case *RenderResponse:
    actual := response.(*RenderResponse)
    log.Printf("RENDER TEMPLATE: %s", actual.Template)
  }
}

// TODO: doc
func responseFinalizeStep(context Context, response interface{}) {
  switch response.(type) {
  default:
    panic("Unknown response type")
  case *NullResponse:
  case *RenderResponse:
    actual := response.(*RenderResponse)
    log.Printf("RENDER TEMPLATE: %s", actual.Template)
  }
}
