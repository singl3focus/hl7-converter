package hl7converter

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/robertkrimen/otto"
)

var (
	ErrIndexOutOfRange = errors.New("index out of range")

	ErrScriptNilKey  = errors.New("script error: key is nil")
	ErrScriptCompile = errors.New("script error: compilation failure")
	ErrScriptRun     = errors.New("script error: run failure")

	ErrFieldEmptyComponents = errors.New("field error: empty components")
	ErrFieldEmptyArray      = errors.New("field error: empty array")

	ErrAliasLinkTagNotExists    = errors.New("alias error: link tag not exists")
	ErrAliasInvalidLinkPosition = errors.New("alias error: invalid link position")
)

func NewErrIndexOutOfRange(idx, max int, elem string) error {
	return NewError(ErrIndexOutOfRange, fmt.Sprintf("index %d max %d elem %s", idx, max, elem))
}

/*
	For flexible work with message
*/

// [Result]
type Result struct {
	LineSeparator string

	Rows []*Row

	aliases Aliases

	vmOnce RetryableOnce
	otto   *otto.Otto
}

func NewResult(ls string, rws []*Row) *Result {
	return &Result{
		LineSeparator: ls,
		Rows:          rws,
		otto:          otto.New(), // TODO: check mem and cpu usage (need optimizations)
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

func (r *Result) Bytes() []byte {
	if r == nil {
		return []byte{}
	}

	var builder bytes.Buffer
	for i, v := range r.Rows {
		builder.Write(v.Bytes())
		if i != len(r.Rows)-1 {
			builder.WriteString(r.LineSeparator)
		}
	}

	return builder.Bytes()
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

// TODO: Add InsertRow

// TODO: Add RemoveRowByIndex

type constraint interface {
	isAllowed()
}

type keyScript string

func (k keyScript) isAllowed() {}

const KeyScript keyScript = "msg"

// UseScript run javascript code block.
// Param 'scr' may be a string, a byte slice, a bytes.Buffer, or an io.Reader,
// but it MUST always be in UTF-8.
// Param 'k' set as required argument in order to specify developer use 'KeyScript' in their script.
func (r *Result) UseScript(k constraint, scr any) error {
	if k == nil {
		return ErrScriptNilKey
	}

	err := r.vmOnce.Do(func() error {
		err := r.otto.Set(string(KeyScript), r)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	script, err := r.otto.Compile("", scr)
	if err != nil {
		return ErrScriptCompile
	}

	_, err = r.otto.Run(script)
	if err != nil {
		return ErrScriptRun
	}

	return nil
}

// FindTag return row with first matched tag.
func (r *Result) FindTag(tag string) (*Row, bool) {
	for _, row := range r.Rows {
		if t, ok := row.Tag(); ok && t == tag {
			return row, true
		}
	}

	return nil, false
}

// TODO: it's repeat logic of converter, how join it to single funcs not
func (r *Result) ApplyAliases(a Aliases) error {
	for name, link := range a {
		elems := strings.Split(link, linkToField) // parse: Tag - Position
		if len(elems) != 2 {
			return NewErrInvalidLink(link)
		}
		
		tag, position := elems[0], elems[1]
		if tag == "" || position == "" {
			return NewErrInvalidLinkElems(link)
		}

		pos, err := strconv.ParseFloat(position, 64)
		if err != nil {
			return err
		}

		row, exist := r.FindTag(tag)
		if !exist {
			return NewError(ErrAliasLinkTagNotExists, fmt.Sprintf("name %s link %s tag %s", name, link, tag))
		}

		if isInt(pos) {
			fieldIndx := int(pos) - 1 // * Danger

			if !row.checkRange(fieldIndx) {
				return NewError(ErrAliasInvalidLinkPosition, fmt.Sprintf("name %s link %s", name, link))
			}

			f := row.Fields[fieldIndx]

			a[name] = f.Value
		} else {
			fieldIndx, componentIndx := int(pos)-1, getTenth(pos)-1 // * Danger

			if !row.checkRange(fieldIndx) {
				return NewError(ErrAliasInvalidLinkPosition, fmt.Sprintf("name %s link %s", name, link))
			}

			f := row.Fields[fieldIndx]			
			
			comp := f.Components()
			if !comp.checkRange(componentIndx) {
				return NewError(ErrAliasInvalidLinkPosition, fmt.Sprintf("invalid component pos, name %s link %s", name, link))
			}

			a[name] = comp[componentIndx]
		}
	}

	r.aliases = a

	return nil
}

func (r *Result) Aliases() (Aliases, bool) {
	return r.aliases, len(r.aliases) != 0
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

func NewRowWithFieldsValues(fs, cs, cas string, f ...string) *Row {
	fds := make([]*Field, 0, len(f))
	for _, v := range f {
		fds = append(fds, NewField(v, cs, cas))
	}

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

func (r *Row) Bytes() []byte {
	if r == nil {
		return []byte{}
	}

	var builder bytes.Buffer
	for i, f := range r.Fields {
		builder.Write(f.Bytes())
		if i != len(r.Fields)-1 {
			builder.WriteString(r.FieldSeparator)
		}
	}

	return builder.Bytes()
}

func (r *Row) Tag() (string, bool) {
	if len(r.Fields) == 0 {
		return "", false
	}

	tag := r.Fields[0]

	return tag.Value, !tag.IsEmpty()
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
	comps     Components

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
		arrSep:   " ",
		compsSep: componentSeparator,
	}
}

func (f *Field) String() string {
	if f == nil {
		return "<nil>"
	}

	return f.Value
}

func (f *Field) Bytes() []byte {
	return []byte(f.Value)
}

func (f *Field) IsEmpty() bool {
	return f.Value == ""
}

func (f *Field) Components() Components {
	f.compsOnce.Do(func() {
		a := strings.ReplaceAll(f.Value, f.arrSep, f.compsSep)
		f.comps = strings.Split(a, f.compsSep) // TODO: Does it have correct behaviour?
	})

	return f.comps
}

// ComponentsChecked returns (Components, error), if slice does not meet expectations.
func (f *Field) ComponentsChecked() (Components, error) {
	comps := f.Components()
	if len(comps) == 0 {
		return nil, ErrFieldEmptyComponents
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
//		/* It's so comfortable, if i'm sure that's array must be exists (len > 0) */
//			_ = res.Rows[0].Fields[3].Array()[0].ChangeValue("180")
//	 	/* I need do some checks, but i'm sure that's array must be exists */
//			_, err = res.Rows[0].Fields[3].Array()
//	 	if err != nil {...}
//
// Ready.
func (f *Field) Array() []*Field {
	f.arrOnce.Do(func() {
		// TODO: Does it have correct behaviour?
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
		return nil, ErrFieldEmptyArray
	}

	return arr, nil
}

func (f *Field) ChangeValue(v string) {
	f.Value = v

	// Reset for re-init in next call of Array(), Components()
	f.arrOnce = sync.Once{}
	f.compsOnce = sync.Once{}
}

type Components []string

func (c Components) Original() []string {
	return c
}

func (c Components) checkRange(i int) bool {
	return i >= 0 && i < len(c)
}
