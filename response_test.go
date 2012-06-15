package congo

import (
  "html/template"
  "net/http"
  "net/url"
  "testing"
)

func createMockRequest() *http.Request {
  return &http.Request{
    Method: "GET",
    URL:    &url.URL{Path: "/"},
  }
}

func createNotFoundTemplates() *template.Template {
  root, _ := template.New("layout").Parse("LAYOUT: {{.Content}}")
  root.New("inner").Parse("NOT FOUND")
  return root
}

func ActionNotFound(c Context) (Context, interface{}) {
  return c, &NotFoundResponse{"inner", "layout"}
}

func ActionRedirect(c Context) (Context, interface{}) {
  return c, &RedirectResponse{"/derp"}
}

func TestNotFoundResponseSetsProperStatusCode(t *testing.T) {
  responseWriter := &MockResponseWriter{}
  request := createMockRequest()
  tpls := createNotFoundTemplates()

  handler := NewHandler().SetTemplateStore(tpls)
  handler.Actions(ActionNotFound)
  handlerFn := MuxHandler(handler)
  handlerFn(responseWriter, request)

  if responseWriter.StatusCode != http.StatusNotFound {
    t.Fail()
  }
}

func TestRedirectResponseSetsProperStatusCode(t *testing.T) {
  responseWriter := &MockResponseWriter{}
  request := createMockRequest()
  tpls := createNotFoundTemplates()

  handler := NewHandler().SetTemplateStore(tpls)
  handler.Actions(ActionRedirect)
  handlerFn := MuxHandler(handler)
  handlerFn(responseWriter, request)

  if responseWriter.StatusCode != http.StatusFound {
    t.Fail()
  }
}
