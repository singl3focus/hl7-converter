package hl7converter

import (
	"errors"
	"strings"
	"sync"

	"github.com/robertkrimen/otto"
)

/*
	For flexible work with message
*/

// [Result]
type Result struct {
	LineSeparator string

	Rows []*Row

	vmOnce sync.Once
	Otto *otto.Otto
}

func NewResult(ls string, rws []*Row) *Result {
	return &Result{
		LineSeparator: ls,
		Rows:          rws,
		Otto:          otto.New(), // todo: check mem and cpu usage (need optimizations)
	}
}

// String returns string representation of result.
// Be careful: if result == nil, it's also returns non-empty string.
func (r *Result) String() string {
	if r == nil {
		return "<nil>"
	}

	var builder strings.Builder
	for i, v := range r.Rows {
		builder.WriteString(v.String())
		if i != len(r.Rows)-1 {
			builder.WriteString(r.LineSeparator)
		}
	}

	return builder.String()
}

func (r *Result) checkRange(i int) bool {
	return i >= 0 && i < len(r.Rows)
}

func (r *Result) SwapRows(p1, p2 int) error {
	if !r.checkRange(p1) {
		return NewErrIndexOutOfRange(p1, len(r.Rows), "rows")
	}

	if !r.checkRange(p2) {
		return NewErrIndexOutOfRange(p2, len(r.Rows), "rows")
	}

	r.Rows[p1], r.Rows[p2] = r.Rows[p2], r.Rows[p1]

	return nil
}

func (r *Result) SetRow(p int, row *Row) error {
	if !r.checkRange(p) {
		return NewErrIndexOutOfRange(p, len(r.Rows), "rows")
	}

	r.Rows[p] = row

	return nil
}


type constraint interface {
	isAllowed()
}

type keyScript string

func (k keyScript) isAllowed() {}

const KeyScript keyScript = "msg"

// UseScript run javascript 
// scr may be a string, a byte slice, a bytes.Buffer, or an io.Reader, but it MUST always be in UTF-8.
func (r *Result) UseScript(k constraint, scr any) error {
	if k == nil {
		return errors.New("what are you doing, bro?!") // todo: ErrNilKeyScript
	}

	r.vmOnce.Do(func() {
		err := r.Otto.Set(string(KeyScript), r) // todo: we need specifie 
		if err != nil {
			panic(err) // todo: what to do?
		}
	})

	script, err := r.Otto.Compile("", scr)
	if err != nil {
		return err // todo: error recognition
	}

	_, err = r.Otto.Run(script)
	if err != nil {
		return err // todo: error recognition
	}

	return nil
}

// [Row]
type Row struct {
	FieldSeparator string

	Fields []*Field
}

func NewRow(fs string, fds []*Field) *Row {
	return &Row{
		FieldSeparator: fs,
		Fields:         fds,
	}
}

func (r *Row) String() string {
	if r == nil {
		return "<nil>"
	}

	var builder strings.Builder
	for i, f := range r.Fields {
		builder.WriteString(f.String())
		if i != len(r.Fields)-1 {
			builder.WriteString(r.FieldSeparator)
		}
	}

	return builder.String()
}

func (r *Row) Tag() (string, bool) {
	if len(r.Fields) == 0 {
		return "", false
	}

	tag := r.Fields[0]

	return tag.Value, tag.Value != ""
}

func (r *Row) checkRange(i int) bool {
	return i >= 0 && i < len(r.Fields)
}

func (r *Row) SwapFields(p1, p2 int) error {
	if !r.checkRange(p1) {
		return NewErrIndexOutOfRange(p1, len(r.Fields), "fields")
	}

	if !r.checkRange(p2) {
		return NewErrIndexOutOfRange(p2, len(r.Fields), "fields")
	}

	r.Fields[p1], r.Fields[p2] = r.Fields[p2], r.Fields[p1]

	return nil
}

