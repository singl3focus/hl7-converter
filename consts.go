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
)

const ( // [SPECIAL CONVERTER SYMBOLS]
	linkElemSt  = "<"
	linkElemEnd = ">"

	AND = "$$" // ??? - maybe alias for component separator
	OR  = "??"

	linkToField = "-"
)

var (
	mapOfOptions = map[string]string {
		"autofill": "automaticly adding  empty fields by count of differents about fields_number and current_fields_number",
	}
)

/*
	ERRORS
*/

var ( // [CONFIG ERRORS]
	ErrModificationNotFound = errors.New("specified modification is not found in config")
	ErrInvalidJSON          = errors.New("JSON is invalid. Specified modification is not 'map[string]any'")
)

var ( // [PARSING JSON CONFIG ERRORS]
	ErrInvalidConfig        = errors.New("validate json has been unsuccessful")
	ErrInvalidJsonExtension = "config (path=%s) doesn't contains extension 'json'"
	ErrNilModification      = "modification (name=%s) was incorrectly read from the file (path=%s), it's empty"
)

const ( // [CONVERTER WARNING MESSAGES]
	outTagNotFound = "WARNING: linked output tag not found, input row:"
	outRowEmpty    = "WARNING: converted row is empty, input row:"
)

var ( // [CONVERTING ERRORS]
	ErrInvalidParseMsg   = errors.New("parse input messsge to struct has been unsuccesful")
	ErrOutputTagNotFound = errors.New("linked tags in tag not found in output modification")
)

const ( // [CONVERTING OPTIONS ERRORS]
	ErrUndefinedOption = "undefined option (name=%s) by tag (name=%s)"
)

const ( // [CONVERTING fERRORS]
	ErrUndefinedPositionTag = "tag (name=%s) has position, but it's not found in tags"

	ErrWrongTagRow = "identify msg and split input msg has been unsuccessful. row %s"

	ErrInputTagModificationNotFound = "tag %s not found in input modification %v"

	ErrUndefinedInputTag  = "undefined input tag %s - not found in config"
	ErrUndefinedOutputTag = "undefined output tag %s"

	ErrWrongSliceOfTag     = "'slice of Tag' by tag %s is empty"
	ErrInputTagMSGNotFound = "tag %s not found in search structure MSG %v"

	ErrWrongFieldsNumber  = "tag(name=%s, structure=%v) has differents in current fields number(%d) and specified count(%d)"
	ErrWrongFieldPosition = "wrong position (position must be more than 1) by fieldName %s in %+v"
	ErrWrongFieldLink     = "specified field link is incorrect, field %s"
	ErrWrongFieldMetadata = "fieldName %s in output modification %v hasn't have a linked_fields/default_value is not specified/input_field not found in linked_fields"

	ErrWrongComponentsNumber    = "commponent count can be equal to 0 or more than 1, else it's field hasn't have components. FieldName %s"
	ErrNegativeComponentsNumber = "commponent count can be equal to 0 or more than 1. FieldName %s"
	ErrWrongComponentSplit      = "incorrect field %s, the field component could not be pulled out"

	ErrManyMultiTags = "converter can working with one multi tag not more, wait: %s, receive: %s"
)
