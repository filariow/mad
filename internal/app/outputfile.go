package app

import (
	"io/ioutil"
	"log"

	"github.com/gotk3/gotk3/gtk"
)

func (a *app) saveToOutputFile() {
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

}

func (a *app) openOutputFile() {
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
