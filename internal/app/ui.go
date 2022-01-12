package app

import (
	"log"
	"unsafe"

	"github.com/filariow/mad/views"
	"github.com/gotk3/gotk3/gtk"
	"github.com/webview/webview"
)

func (a *app) loadUI() {
	// builder
	b, err := gtk.BuilderNewFromString(views.Mad)
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

	a.notificationLabel = nl

	onr, err := b.GetObject("notification_revealer")
	if err != nil {
		log.Fatal(err)
	}

	nr, ok := onr.(*gtk.Revealer)
	if !ok {
		log.Fatal("gtk object with id 'notification_revealer' is not a revealer")
	}

	a.notificationRevealer = nr

	// webview
	owb, err := b.GetObject("webview_box")
	if err != nil {
		log.Fatal(err)
	}

	wb, ok := owb.(*gtk.Box)
	if !ok {
		log.Fatal("gtk object with id 'webview_box' is not a Box")
	}
	a.webviewBox = wb

	owr, err := b.GetObject("webview_revealer")
	if err != nil {
		log.Fatal(err)
	}

	wr, ok := owr.(*gtk.Revealer)
	if !ok {
		log.Fatal("gtk object with id 'webview_revealer' is not a Revealer")
	}
	a.webviewRevealer = wr

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

	a.webviewContainer = f
	a.webview = webview.NewWindow(false, p)

	// horizontal box
	hbo, err := b.GetObject("horizontal_box")
	if err != nil {
		log.Fatal(err)
	}

	hb, ok := hbo.(*gtk.Box)
	if !ok {
		log.Fatal("gtk object with id 'horizontal_box' is not a Box")
	}
	a.horizontal_box = hb

	// editor
	eco, err := b.GetObject("editor_container")
	if err != nil {
		log.Fatal(err)
	}

	ec, ok := eco.(*gtk.Viewport)
	if !ok {
		log.Fatalf("gtk object with id 'editor_container' is not a ScrolledWindow")
	}
	a.editorContainer = ec

	eo, err := b.GetObject("editor")
	if err != nil {
		log.Fatal(err)
	}

	e, ok := eo.(*gtk.TextView)
	if !ok {
		log.Fatal("gtk object with id 'editor' is not a TextView")
	}
	a.editor = e

	// textbuffer
	tbo, err := b.GetObject("textbuffer_editor")
	if err != nil {
		log.Fatal(err)
	}

	tb, ok := tbo.(*gtk.TextBuffer)
	if !ok {
		log.Fatal("gtk object with id 'textbuffer' is not a TextBuffer")
	}
	a.textBuffer = tb

	// Window
	ow, err := b.GetObject("main_window")
	if err != nil {
		log.Fatal(err)
	}

	w, ok := ow.(*gtk.ApplicationWindow)
	if !ok {
		log.Fatal("gtk object with id 'main_window' is not a window")
	}

	a.mainWindow = w

	do, err := b.GetObject("dialog_table_size")
	if err != nil {
		log.Fatalf("can not get object with id 'dialog_table_size': %s", err)
	}

	d, ok := do.(*gtk.Dialog)
	if !ok {
		log.Fatal("object with id 'dialog_table_size' is not a gtk.Dialog")
	}
	a.tableDialog = d
}
