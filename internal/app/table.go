package app

import (
	"log"

	"github.com/filariow/mad/views"
	"github.com/gotk3/gotk3/gtk"
)

func (a *app) addTable() {
	// builder
	b, err := gtk.BuilderNewFromString(views.Mad)
	if err != nil {
		log.Fatal(err)
	}

	do, err := b.GetObject("dialog_table_size")
	if err != nil {
		log.Fatalf("can not get object with id 'dialog_table_size': %s", err)
	}

	d, ok := do.(*gtk.Dialog)
	if !ok {
		log.Fatal("object with id 'dialog_table_size' is not a gtk.Dialog")
	}

	// dialog table cols spin button
	csbo, err := b.GetObject("dialog_table_cols_spin")
	if err != nil {
		log.Fatalf("can not get object with id 'dialog_table_cols_spin': %s", err)
	}

	csb, ok := csbo.(*gtk.SpinButton)
	if !ok {
		log.Fatal("object with id 'dialog_table_cols_spin' is not a gtk.SpinButton")
	}

	// dialog table rows spin button
	rsbo, err := b.GetObject("dialog_table_rows_spin")
	if err != nil {
		log.Fatalf("can not get object with id 'dialog_table_rows_spin': %s", err)
	}

	rsb, ok := rsbo.(*gtk.SpinButton)
	if !ok {
		log.Fatal("object with id 'dialog_table_rows_spin' is not a gtk.SpinButton")
	}

	defer d.Destroy()
	if r := d.Run(); r != gtk.RESPONSE_OK {
		return
	}

	cc := csb.GetValueAsInt()
	rr := rsb.GetValueAsInt()

	t := "|"
	for c := 1; c <= cc; c++ {
		t += " header |"
	}
	t += "\n|"

	for c := 1; c <= cc; c++ {
		t += " ---          |"
	}

	for r := 1; r <= rr; r++ {
		t += "\n|"
		for c := 1; c <= cc; c++ {
			t += " cell        |"
		}
	}
	t += "\n"

	a.textBuffer.InsertAtCursor(t)
}
