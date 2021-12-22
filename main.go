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

func main() {
	os.Exit(run())
}

func run() int {
	// Initialize GTK without parsing any command line arguments.
	gtk.Init(nil)

	a, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal(err)
	}

	a.Connect("startup", func() {
		log.Println("startup")
	})
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
	defer func() {
		err := os.RemoveAll(td)
		if err != nil {
			log.Println(err)
		}
	}()

	a.Connect("activate", activate(a, tf))
	return a.Run(os.Args)
}

func activate(a *gtk.Application, sf *os.File) func() {
	// TODO: refactor this method: move init in main, allow defers
	return func() {
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
		// defer wv.Destroy()

		wv.Navigate("file://" + sf.Name())

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
			t, err := tb.GetText(s, e, false)
			if err != nil {
				log.Printf("can not retrieve text from textbuffer: %s", err)
				return
			}

			r := html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags | html.HrefTargetBlank})
			p := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs)
			h := markdown.ToHTML([]byte(t), p, r)

			if err := ioutil.WriteFile(sf.Name(), h, 0777); err != nil {
				log.Printf("can not update content in temp file: %s", err)
				return
			}
			wv.Navigate("file://" + sf.Name())
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
		a.AddWindow(w)
	}
}
