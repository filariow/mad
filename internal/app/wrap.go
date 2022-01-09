package app

import (
	"log"

	"github.com/gotk3/gotk3/gtk"
)

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
