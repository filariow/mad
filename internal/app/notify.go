package app

import (
	"fmt"
	"time"
)

func (a *app) notify(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	a.notificationLabel.SetText(msg)
	a.notificationRevealer.SetRevealChild(true)
	go func() {
		time.Sleep(3 * time.Second)
		a.notificationRevealer.SetRevealChild(false)
	}()
}
