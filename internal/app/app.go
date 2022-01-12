package app

import (
	_ "embed"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/filariow/mad/internal/history"
	"github.com/filariow/mad/static"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/webview/webview"
)

const appID = "com.filariow.mad"

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
	a, err := gtk.ApplicationNew(appID, glib.APPLICATION_NON_UNIQUE)
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

	if err := a.activateStyleProvider(); err != nil {
		log.Fatalf("can not activate style provider: %s", err)
	}
	a.activateEditor()

	a.setFont(12)

	a.mainWindow.ToWidget().AddEvents(int(gdk.KEY_PRESS_MASK))
	a.mainWindow.Connect("key_press_event", a.trapKeyEvent)
	a.mainWindow.Connect("destroy", a.destroy)

	a.mainWindow.Show()
	a.a.AddWindow(a.mainWindow)

	go func() {
		time.Sleep(500 * time.Millisecond)
		a.webviewRevealer.SetRevealChild(true)
	}()
}
