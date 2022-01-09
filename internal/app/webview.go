package app

import "time"

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
