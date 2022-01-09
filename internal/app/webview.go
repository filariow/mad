package app

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/filariow/mad/internal/md"
)

func (a *app) toogleView() {
	r := !a.webviewRevealer.GetChildRevealed()
	if r {
		a.webviewBox.SetVisible(r)
	}
	a.webviewRevealer.SetRevealChild(r)

	if !r {
		go func() {
			d := a.webviewRevealer.GetTransitionDuration()
			time.Sleep(time.Duration(d) * time.Millisecond)
			a.webviewBox.SetVisible(r)
		}()
	}
}

func (a *app) updatePreview() {
	t, err := a.GetTextMarkdown()
	if err != nil {
		msg := fmt.Sprintf("can not update preview: %s", err)
		a.notify(msg)
		log.Println(msg)
		return
	}
	c := md.MarkdownToHTMLOrFrontPage([]byte(t))
	if err := ioutil.WriteFile(a.f.Name(), c, 0777); err != nil {
		log.Printf("can not update content in temp file: %s", err)
		return
	}
	a.webview.Navigate("file://" + a.f.Name())

}
