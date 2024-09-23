package hl7converter

import "errors"

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

	AND = "$$" // ??? - maybe alias for component separator
	OR  = "??"

	itSymbol = 1
	itLink   = 0
)

var (
	mapOfOptions = map[string]string{
		"autofill": "automaticly adding  empty fields by count of differents about fields_number and current_fields_number",
	}
)

/*
	ERRORS
*/

var ( // [CONFIG ERRORS]
	ErrModificationNotFound = errors.New("specified modification is not found in config")
	ErrInvalidJSON          = errors.New("invalid JSON, specified modification is not 'map[string]any'")
)

var ( // [PARSING JSON CONFIG ERRORS]
	ErrInvalidConfig        = errors.New("validate json has been unsuccessful")
	ErrInvalidJsonExtension = "config (path=%s) doesn't contains extension 'json'"
	ErrNilModification      = "modification (name=%s) was incorrectly read from the file (path=%s), it's empty"
)


var ( // [CONVERTING ERRORS]
	ErrInvalidParseMsg   = errors.New("parse input messsge to struct has been unsuccesful")
	ErrOutputTagNotFound = errors.New("linked tags in tag not found in output modification")
)

const ( // [CONVERTING OPTIONS f-ERRORS]
	ErrUndefinedOption = "undefined option (name=%s) by tag (name=%s)"
)

const ( // [CONVERTING f-ERRORS]
	ErrUndefinedPositionTag = "tag (name=%s) has position, but it's not found in tags"

	ErrInvalidLink       = "invalid link (str=%s)"
	ErrInvalidLinkElems  = "invalid link (str=%s), some elems empty"
	ErrWrongParamCount   = "field (value=%s), you can use only one (param=%s)"
	ErrEmptyDefaultValue = "field (value=%s) has empty default value"

	ErrWrongTagRow         = "identify msg and split input msg has been unsuccessful (row=%s)"
	ErrUndefinedInputTag   = "undefined input tag(name=%s), some info: %s"
	ErrUndefinedOutputTag  = "undefined output tag (name=%s), some info: %s"
	ErrTooBigIndex         = "index(number=%d) of output rows more than max index(number=%d) of input rows"
	ErrWrongSliceOfTag     = "'slice of Tag' by tag %s is empty"
	ErrInputTagMSGNotFound = "tag %s not found in search structure MSG %v"

	ErrWrongFieldsNumber  = "tag(name=%s, structure=%v) has invalid specified count of fields number, current fields number(%d)"
	ErrWrongFieldPosition = "wrong position (position must be more than 1) by fieldName %s in %+v"
	ErrWrongFieldLink     = "specified field link is incorrect, field %s"
	ErrWrongFieldMetadata = "fieldName %s in output modification %v hasn't have a linked_fields/default_value is not specified/input_field not found in linked_fields"

	ErrWrongComponentsNumber = "component not found (line=%s), but link (value=%s) is exist"
	ErrWrongComponentLink    = "link (value=%s) component position (value=%d) more than max components count (number=%d) of input row with tag (name=%s)"

	ErrNegativeComponentsNumber = "commponent count can be equal to 0 or more than 1. FieldName %s"
	ErrWrongComponentSplit      = "incorrect field %s, the field component could not be pulled out"

	ErrManyMultiTags = "converter can working with one multi tag not more, wait: %s, receive: %s"
)
