package app

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func (a *app) trapKeyEvent(_ *gtk.ApplicationWindow, e *gdk.Event) {
	s := gdk.EventKeyNewFromEvent(e)
	if c := (s.State() & gdk.CONTROL_MASK); c == 0 {
		return
	}

	k := s.KeyVal()
	switch k {
	case gdk.KEY_s:
		a.saveToOutputFile()
		return

	case gdk.KEY_o:
		a.openOutputFile()
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
		a.historyUndo()
		return

	case gdk.KEY_y:
		fallthrough
	case gdk.KEY_Z:
		a.historyDo()
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

	case gdk.KEY_T:
		fallthrough
	case gdk.KEY_t:
		a.addTable()
		return
	}
}
