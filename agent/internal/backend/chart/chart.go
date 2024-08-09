package chart

import (
	"bytes"
	"fmt"
	"html/template"
	"io"

	"github.com/go-echarts/go-echarts/v2/render"
)

type Renderer interface {
	Render(io.Writer) error
}

func renderToHtml(c render.Renderer) template.HTML {
	var buf bytes.Buffer
	_ = c.Render(&buf)
	return template.HTML(buf.String())
}

type chartRenderer struct {
	c             interface{}
	before        []func()
	chartName     string
	chartTemplate string
}

func newchartRenderer(c interface{}, chartName string, chartTemplate string, before ...func()) render.Renderer {
	return &chartRenderer{
		c:             c,
		chartName:     chartName,
		chartTemplate: chartTemplate,
		before:        before}
}

func (r *chartRenderer) Render(w io.Writer) error {
	for _, fn := range r.before {
		fn()
	}

	tpl := template.
		Must(template.New(r.chartName).
			Funcs(template.FuncMap{
				"safeJS": func(s interface{}) template.JS {
					return template.JS(fmt.Sprint(s))
				},
			}).
			Parse(r.chartTemplate),
		)

	err := tpl.ExecuteTemplate(w, r.chartName, r.c)
	return err
}

// To satisfy the interface
func (r *chartRenderer) RenderContent() []byte {
	return []byte{}
	// To satisfy the interface
}
func (r *chartRenderer) RenderSnippet() render.ChartSnippet {
	return render.ChartSnippet{}
}
