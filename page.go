package congo

type PageContext interface {
  Content() string
  Write([]byte) (int, error)
}

type Page struct {
  content string
}

func NewPage() *Page {
  return &Page{}
}

func (p *Page) Content() string {
  return p.content
}

func (p *Page) Write(data []byte) (int, error) {
  p.content = p.content + string(data)
  return len(data), nil
}
