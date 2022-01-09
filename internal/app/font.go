package app

import "fmt"

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
