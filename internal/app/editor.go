package app

import (
	"github.com/filariow/mad/internal/history"
	"github.com/gotk3/gotk3/gtk"
)

func (a *app) activateEditor() {
	a.textBuffer.Connect("changed", a.updatePreview)

	a.textBuffer.Connect("insert-text", a.trapInsertText)
	a.textBuffer.Connect("delete-range", a.trapDeleteRange)
}

func (a *app) trapInsertText(_ *gtk.TextBuffer, location *gtk.TextIter, text string, length int) {
	if !a.trapTextChanges {
		return
	}

	r := history.NewInsertRecord(text, location.GetOffset(), length)
	a.history.Add(r)
}

func (a *app) trapDeleteRange(_ *gtk.TextBuffer, start, end *gtk.TextIter) {
	if !a.trapTextChanges {
		return
	}

	t := start.GetSlice(end)
	r := history.NewDeleteRecord(t, start.GetOffset(), end.GetOffset())
	a.history.Add(r)
}

func (a *app) GetTextMarkdown() (string, error) {
	s, e := a.textBuffer.GetBounds()
	t, err := a.textBuffer.GetText(s, e, false)
	return t, err
}
