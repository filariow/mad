package md

import (
	"github.com/filariow/mad/static"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var (
	htmlDocStart = []byte("<html><head/><body>")
	htmlDocEnd   = []byte("</body></html>")
)

func formatHTML(d []byte) []byte {
	size := len(htmlDocStart) + len(d) + len(htmlDocEnd)
	h := make([]byte, 0, size)

	h = append(h, htmlDocStart...)
	h = append(h, d...)
	h = append(h, htmlDocEnd...)
	return h
}

func MarkdownToHTML(d []byte) []byte {
	r := html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags | html.HrefTargetBlank})
	p := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs)
	c := markdown.ToHTML(d, p, r)
	c = formatHTML(c)
	return c
}

func MarkdownToHTMLOrFrontPage(d []byte) []byte {
	if len(d) == 0 {
		return static.Front
	}
	return MarkdownToHTML(d)
}
