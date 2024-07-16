package hl7converter

import "errors"

// [SERVICE DATA]
const (
	linkToFieldSeparator = "-"
)

// [SERVICE MESSAGE]
const (
	outTag404 = "BE CAREFUL: linked output tag not found, input row:"
	outRowEmpty = "WARNING: converted row is empty, input row: %s"
)


// [CONVERTEING ERRORS]
var (
	ErrOutputTagNotFound = errors.New("linked tags in tag not found in output modification")
)


// [CONVERTEING ERRORS with Format]
const (
	ErrWrongTagRow = "identify msg and split input msg has been unsuccessful. row %s"

	ErrInputTagModificationNotFound = "tag %s not found in input modification %v"	

	ErrUndefinedInputTag = "undefined input tag %s - not found in config"
	ErrUndefinedOutputTag ="undefined output tag %s"


	ErrWrongSliceOfTag = "SliceOfTag by tag %s is empty"
	ErrInputTagMSGNotFound = "tag %s not found in search structure MSG %v"

	ErrWrongFieldPosition = "wrong position (position must be more than 1) by fieldName %s in %+v"
	ErrWrongFieldLink = "specified field link is incorrect, field %s"
	ErrWrongFieldMetadata = "fieldName %s in output modification %v hasn't have a linked_fields/default_value is not specified/input_field not found in linked_fields"

	ErrWrongComponentsNumber = "commponent count can be equal to 0 or more than 1, else it's field hasn't have components. FieldName %s"
	ErrNegativeComponentsNumber = "commponent count can be equal to 0 or more than 1. FieldName %s"
	ErrWrongComponentSplit = "incorrect field %s, the field component could not be pulled out"

	ErrManyMultiTags = "converter can working with one multi tag not more, wait: %s, receive: %s"
)
