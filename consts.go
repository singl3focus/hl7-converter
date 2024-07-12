package hl7converter

import "errors"

// [SERVICE DATA]
const (
	linkToFieldSeparator = "-"
)

// [SEARCH STRUCTURE ERRORS]
var (
	ErrWrongSliceOfTag = "SliceOfTag by tag %s is empty"
	ErrInputTagMSGNotFound = "tag %s not found in search structure MSG %v"
)


// [CONVERTEING ERRORS]
var (
	ErrWrongTagRow = "identify msg and split input msg has been unsuccessful. row %s"

	ErrInputTagModificationNotFound = "tag %s not found in input modification %v"
	
	ErrOutputTagNotFound = errors.New("linked tags in tag not found in output modification")

	ErrUndefinedInputTag = "undefined input tag %s - not found in config"
	ErrUndefinedOutputTag ="undefined output tag %s"

	ErrWrongComponentsNumber = "commponent count can be equal to 0 or more than 1, else it's field hasn't have components. FieldName %s"
	ErrNegativeComponentsNumber = "commponent count can be equal to 0 or more than 1. FieldName %s"

	ErrWrongFieldPosition = "wrong position %d (position must be more than 1) by fieldName %s"


	ErrManyMultiTags = "converter can working with one multi tag not more, wait: %s, receive: %s"
)