// ChangeFieldPosition move field from old position (oldp) to new position (newp)
// and set empty field to new position
func (r *Row) ChangeFieldPosition(oldp, newp int) error {
	if !r.checkRange(oldp) {
		return NewErrIndexOutOfRange(oldp, len(r.Fields), "fields")
	}

	if !r.checkRange(newp) {
		return NewErrIndexOutOfRange(newp, len(r.Fields), "fields")
	}

	r.Fields[newp], r.Fields[oldp] = r.Fields[oldp], &Field{}

	return nil
}

// SetField
func (r *Row) SetField(p int, f *Field) error {
	if !r.checkRange(p) {
		return NewErrIndexOutOfRange(p, len(r.Fields), "fields")
	}

	r.Fields[p] = f

	return nil
}

// [Field]
type Field struct {
	Value string

	compsSep  string
	compsOnce sync.Once
	comps     []string

	arrSep  string
	arrOnce sync.Once
	array   []*Field
}

func NewField(value, componentSeparator, componentArraySeparator string) *Field {
	field := &Field{
		Value:    value,
		arrSep:   componentArraySeparator,
		compsSep: componentSeparator,
	}

	return field
}

func newArrayField(value, componentSeparator string) *Field {
	return &Field{
		Value: value,
		// Array field cannot have an array separator because it's smallest unit,
		// but for correct work of strings.ReplaceAll we use " " instead of default string value "" 
		arrSep: " ",
		compsSep: componentSeparator,
	}
}

func (f *Field) String() string {
	if f == nil {
		return "<nil>"
	}

	return f.Value
}

func (f *Field) Components() []string {
	f.compsOnce.Do(func() {
		a := strings.ReplaceAll(f.Value, f.arrSep, f.compsSep)
		f.comps = strings.Split(a, f.compsSep) // todo: Does it have correct behaviour?
	})
	return f.comps
}

// ComponentsChecked returns ([]string, error), if slice does not meet expectations.
func (f *Field) ComponentsChecked() ([]string, error) {
	comps := f.Components()
	if len(comps) == 0 {
		return nil, errors.New("field error: empty components") // todo: ErrEmptyComponents
	}
	return comps, nil
}

// Array return slice of arr fields that's sepatated by atrSep.
// The returned slice is always non-nil, but it's can be empty if arrSep not found in Value.
// Notation(comma ok) not be here( this notation can specify on empty slice),
// because this makes it easier to use.
//
// I know this can have vulnerabilities, but it's more flexible. This behaviour is repeated in Components().
//
// What i means:
//
//			/* It's so comfortable, if i'm sure that's array must be exists (len > 0) */
//			_ = res.Rows[0].Fields[3].Array()[0].ChangeValue("180")
//	 	/* I need do some checks, but i'm sure that's array must be exists */
//			_, err = res.Rows[0].Fields[3].Array()
//	 	if err != nil {...}
func (f *Field) Array() []*Field {
	f.arrOnce.Do(func() {
		// todo: Does it have correct behaviour?
		arrElem := strings.Split(f.Value, f.arrSep)

		fieldArr := make([]*Field, 0, len(arrElem))
		if len(arrElem) > 1 {
			for _, v := range arrElem {
				fieldArr = append(fieldArr, newArrayField(v, f.compsSep))
			}
		}

		f.array = fieldArr
	})

	return f.array
}

// ArrayChecked returns ([]*Field, error), if slice does not meet expectations.
func (f *Field) ArrayChecked() ([]*Field, error) {
	arr := f.Array()
	if len(arr) == 0 {
		return nil, errors.New("field error: empty array") // todo: ErrEmptyArray
	}
	return arr, nil
}

func (f *Field) ChangeValue(v string) {
	f.Value = v

	// Reset for re-init in next call of Array(), Components()
	f.arrOnce = sync.Once{}
	f.compsOnce = sync.Once{}
}
