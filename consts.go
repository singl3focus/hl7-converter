package hl7converter

const (
	linkToFieldSeparator = "-"
)

// [ERRORS]
var (
	ErrUndefinedInputTag = "unndefined input tag %s"
	ErrUndefinedOutputTag ="unndefined output tag %s"

	ErrWrongComponentsNumber = "Commponent count can be equal to 0 or more than 1, else it's field hasn't have components. FieldName %s"
	ErrNegativeComponentsNumber = "Commponent count can be equal to 0 or more than 1. FieldName %s"
	
	ErrWrongFieldPosition = "incorrect position %d (position must be more than 1) by fieldName %s"
)
