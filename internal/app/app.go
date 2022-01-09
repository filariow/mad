package app

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
	"unsafe"

	"github.com/filariow/mad/internal/history"
	"github.com/filariow/mad/static"
	"github.com/filariow/mad/views"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/webview/webview"
)

const appID = "com.filariow.mad"

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
	textBuffer           *gtk.TextBuffer

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

		var c []byte
		if te != "" {
			r := html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags | html.HrefTargetBlank})
			p := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs)
			c = markdown.ToHTML([]byte(te), p, r)
		} else {
			c = static.Front
		}

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
	a.mainWindow.Connect("key_press_event", func(_ *gtk.ApplicationWindow, e *gdk.Event) {
		s := gdk.EventKeyNewFromEvent(e)

		c := (s.State() & gdk.CONTROL_MASK)
		if c == gdk.CONTROL_MASK {
			k := s.KeyVal()
			switch k {
			case gdk.KEY_s:
				s, e := a.textBuffer.GetBounds()
				d, err := a.textBuffer.GetText(s, e, false)
				if err != nil {
					a.notify("can not read text from buffer: %s", err)
					return
				}

				if a.outputFile == "" {
					a.setOutputFile("Save")
					if a.outputFile == "" {
						return
					}
				}

				if err := ioutil.WriteFile(a.outputFile, []byte(d), 0777); err != nil {
					a.notify("can not save output file '%s': %s", a.outputFile, err)
					return
				}

				a.notify("saved into '%s'", a.outputFile)
				return

			case gdk.KEY_o:
				a.setOutputFile("Open")
				if a.outputFile == "" {
					return
				}
				s, err := ioutil.ReadFile(a.outputFile)
				if err != nil {
					a.notify(err.Error())
					return
				}

				a.textBuffer.SetText(string(s))
				a.notify("opened file: %s", a.outputFile)
				return

			case gdk.KEY_p:
				a.toogleView()
				return

			case gdk.KEY_q:
				if !a.closeRequested {
					a.closeRequested = true
					a.mainWindow.Destroy()
				}
				return
			case gdk.KEY_plus:
				fallthrough
			case gdk.KEY_equal:
				a.applyDeltaToFont(1)
				return
			case gdk.KEY_minus:
				a.applyDeltaToFont(-1)
				return
			case gdk.KEY_z:
				fallthrough
			case gdk.KEY_Y:
				a.historyApply(func(t string) *string {
					return a.history.Undo(t)
				})
				return
			case gdk.KEY_y:
				fallthrough
			case gdk.KEY_Z:
				a.historyApply(func(t string) *string {
					return a.history.Do(t)
				})
				return
			case gdk.KEY_i:
				fallthrough
			case gdk.KEY_I:
				a.italic()
				return
			case gdk.KEY_b:
				fallthrough
			case gdk.KEY_B:
				a.bolderize()
				return
			}
		}
	})

	a.mainWindow.Show()
	a.a.AddWindow(a.mainWindow)

	go func() {
		time.Sleep(500 * time.Millisecond)
		a.webviewRevealer.SetRevealChild(true)
	}()
}

func (a *app) bolderize() {
	ba := "**"
	s, e, ok := a.textBuffer.GetSelectionBounds()
	if ok {
		a.wrapSelection(s, e, ba, ba)
		return
	}
	a.wrapAtCursor(ba, ba)
}

func (a *app) italic() {
	ia := "*"
	s, e, ok := a.textBuffer.GetSelectionBounds()
	if ok {
		a.wrapSelection(s, e, ia, ia)
		return
	}
	a.wrapAtCursor(ia, ia)
}

func (a *app) wrapSelection(s, e *gtk.TextIter, left, right string) {
	os := s.GetOffset()
	osce := a.textBuffer.GetIterAtOffset(os + len(left))
	st, err := a.textBuffer.GetText(s, osce, false)
	if err != nil {
		log.Printf("can not read first two runes from selection: %s", err)
		return
	}

	if st == left {
		oe := e.GetOffset()
		oecs := a.textBuffer.GetIterAtOffset(oe - len(right))
		et, err := a.textBuffer.GetText(oecs, e, false)
		if err != nil {
			log.Printf("can not read last two runes from selection: %s", err)
			return
		}

		if et == right {
			mss := a.textBuffer.CreateMark("wrapping-start-start", s, false)
			defer a.textBuffer.DeleteMark(mss)
			mse := a.textBuffer.CreateMark("wrapping-start-end", osce, false)
			defer a.textBuffer.DeleteMark(mse)
			mes := a.textBuffer.CreateMark("wrapping-end-start", oecs, false)
			defer a.textBuffer.DeleteMark(mes)
			mee := a.textBuffer.CreateMark("wrapping-end-end", e, false)
			defer a.textBuffer.DeleteMark(mee)
			a.textBuffer.Delete(a.textBuffer.GetIterAtMark(mss), a.textBuffer.GetIterAtMark(mse))
			a.textBuffer.Delete(a.textBuffer.GetIterAtMark(mes), a.textBuffer.GetIterAtMark(mee))
			return
		}
	}

	ms := a.textBuffer.CreateMark("wrapping-start", s, false)
	defer a.textBuffer.DeleteMark(ms)
	me := a.textBuffer.CreateMark("wrapping-end", e, false)
	defer a.textBuffer.DeleteMark(me)
	ims := a.textBuffer.GetIterAtMark(ms)
	a.textBuffer.Insert(ims, left)
	ime := a.textBuffer.GetIterAtMark(me)
	a.textBuffer.Insert(ime, right)

	ims = a.textBuffer.GetIterAtMark(ms)
	nims := a.textBuffer.GetIterAtOffset(ims.GetOffset() - len(left))
	a.textBuffer.SelectRange(nims, ime)
}

func (a *app) wrapAtCursor(left, right string) {
	t := left + right
	i := a.textBuffer.GetIterAtMark(a.textBuffer.GetInsert()).GetOffset()
	te := a.textBuffer.GetEndIter().GetOffset()
	ll, lr := len(left), len(right)
	if i >= ll && i+lr <= te {
		s := a.textBuffer.GetIterAtOffset(i - ll)
		e := a.textBuffer.GetIterAtOffset(i + lr)

		lt, err := a.textBuffer.GetText(s, e, false)
		if err != nil {
			log.Printf("can not retrieve text near selection")
			return
		}

		if lt == t {
			a.textBuffer.Delete(s, e)
			return
		}
	}
	a.textBuffer.InsertAtCursor(t)
	ic := a.textBuffer.GetIterAtOffset(i + ll)
	a.textBuffer.PlaceCursor(ic)

}

func (a *app) GetTextMarkdown() (string, error) {
	s, e := a.textBuffer.GetBounds()
	t, err := a.textBuffer.GetText(s, e, false)
	return t, err
}

func (a *app) historyApply(f func(string) *string) {
	a.trapTextChanges = false
	defer func() {
		a.trapTextChanges = true
	}()

	s, e := a.textBuffer.GetBounds()
	t, err := a.textBuffer.GetText(s, e, false)
	if err != nil {
		log.Printf("can not retrieve textbuffer text: %s", err)
		return
	}

	if tu := f(t); tu != nil {
		a.textBuffer.SetText(*tu)
	}
}

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
	a.webview = webview.NewWindow(true, p)

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
