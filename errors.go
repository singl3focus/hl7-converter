package hl7converter

import (
	"errors"
	"fmt"
)


var _ error = Error{} // check of error implementation

type Error struct {
	Err            error
	AdditionalInfo string
}

func NewError(reason error, info string) Error {
	Assert(reason != nil, "init Error with nil error as reason, info %s", info)

	return Error{
		Err: reason,
		AdditionalInfo: info,
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("%s, %s", e.Err.Error(), e.AdditionalInfo)
}

// Is allow errors.Is define target err
func (e Error) Is(target error) bool {
	return e.Err.Error() == target.Error()
}


func Assert(condition bool, format string, fields ...any) {
    if !condition {
        panic(fmt.Sprintf(format, fields...))
    }
}



var (
	FatalErrOfConverting     = errors.New("convert error: parse input message to struct has been unsuccesful")
	ErrInvalidParseMsg       = errors.New("convert error: parse input message to struct has been unsuccesful")
	ErrInputTagNotFound      = errors.New("convert error: input tag not found")
	ErrOutputTagNotFound     = errors.New("convert error: linked tags in tag not found in output modification")
	ErrUndefinedOption       = errors.New("convert error: undefined option by tag")
	ErrUndefinedPositionTag  = errors.New("convert error: tag has position, but it's not found in tags")
	ErrInvalidLink           = errors.New("convert error: invalid link")
	ErrInvalidLinkElems      = errors.New("convert error: invalid link, some elems empty")
	ErrWrongParamCount       = errors.New("convert error: you can use only one special symbol in field")
	ErrEmptyDefaultValue     = errors.New("convert error: field has empty default value")
	ErrUndefinedInputTag     = errors.New("convert error: undefined input tag")
	ErrTooBigIndex           = errors.New("convert error: index of output rows more than max index of input rows")
	ErrWrongFieldsNumber     = errors.New("convert error: tag has invalid specified count of fields number")
	ErrWrongComponentsNumber = errors.New("convert error: component not found but link is exist")
	ErrWrongComponentLink    = errors.New("convert error: link component position more than max components count of input row with tag")
)


func NewFatalErrOfConverting(r any) error {
	return &Error{
		Err:            FatalErrOfConverting,
		AdditionalInfo: fmt.Sprintf("recovered %+v", r),
	}
}

func NewErrInvalidParseMsg(err string) error {
	return &Error{
		Err:            ErrInvalidParseMsg,
		AdditionalInfo: fmt.Sprintf("system error %s", err),
	}
}

func NewErrInputTagNotFound(row string) error {
	return &Error{
		Err:            ErrInputTagNotFound,
		AdditionalInfo: fmt.Sprintf("row %s", row),
	}
}

func NewErrOutputTagNotFound(tag string) error {
	return &Error{
		Err:            ErrOutputTagNotFound,
		AdditionalInfo: fmt.Sprintf("input tag %s", tag),
	}
}

func NewErrUndefinedOption(option, tag string) error {
	return &Error{
		Err:            ErrUndefinedOption,
		AdditionalInfo: fmt.Sprintf("option %s tag %s, avaliable %+v", option, tag, mapOfOptions),
	}
}

func NewErrUndefinedPositionTag(tag string) error {
	return &Error{
		Err:            ErrUndefinedPositionTag,
		AdditionalInfo: fmt.Sprintf("tag %s", tag),
	}
}

func NewErrInvalidLink(link string) error {
	return &Error{
		Err:            ErrInvalidLink,
		AdditionalInfo: fmt.Sprintf("link %s", link),
	}
}

func NewErrInvalidLinkElems(link string) error {
	return &Error{
		Err:            ErrInvalidLinkElems,
		AdditionalInfo: fmt.Sprintf("link %s", link),
	}
}

func NewErrWrongParamCount(field, param string) error {
	return &Error{
		Err:            ErrWrongParamCount,
		AdditionalInfo: fmt.Sprintf("field %s param %s", field, param),
	}
}

func NewErrEmptyDefaultValue(field string) error {
	return &Error{
		Err:            ErrEmptyDefaultValue,
		AdditionalInfo: fmt.Sprintf("field %s", field),
	}
}

func NewErrUndefinedInputTag(tag, someinfo string) error {
	return &Error{
		Err:            ErrUndefinedInputTag,
		AdditionalInfo: fmt.Sprintf("tag %s additional info %s", tag, someinfo),
	}
}

func NewErrTooBigIndex(idx, maxIdx int) error {
	return &Error{
		Err:            ErrTooBigIndex,
		AdditionalInfo: fmt.Sprintf("index %d maxIndex %d", idx, maxIdx),
	}
}

func NewErrWrongFieldsNumber(tag string, tagstruct *Tag, currentFieldsNumb int) error {
	return &Error{
		Err:            ErrWrongFieldsNumber,
		AdditionalInfo: fmt.Sprintf("tagName %s tagSturcture %+v current fields has %d", tag, tagstruct, currentFieldsNumb),
	}
}

func NewErrWrongComponentsNumber(inputdata, link string) error {
	return &Error{
		Err:            ErrWrongComponentsNumber,
		AdditionalInfo: fmt.Sprintf("line %s link %s", inputdata, link),
	}
}

func NewErrWrongComponentLink(link string, compPos, compCount int, inputTag string) error {
	return &Error{
		Err:            ErrWrongComponentLink,
		AdditionalInfo: fmt.Sprintf("tag %s link %s componentPosition %d componentCount %d", inputTag, link, compPos, compCount),
	}
}
