package app

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
)

//go:inline
func (a *app) activateStyleProvider() error {
	cssp, err := gtk.CssProviderNew()
	if err != nil {
		return err
	}
	a.cssProvider = cssp

	s, err := a.mainWindow.ToWidget().GetScreen()
	if err != nil {
		return fmt.Errorf("can not retrieve the Screen for main window: %w", err)
	}
	gtk.AddProviderForScreen(s, cssp, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	return nil
}
