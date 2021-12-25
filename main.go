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
	a                    *gtk.Application
	mainWindow           *gtk.ApplicationWindow
	editorContainer      *gtk.Viewport
	editor               *gtk.TextView
	webviewBox           *gtk.Box
	webviewRevealer      *gtk.Revealer
	webviewContainer     *gtk.Viewport
	webview              webview.WebView
	notificationLabel    *gtk.Label
	notificationRevealer *gtk.Revealer
	horizontal_box       *gtk.Box
	cssProvider          *gtk.CssProvider

	f                *os.File
	outputFile       string
	tempFolder       string
	fullScreenEditor bool
	closeRequested   bool
	fontSize         int
}

func main() {
	os.Exit(run())
}

func run() int {
	// Initialize GTK without parsing any command line arguments.
	gtk.Init(nil)

	a := newApp()
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

	l := app{
		a:                a,
		f:                tf,
		tempFolder:       td,
		fullScreenEditor: true,
		fontSize:         12,
	}

	a.Connect("activate", l.activate)
	return &l
}

func (t *app) destroy() {
	if t.webview != nil {
		t.webview.Destroy()
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
	owb, err := b.GetObject("webview_box")
	if err != nil {
		log.Fatal(err)
	}

	wb, ok := owb.(*gtk.Box)
	if !ok {
		log.Fatal("gtk object with id 'webview_box' is not a Box")
	}
	t.webviewBox = wb

	owr, err := b.GetObject("webview_revealer")
	if err != nil {
		log.Fatal(err)
	}

	wr, ok := owr.(*gtk.Revealer)
	if !ok {
		log.Fatal("gtk object with id 'webview_revealer' is not a Revealer")
	}
	t.webviewRevealer = wr

	owv, err := b.GetObject("webview_container")
	if err != nil {
		log.Fatal(err)
	}

	f, ok := owv.(*gtk.Viewport)
	if !ok {
		log.Fatal("gtk object with id 'webview_container' is not a ScrolledWindow")
	}

	fw := f.ToWidget().GObject
	p := unsafe.Pointer(fw)

	t.webviewContainer = f
	t.webview = webview.NewWindow(true, p)

	t.webview.Navigate("file://" + t.f.Name())

	// horizontal box
	hbo, err := b.GetObject("horizontal_box")
	if err != nil {
		log.Fatal(err)
	}

	hb, ok := hbo.(*gtk.Box)
	if !ok {
		log.Fatal("gtk object with id 'horizontal_box' is not a Box")
	}
	t.horizontal_box = hb

	// editor
	eco, err := b.GetObject("editor_container")
	if err != nil {
		log.Fatal(err)
	}

	ec, ok := eco.(*gtk.Viewport)
	if !ok {
		log.Fatalf("gtk objecgt with id 'editor_container' is not a ScrolledWindow")
	}
	t.editorContainer = ec

	eo, err := b.GetObject("editor")
	if err != nil {
		log.Fatal(err)
	}

	e, ok := eo.(*gtk.TextView)
	if !ok {
		log.Fatal("gtk object with id 'editor' is not a TextView")
	}
	t.editor = e
	e.GrabFocus()

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

		var c []byte
		if te != "" {
			r := html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags | html.HrefTargetBlank})
			p := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs)
			c = markdown.ToHTML([]byte(te), p, r)
		} else {
			c = front
		}

		if err := ioutil.WriteFile(t.f.Name(), c, 0777); err != nil {
			log.Printf("can not update content in temp file: %s", err)
			return
		}
		t.webview.Navigate("file://" + t.f.Name())
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

	t.mainWindow = w
	cssp, err := gtk.CssProviderNew()
	if err != nil {
		t.notify(err.Error())
		return
	}
	t.cssProvider = cssp

	s, err := t.mainWindow.ToWidget().GetScreen()
	if err != nil {
		log.Fatalf("can not retrieve the Screen for main window: %s", err)
	}
	gtk.AddProviderForScreen(s, cssp, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
	t.setFont(12)

	w.Connect("destroy", t.destroy)

	w.ToWidget().AddEvents(int(gdk.KEY_PRESS_MASK))
	w.Connect("key_press_event", func(_ *gtk.ApplicationWindow, e *gdk.Event) {
		s := gdk.EventKeyNewFromEvent(e)

		c := (s.State() & gdk.CONTROL_MASK)
		if c == gdk.CONTROL_MASK {
			k := s.KeyVal()
			switch k {
			case gdk.KEY_s:
				s, e := tb.GetBounds()
				d, err := tb.GetText(s, e, false)
				if err != nil {
					t.notify("can not read text from buffer: %s", err)
					return
				}

				if t.outputFile == "" {
					t.setOutputFile("Save")
					if t.outputFile == "" {
						return
					}
				}

				if err := ioutil.WriteFile(t.outputFile, []byte(d), 0777); err != nil {
					t.notify("can not save output file '%s': %s", t.outputFile, err)
					return
				}

				t.notify("saved into '%s'", t.outputFile)
				return

			case gdk.KEY_o:
				t.setOutputFile("Open")
				if t.outputFile == "" {
					return
				}
				s, err := ioutil.ReadFile(t.outputFile)
				if err != nil {
					t.notify(err.Error())
					return
				}

				tb.SetText(string(s))
				t.notify("opened file: %s", t.outputFile)
				return

			case gdk.KEY_p:
				t.toogleView()
				return

			case gdk.KEY_q:
				if !t.closeRequested {
					t.closeRequested = true
					t.mainWindow.Destroy()
				}
				return
			case gdk.KEY_plus:
				fallthrough
			case gdk.KEY_equal:
				t.applyDeltaToFont(1)
				return
			case gdk.KEY_minus:
				t.applyDeltaToFont(-1)
				return
			}
		}
	})

	w.Show()
	t.a.AddWindow(w)

	go func() {
		time.Sleep(500 * time.Millisecond)
		t.webviewRevealer.SetRevealChild(true)
	}()
}

func (a *app) applyDeltaToFont(delta int) {
	s := a.fontSize + delta
	if err := a.setFont(s); err != nil {
		a.notify(err.Error())
	}
}

func (a *app) setFont(size int) error {
	if size <= 0 {
		return fmt.Errorf("font size must be positive, invalid value %d", size)
	}

	a.fontSize = size
	css := fmt.Sprintf("textview { font-size: %dpt; }", size)
	if err := a.cssProvider.LoadFromData(css); err != nil {
		return err
	}

	return nil
}

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

func (a *app) setOutputFile(action string) {
	f, err := gtk.FileChooserDialogNewWith2Buttons("Output File",
		a.a.GetActiveWindow(),
		gtk.FILE_CHOOSER_ACTION_SAVE,
		"Open",
		gtk.RESPONSE_ACCEPT,
		"Cancel",
		gtk.RESPONSE_CANCEL)
	if err != nil {
		log.Println(err)
	}

	if r := f.Run(); r == gtk.RESPONSE_ACCEPT {
		a.outputFile = f.GetFilename()
	}

	f.Destroy()
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
