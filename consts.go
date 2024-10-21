package hl7converter

import (
	"errors"
	"fmt"
)

/*
APPLICATION
*/
const (
	CfgJSON       = "config.json"
	CfgJSONSchema = "config.schema.json"
)

const ( // [CONVERTING PARAMS]
	ignoredIndx = 0

	ignoredFieldsNumber = -1
)

const ( // [SPECIAL CONVERTER SYMBOLS]
	linkElemSt  = "<"
	linkElemEnd = ">"
	linkToField = "-"

	OR = "??"

	itSymbol = 1
	itLink   = 0
)

var (
	mapOfOptions = map[string]string{
		"autofill": "automaticly adding empty fields by count of differents about fields_number and current_fields_number",
	}
)

/*
	ERRORS
*/

var ( // [PARSING JSON CONFIG ERRORS]
	ErrInvalidConfig = errors.New("validate json has been unsuccessful")
)

// ______________________________[CONFIG ERRORS]______________________________

func NewErrModificationNotFound(mdfname string) error {
	return fmt.Errorf("specified modification (name=%s) is not found in config", mdfname)
}

func NewErrInvalidJSON(mdf any) error {
	return fmt.Errorf("invalid JSON, specified modification (value=%v) is not 'map[string]any'", mdf)
}

func NewErrInvalidJsonExtension(path string) error {
	return fmt.Errorf("config (path=%s) doesn't contains extension 'json'", path)
}

func NewErrNilModification(name, path string) error {
	return fmt.Errorf("modification (name=%s) was incorrectly read from the file (path=%s), it's empty", name, path)
}

// ______________________________[CONVERTING ERRORS]______________________________

func NewErrInvalidParseMsg(errMsg string) error {
	return fmt.Errorf("parse input messsge to struct has been unsuccesful (error=%s)", errMsg)
}

func NewErrInputTagNotFound(row string) error {
	return fmt.Errorf("input tag not found (row=%s)", row)
}

func NewErrOutputTagNotFound(tagname string) error {
	return fmt.Errorf("linked tags in tag (name=%s) not found in output modification", tagname)
}

func NewErrUndefinedOption(optionname, tagname string) error {
	return fmt.Errorf("undefined option (name=%s) by tag (name=%s); avaliable: %+v", optionname, tagname, mapOfOptions)
}

func NewErrUndefinedPositionTag(tagname string) error {
	return fmt.Errorf("tag (name=%s) has position, but it's not found in tags", tagname)
}

func NewErrInvalidLink(link string) error {
	return fmt.Errorf("invalid link (str=%s)", link)
}

func NewErrInvalidLinkElems(link string) error {
	return fmt.Errorf("invalid link (str=%s), some elems empty", link)
}

func NewErrWrongParamCount(fv, p string) error {
	return fmt.Errorf("field (value=%s), you can use only one special symbol (param=%s)", fv, p)
}

func NewErrEmptyDefaultValue(fv string) error {
	return fmt.Errorf("field (value=%s) has empty default value", fv)
}


func NewErrUndefinedInputTag(tagname, someinfo string) error {
	return fmt.Errorf("undefined input tag (name=%s), some info: %s", tagname, someinfo)
}

func NewErrTooBigIndex(indx, maxIndx int) error {
	return fmt.Errorf("index (number=%d) of output rows more than max index (number=%d) of input rows", indx, maxIndx)
}


const (
	ErrWrongFieldsNumber  = "tag (name=%s, structure=%v) has invalid specified count of fields number, current fields has (number=%d)"
	ErrWrongFieldPosition = "wrong position (position must be more than 1) by fieldName %s in %+v"
	ErrWrongFieldLink     = "specified field link is incorrect, field %s"
	ErrWrongFieldMetadata = "fieldName %s in output modification %v hasn't have a linked_fields/default_value is not specified/input_field not found in linked_fields"

	ErrWrongComponentsNumber = "component not found (line=%s), but link (value=%s) is exist"
	ErrWrongComponentLink    = "link (value=%s) component position (value=%d) more than max components count (number=%d) of input row with tag (name=%s)"

	ErrNegativeComponentsNumber = "commponent count can be equal to 0 or more than 1, field (name=%s)"
	ErrWrongComponentSplit      = "incorrect field (value=%s), the component could not be pulled out"

	ErrManyMultiTags = "converter can working with one multi tag not more(wait: %s, receive: %s)"
)


// ______________________________[Main Model Errors]______________________________


func NewErrIndexOutOfRange(idx, max uint, elemname string) error {
	return fmt.Errorf("index (value=%d) out of range, max (value=%d), elem (name=%s)", idx, max, elemname)
}
