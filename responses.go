package congo

// The NullResponse allows a handler to process normally and not complain
// about a missing or unknown response.  It is really only useful in testing.
type NullResponse struct {
}

type RenderResponse struct {
  Template string
  Layout   string
}

type RedirectResponse struct {
  Path string
}

type NotFoundResponse struct {
  Template string
  Layout   string
}
