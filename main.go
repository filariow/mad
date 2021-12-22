package main

import (
	"io/ioutil"
	"log"
	"os"
	"unsafe"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/webview/webview"
)

const appID = "com.filariow.mad"

type app struct {
	a          *gtk.Application
	w          webview.WebView
	tempFolder string
	f          *os.File
}

func main() {
	os.Exit(run())
}

func run() int {
	// Initialize GTK without parsing any command line arguments.
	gtk.Init(nil)

	a := newApp()
	defer a.destroy()
	return a.run(os.Args)
}

func newApp() *app {
	a, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal(err)
	}

	l := app{a: a}

	a.Connect("startup", func() {
		td, err := ioutil.TempDir("", "mad")
		if err != nil {
			log.Fatal(err)
		}

		tf, err := ioutil.TempFile(td, "mad.md.html")
		if err != nil {
			log.Fatal(err)
		}
		d := []byte("hello world!")
		if err = ioutil.WriteFile(tf.Name(), d, 0777); err != nil {
			log.Fatal(err)
		}
		l.tempFolder = td
		l.f = tf
	})

	a.Connect("activate", l.activate)
	return &l
}

func (t *app) destroy() {
	t.w.Destroy()

	err := os.RemoveAll(t.tempFolder)
	if err != nil {
		log.Println(err)
	}
}

func (t *app) run(args []string) int {
	return t.a.Run(args)
}

func (t *app) activate() {
	log.Println("activate")

	// builder
	b, err := gtk.BuilderNewFromFile("views/mad.glade")
	if err != nil {
		log.Fatal(err)
	}

	// webview
	owv, err := b.GetObject("webview_container")
	if err != nil {
		log.Fatal(err)
	}

	f, ok := owv.(*gtk.ScrolledWindow)
	if !ok {
		log.Fatal("gtk object with id 'webview_container' is not a ScrolledWindow")
	}

	fw := f.ToWidget().GObject
	p := unsafe.Pointer(fw)
	wv := webview.NewWindow(true, p)
	t.w = wv

	wv.Navigate("file://" + t.f.Name())

	// textbuffer
	tbo, err := b.GetObject("textbuffer")
	if err != nil {
		log.Fatal(err)
	}

	tb, ok := tbo.(*gtk.TextBuffer)
	if !ok {
		log.Fatal("gtk object with id 'textbuffer' is not a TextBuffer")
	}

	tb.Connect("changed", func() {
		s, e := tb.GetBounds()
		te, err := tb.GetText(s, e, false)
		if err != nil {
			log.Printf("can not retrieve text from textbuffer: %s", err)
			return
		}

		r := html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags | html.HrefTargetBlank})
		p := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs)
		h := markdown.ToHTML([]byte(te), p, r)

		if err := ioutil.WriteFile(t.f.Name(), h, 0777); err != nil {
			log.Printf("can not update content in temp file: %s", err)
			return
		}
		wv.Navigate("file://" + t.f.Name())
	})

	// Window
	ow, err := b.GetObject("main_window")
	if err != nil {
		log.Fatal(err)
	}

	w, ok := ow.(*gtk.Window)
	if !ok {
		log.Fatal("gtk object with id 'main_window' is not a window")
	}

	w.Show()
	t.a.AddWindow(w)
}
