package hl7converter

import (
	"errors"
	"fmt"
)

type ErrWrapped struct {
	Err            error
	AdditionalInfo string
}

// Реализуем метод для ошибки, чтобы реализовать интерфейс error
func (e *ErrWrapped) Error() string {
	return fmt.Sprintf("%s, %s", e.Err.Error(), e.AdditionalInfo)
}

// Метод Is позволяет errors.Is работать с нашей ошибкой
func (e *ErrWrapped) Is(target error) bool {
	return e.Err.Error() == target.Error()
}

var (
	// CONFIG ERRORS
	ErrInvalidJSON          = errors.New("config error: specified modification is not 'map[string]any'")
	ErrInvalidJsonExtension = errors.New("config error: path doesn't contains extension 'json'")
	ErrInvalidConfig        = errors.New("config error: validate json has been unsuccessful")
	ErrModificationNotFound = errors.New("config error: specified modification is not found in config")
	ErrNilModification      = errors.New("config error: modification was incorrectly read from the file because it's empty")

	// CONVERTING ERRORS
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

	// RESULT_MODEL ERRORS
	ErrIndexOutOfRange = errors.New("index out of range")
)

// ______________________________[CONFIG ERRORS]______________________________

func NewErrModificationNotFound(modificationName string) error {
	return &ErrWrapped{
		Err:            ErrModificationNotFound,
		AdditionalInfo: fmt.Sprintf("modification %s", modificationName),
	}
}

func NewErrInvalidJSON(modification any) error {
	return &ErrWrapped{
		Err:            ErrInvalidJSON,
		AdditionalInfo: fmt.Sprintf("value %v", modification),
	}
}

func NewErrInvalidJsonExtension(path string) error {
	return &ErrWrapped{
		Err:            ErrInvalidJsonExtension,
		AdditionalInfo: fmt.Sprintf("path %s", path),
	}
}

func NewErrNilModification(modificationName, path string) error {
	return &ErrWrapped{
		Err:            ErrNilModification,
		AdditionalInfo: fmt.Sprintf("modification %s path %s", modificationName, path),
	}
}

// ______________________________[CONVERTING ERRORS]______________________________

func NewFatalErrOfConverting(r any) error {
	return &ErrWrapped{
		Err:            FatalErrOfConverting,
		AdditionalInfo: fmt.Sprintf("recovered %+v", r),
	}
}

func NewErrInvalidParseMsg(err string) error {
	return &ErrWrapped{
		Err:            ErrInvalidParseMsg,
		AdditionalInfo: fmt.Sprintf("system error %s", err),
	}
}

func NewErrInputTagNotFound(row string) error {
	return &ErrWrapped{
		Err:            ErrInputTagNotFound,
		AdditionalInfo: fmt.Sprintf("row %s", row),
	}
}

func NewErrOutputTagNotFound(tag string) error {
	return &ErrWrapped{
		Err:            ErrOutputTagNotFound,
		AdditionalInfo: fmt.Sprintf("input tag %s", tag),
	}
}

func NewErrUndefinedOption(option, tag string) error {
	return &ErrWrapped{
		Err:            ErrUndefinedOption,
		AdditionalInfo: fmt.Sprintf("option %s tag %s, avaliable %+v", option, tag, mapOfOptions),
	}
}

func NewErrUndefinedPositionTag(tag string) error {
	return &ErrWrapped{
		Err:            ErrUndefinedPositionTag,
		AdditionalInfo: fmt.Sprintf("tag %s", tag),
	}
}

func NewErrInvalidLink(link string) error {
	return &ErrWrapped{
		Err:            ErrInvalidLink,
		AdditionalInfo: fmt.Sprintf("link %s", link),
	}
}

func NewErrInvalidLinkElems(link string) error {
	return &ErrWrapped{
		Err:            ErrInvalidLinkElems,
		AdditionalInfo: fmt.Sprintf("link %s", link),
	}
}

func NewErrWrongParamCount(field, param string) error {
	return &ErrWrapped{
		Err:            ErrWrongParamCount,
		AdditionalInfo: fmt.Sprintf("field %s param %s", field, param),
	}
}

func NewErrEmptyDefaultValue(field string) error {
	return &ErrWrapped{
		Err:            ErrEmptyDefaultValue,
		AdditionalInfo: fmt.Sprintf("field %s", field),
	}
}

func NewErrUndefinedInputTag(tag, someinfo string) error {
	return &ErrWrapped{
		Err:            ErrUndefinedInputTag,
		AdditionalInfo: fmt.Sprintf("tag %s additional info %s", tag, someinfo),
	}
}

func NewErrTooBigIndex(idx, maxIdx int) error {
	return &ErrWrapped{
		Err:            ErrTooBigIndex,
		AdditionalInfo: fmt.Sprintf("index %d maxIndex %d", idx, maxIdx),
	}
}

func NewErrWrongFieldsNumber(tag string, tagstruct *Tag, currentFieldsNumb int) error {
	return &ErrWrapped{
		Err:            ErrWrongFieldsNumber,
		AdditionalInfo: fmt.Sprintf("tagName %s tagSturcture %+v current fields has %d", tag, tagstruct, currentFieldsNumb),
	}
}

func NewErrWrongComponentsNumber(inputdata, link string) error {
	return &ErrWrapped{
		Err:            ErrWrongComponentsNumber,
		AdditionalInfo: fmt.Sprintf("line %s link %s", inputdata, link),
	}
}

func NewErrWrongComponentLink(link string, compPos, compCount int, inputTag string) error {
	return &ErrWrapped{
		Err:            ErrWrongComponentLink,
		AdditionalInfo: fmt.Sprintf("tag %s link %s componentPosition %d componentCount %d", inputTag, link, compPos, compCount),
	}
}

// ______________________________[RESULT_MODEL ERRORS]______________________________

func NewErrIndexOutOfRange(idx, max int, elem string) error {
	return &ErrWrapped{
		Err:            ErrIndexOutOfRange,
		AdditionalInfo: fmt.Sprintf("index %d max %d elem %s", idx, max, elem),
	}
}
