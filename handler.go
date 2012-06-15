package congo

import (
  "net/http"
  "html/template"
)

// HandlerAction accept a Context interface and optionally return that
// same Context back or another implementation of Context paired with
// any of the responses available in responses.go as the second return value
// (optionally nil).  HandlerAction registered as actions in a Handler that
// return a response cause the chain of actions to halt for that particular
// Handler when it is invoked.
type HandlerAction (func(Context) (Context, interface{}))

type Handler struct {
  templateStore *template.Template
  actions       []HandlerAction
}

func NewHandler() *Handler {
  return &Handler{
    defaultTemplateStore,
    make([]HandlerAction, 0),
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

// Appends a HandlerAction to the action chain for this Handler.  See the type
// description of HandlerAction for details on it.
func (h *Handler) Action(ep HandlerAction) *Handler {
  h.actions = append(h.actions, ep)
  return h
}

// TODO: doc
func (h *Handler) applyActions(context Context) (bool, Context, interface{}) {
  var response interface{} = nil

  for _, action := range h.actions {
    newContext, newResponse := action(context)
    // If the return value is non-nil, check that it still implements the
    // HandlerAction interface.  If not, return false with a nil resulting
    // context and nil response indicating that there was a failure.
    // Otherwise, keep going.
    if newContext != nil {
      if assertContext, ok := newContext.(Context); ok {
        context = assertContext
      } else {
        return false, nil, nil
      }
    }

    // If an HandlerAction has returned a response then halt the chain.
    if newResponse != nil {
      response = newResponse
      break
    }
  }

  return true, context, response
}

// TODO: doc
func Handle(h *Handler) func(http.ResponseWriter, *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    var ctx Context = NewBaseContext()

    // "apply" the actions (inside applyActions, the before actions are
    // applied in order with the "action" applied last). Each of these can
    // modify the context or return a response (those defined in responses.go)
    //ok, newCtx, response := h.applyActions(ctx)
    _, newCtx, _ := h.applyActions(ctx)
    // TODO: check for ok (false means failure in applying HandlerAction chain)
    // TODO: check for nil response (there should be one at this point)

    // The resulting context must implement Context.
    if _, ok := newCtx.(Context); ok {
      ctx = newCtx
    } else {
      panic("Resulting context does not implement congo.Context")
    }

    // act on action response for any template rendering, etc.

    // act on response for final actions (write to w, redirect, etc)
  }
}