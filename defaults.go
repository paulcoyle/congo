package congo

import (
  "html/template"
)

// Holds a default Template for use by handlers when they are created.
var defaultTemplateStore *template.Template

func DefaultTemplateStore() *template.Template {
  return defaultTemplateStore
}

func SetDefaultTemplateStore(store *template.Template) {
  defaultTemplateStore = store
}
