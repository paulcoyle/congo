package congo

import (
  "html/template"
  "net/http"
  "testing"
)

type AugmentedContext struct {
  Context
}

func (ac *AugmentedContext) Augmented() string {
  return "DAMN RIGHT!"
}

type MockResponseWriter struct {
  StatusCode int
  Content    string
}

func (mw *MockResponseWriter) Header() http.Header {
  return make(map[string][]string)
}

func (mw *MockResponseWriter) Write(b []byte) (int, error) {
  if mw.StatusCode == 0 {
    mw.WriteHeader(http.StatusOK)
  }
  mw.Content = mw.Content + string(b)
  return len(b), nil
}

func (mw *MockResponseWriter) WriteHeader(statusCode int) {
  mw.StatusCode = statusCode
}

func createBasicTemplates() *template.Template {
  root, _ := template.New("layout").Parse("LAYOUT: {{.Content}}")
  root.New("inner").Parse("INNER: HAI!")
  return root
}

func ActionRenderTemplate(c Context) (Context, interface{}) {
  return c, &RenderResponse{"inner", "layout"}
}

func createContextUsingTemplates() *template.Template {
  root, _ := template.New("layout").Parse("LAYOUT: {{.Content}}")
  root.New("inner").Parse("INNER: HAI! {{.Augmented}}")
  return root
}

func ActionAugmentContextRenderTemplate(c Context) (Context, interface{}) {
  aug := &AugmentedContext{c}
  return aug, &RenderResponse{"inner", "layout"}
}

func TestTemplateAndLayoutRender(t *testing.T) {
  responseWriter := &MockResponseWriter{}
  tpls := createBasicTemplates()

  handler := NewHandler().SetTemplateStore(tpls)
  handler.Actions(ActionRenderTemplate)
  handlerFn := MuxHandler(handler)
  handlerFn(responseWriter, nil)

  expected := "LAYOUT: INNER: HAI!"
  if responseWriter.Content != expected {
    t.Fail()
  }
}

func TestTemplateAndLayoutRenderWithAugmentedContext(t *testing.T) {
  responseWriter := &MockResponseWriter{}
  tpls := createContextUsingTemplates()

  handler := NewHandler().SetTemplateStore(tpls)
  handler.Actions(ActionAugmentContextRenderTemplate)
  handlerFn := MuxHandler(handler)
  handlerFn(responseWriter, nil)

  expected := "LAYOUT: INNER: HAI! DAMN RIGHT!"
  if responseWriter.Content != expected {
    t.Fail()
  }
}
