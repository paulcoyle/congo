package congo

import (
  "log"
  "net/http"
)

// A function that accepts what should be a PageContext interface
// and optionally returns that value back (or another that still
// implements PageContext) and any of the actions available in
// actions.go as the second return value.
type Passthrough (func(PageContext) (PageContext, interface{}))

type Handler struct {
  actions []Passthrough
}

func NewHandler() *Handler {
  return &Handler{make([]Passthrough, 0)}
}

func (h *Handler) Action(ep Passthrough) *Handler {
  h.actions = append(h.actions, ep)
  return h
}

// TODO: more docs
// IMPORTANT: if an action returns a non-nil response, further actions are not
//            run and the function returns.
func (h *Handler) applyActions(context PageContext) (bool, PageContext, interface{}) {
  var response interface{} = nil

  for _, action := range h.actions {
    newContext, newResponse := action(context)
    // If the return value is non-nil, check that it still implements
    // the PageContext interface.  If not, return false with a nil
    // resulting context which indicates that there was a failure.
    // Otherwise, keep going.
    if newContext != nil {
      if assertContext, ok := newContext.(PageContext); ok {
        context = assertContext
      } else {
        return false, nil, nil
      }
    }

    // Is there a response? If so, we're done here.
    if newResponse != nil {
      response = newResponse
      break
    }
  }

  return true, context, response
}

func Handle(h *Handler) func(http.ResponseWriter, *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    // Create the base page context "Page"
    var ctx PageContext = NewPage()

    // "apply" the actions (inside applyActions, the before actions are
    // applied in order with the "action" applied last). Each of these can
    // modify the context or return a response (those defined in responses.go)
    //ok, newCtx, response := h.applyActions(ctx)
    _, newCtx, _ := h.applyActions(ctx)
    // TODO: check for ok, act differently...
    // TODO: check for nil response
    if _, ok := newCtx.(PageContext); ok {
      ctx = newCtx
    } else {
      log.Fatal("Damnit")
    }

    // act on action response for template rendering
    // log.Printf("Context type:  %s\n", reflect.TypeOf(ctx))
    // log.Printf("               %v\n", ctx)
    // log.Printf("Response type: %s\n", reflect.TypeOf(response))
    // log.Printf("               %v\n", response)

    // act on response for final actions (write to w, redirect, etc)
  }
}