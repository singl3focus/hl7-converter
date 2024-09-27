package hl7converter

import "fmt"

/*
	For working with output message
*/

type Result struct {
	LineSeparator string

	Rows []*Row
}

func (r *Result) AssembleMessage() string {
	result := ""

	for i, v := range r.Rows {
		result += v.AssembleRow()

		if i != len(r.Rows)-1 {
			result += r.LineSeparator
		}
	}

	return result
}

func (r *Result) checkRange(i uint) bool {
	return int(i) < len(r.Rows)
}

func (r *Result) SwapRows(p1, p2 uint) error {
	if !r.checkRange(p1) {
		return fmt.Errorf(ErrIndexOutOfRange, p1)
	}

	if !r.checkRange(p2) {
		return fmt.Errorf(ErrIndexOutOfRange, p2)
	}

	temp := r.Rows[p1]
	r.Rows[p1] = r.Rows[p2]
	r.Rows[p2] = temp

	return nil
}

/*__________________________*/

type Row struct {
	FieldSeparator string

	Fields []*Field
}

func (r *Row) AssembleRow() string {
	result := ""

	for i, v := range r.Fields {
		result += v.Value

		if i != len(r.Fields)-1 {
			result += r.FieldSeparator
		}
	}

	return result
}

func (r *Row) Tag() (string, bool) {
	tag := r.Fields[0]
	if tag.Value == "" {
		return "", false
	}

	return tag.Value, true
}

func (r *Row) checkRange(i uint) bool {
	return int(i) < len(r.Fields)
}

func (r *Row) SwapFields(p1, p2 uint) error {
	if !r.checkRange(p1) {
		return fmt.Errorf(ErrIndexOutOfRange, p1)
	}

	if !r.checkRange(p2) {
		return fmt.Errorf(ErrIndexOutOfRange, p2)
	}

	temp := r.Fields[p1]
	r.Fields[p1] = r.Fields[p2]
	r.Fields[p2] = temp

	return nil
}

func (r *Row) ChangeFieldPosition(oldp, newp uint) error {
	if !r.checkRange(oldp) {
		return fmt.Errorf(ErrIndexOutOfRange, oldp)
	}

	if !r.checkRange(newp) {
		return fmt.Errorf(ErrIndexOutOfRange, newp)
	}

	r.Fields[newp] = r.Fields[oldp]
	r.Fields[oldp] = &Field{}

	return nil
}

/*__________________________*/

type Field struct {
	Value string
}

func (f *Field) ChangeValue(nv string) {
	f.Value = nv
}
