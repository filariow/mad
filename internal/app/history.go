package app

import "log"

func (a *app) historyDo() {
	a.historyApply(func(t string) *string {
		return a.history.Do(t)
	})
}

func (a *app) historyUndo() {
	a.historyApply(func(t string) *string {
		return a.history.Undo(t)
	})
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
