package app

import (
	_ "embed"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/filariow/mad/internal/history"
	"github.com/filariow/mad/internal/md"
	"github.com/filariow/mad/static"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/webview/webview"
)

const (
	appID = "com.filariow.mad"
)

type app struct {
	a                    *gtk.Application
	builder              *gtk.Builder
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
	textBuffer           *gtk.TextBuffer
	tableDialog          *gtk.Dialog

	history          history.History
	trapTextChanges  bool
	f                *os.File
	outputFile       string
	tempFolder       string
	fullScreenEditor bool
	closeRequested   bool
	fontSize         int
}

func New() *app {
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

	if err = ioutil.WriteFile(tf.Name(), static.Front, 0777); err != nil {
		log.Fatal(err)
	}

	l := app{
		a:                a,
		f:                tf,
		tempFolder:       td,
		fullScreenEditor: true,
		fontSize:         12,
		history:          history.NewUnbounded(),
		trapTextChanges:  true,
	}

	a.Connect("activate", l.activate)
	return &l
}

func (a *app) destroy() {
	if a.webview != nil {
		a.webview.Destroy()
	}

	err := os.RemoveAll(a.tempFolder)
	if err != nil {
		log.Println(err)
	}
}

func (a *app) Run(args []string) int {
	return a.a.Run(args)
}

func (a *app) activate() {
	a.loadUI()

	a.webview.Navigate("file://" + a.f.Name())
	a.editor.GrabFocus()

	a.textBuffer.Connect("changed", func() {
		s, e := a.textBuffer.GetBounds()
		te, err := a.textBuffer.GetText(s, e, false)
		if err != nil {
			log.Printf("can not retrieve text from textbuffer: %s", err)
			return
		}

		c := md.MarkdownToHTMLOrFrontPage([]byte(te))
		if err := ioutil.WriteFile(a.f.Name(), c, 0777); err != nil {
			log.Printf("can not update content in temp file: %s", err)
			return
		}
		a.webview.Navigate("file://" + a.f.Name())
	})

	a.textBuffer.Connect("insert-text", func(_ *gtk.TextBuffer, location *gtk.TextIter, text string, length int) {
		if !a.trapTextChanges {
			return
		}
		r := history.NewInsertRecord(text, location.GetOffset(), length)
		a.history.Add(r)
	})

	a.textBuffer.Connect("delete-range", func(_ *gtk.TextBuffer, start, end *gtk.TextIter) {
		if !a.trapTextChanges {
			return
		}
		t := start.GetSlice(end)
		r := history.NewDeleteRecord(t, start.GetOffset(), end.GetOffset())
		a.history.Add(r)
	})

	cssp, err := gtk.CssProviderNew()
	if err != nil {
		a.notify(err.Error())
		return
	}
	a.cssProvider = cssp

	s, err := a.mainWindow.ToWidget().GetScreen()
	if err != nil {
		log.Fatalf("can not retrieve the Screen for main window: %s", err)
	}
	gtk.AddProviderForScreen(s, cssp, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
	a.setFont(12)

	a.mainWindow.Connect("destroy", a.destroy)

	a.mainWindow.ToWidget().AddEvents(int(gdk.KEY_PRESS_MASK))
	a.mainWindow.Connect("key_press_event", a.trapKeyEvent)

	a.mainWindow.Show()
	a.a.AddWindow(a.mainWindow)

	go func() {
		time.Sleep(500 * time.Millisecond)
		a.webviewRevealer.SetRevealChild(true)
	}()
}

func (a *app) GetTextMarkdown() (string, error) {
	s, e := a.textBuffer.GetBounds()
	t, err := a.textBuffer.GetText(s, e, false)
	return t, err
}
