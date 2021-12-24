package main

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
	"unsafe"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/webview/webview"
)

const appID = "com.filariow.mad"

//go:embed views/mad2.ui
var ui string

//go:embed front.html
var front []byte

type app struct {
	a          *gtk.Application
	w          webview.WebView
	tempFolder string
	f          *os.File
	outputFile string

	notificationLabel    *gtk.Label
	notificationRevealer *gtk.Revealer
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

	td, err := ioutil.TempDir("", "mad")
	if err != nil {
		log.Fatal(err)
	}

	tf, err := ioutil.TempFile(td, "mad.md.html")
	if err != nil {
		log.Fatal(err)
	}

	if err = ioutil.WriteFile(tf.Name(), front, 0777); err != nil {
		log.Fatal(err)
	}

	of := os.ExpandEnv("${HOME}/mad.md")

	l := app{
		a:          a,
		f:          tf,
		tempFolder: td,
		outputFile: of,
	}

	a.Connect("activate", l.activate)
	return &l
}

func (t *app) destroy() {
	if t.w != nil {
		t.w.Destroy()
	}

	err := os.RemoveAll(t.tempFolder)
	if err != nil {
		log.Println(err)
	}
}

func (t *app) run(args []string) int {
	return t.a.Run(args)
}

func (t *app) activate() {
	// builder
	b, err := gtk.BuilderNewFromString(ui)
	if err != nil {
		log.Fatal(err)
	}

	// notification
	onl, err := b.GetObject("notification_label")
	if err != nil {
		log.Fatal(err)
	}

	nl, ok := onl.(*gtk.Label)
	if !ok {
		log.Fatal("gtk object with id 'notification_label' is not a label")
	}

	t.notificationLabel = nl

	onr, err := b.GetObject("notification_revealer")
	if err != nil {
		log.Fatal(err)
	}

	nr, ok := onr.(*gtk.Revealer)
	if !ok {
		log.Fatal("gtk object with id 'notification_revealer' is not a revealer")
	}

	t.notificationRevealer = nr

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
	tbo, err := b.GetObject("textbuffer_editor")
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

	w, ok := ow.(*gtk.ApplicationWindow)
	if !ok {
		log.Fatal("gtk object with id 'main_window' is not a window")
	}

	w.ToWidget().AddEvents(int(gdk.KEY_PRESS_MASK))
	w.Connect("key_press_event", func(_ *gtk.ApplicationWindow, e *gdk.Event) {
		s := gdk.EventKeyNewFromEvent(e)
		k := s.KeyVal()

		c := (s.State() & gdk.CONTROL_MASK)
		if c == gdk.CONTROL_MASK && k == gdk.KEY_s {

			s, e := tb.GetBounds()
			d, err := tb.GetText(s, e, false)
			if err != nil {
				log.Printf("can not read text from buffer: %s", err)
				return
			}

			if err := ioutil.WriteFile(t.outputFile, []byte(d), 0777); err != nil {
				log.Printf("can not save output file '%s': %s", t.outputFile, err)
			}

			t.notify("saved into '%s'", t.outputFile)
		}
	})

	w.Show()
	t.a.AddWindow(w)
}

func (a *app) notify(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	a.notificationLabel.SetText(msg)
	a.notificationRevealer.SetRevealChild(true)
	go func() {
		time.Sleep(3 * time.Second)
		a.notificationRevealer.SetRevealChild(false)
	}()
}
