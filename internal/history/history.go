package history

import (
	"fmt"
)

type History interface {
	Do(string) *string
	Undo(string) *string
	Add(Record)
}

type unboundedHistory struct {
	records Records
	cursor  int
	last    int
}

type Records []Record

func (rr Records) String() string {
	s := "["
	for r := range rr {
		s += fmt.Sprintf(" %v,", r)
	}
	s = s[:len(s)-1] + " ]"
	return s
}

type Record interface {
	Do(to string) string
	Undo(to string) string
}

func NewUnbounded() History {
	return &unboundedHistory{
		records: make([]Record, 0),
	}
}

func (h *unboundedHistory) Do(t string) *string {
	if h.last == h.cursor {
		return nil
	}

	r := h.records[h.cursor]
	h.cursor += 1

	tu := r.Do(t)
	return &tu
}

func (h *unboundedHistory) Undo(t string) *string {
	if h.cursor == 0 {
		return nil
	}

	r := h.records[h.cursor-1]
	h.cursor -= 1

	tu := r.Undo(t)
	return &tu
}

func (h *unboundedHistory) Add(r Record) {
	h.records = append(h.records[:h.cursor], r)
	h.cursor += 1
	h.last = h.cursor
}

func NewInsertRecord(text string, offset, length int) Record {
	return &insertRecord{
		text:   text,
		offset: offset,
		length: length,
	}
}

type insertRecord struct {
	offset int
	length int
	text   string
}

func (r *insertRecord) Undo(to string) string {
	s, e := to[0:r.offset], to[r.offset+r.length:]
	if r.offset > 0 {
		s = to[0:r.offset]
	}

	return s + e
}

func (r *insertRecord) Do(to string) string {
	s, e := "", to[r.offset:]
	if r.offset > 0 {
		s = to[0:r.offset]
	}

	return s + r.text + e
}

func NewDeleteRecord(text string, start, end int) Record {
	return &deleteRecord{
		text:  text,
		start: start,
		end:   end,
	}
}

type deleteRecord struct {
	start int
	end   int
	text  string
}

func (r *deleteRecord) Do(to string) string {
	s, e := "", to[r.end:]
	if r.start > 0 {
		s = to[0:r.start]
	}

	return s + e
}

func (r *deleteRecord) Undo(to string) string {
	s, e := "", to[r.start:]
	if r.start > 0 {
		s = to[0:r.start]
	}

	return s + r.text + e
}
