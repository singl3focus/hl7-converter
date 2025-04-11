package hl7converter

// APPLICATION

const (
	CfgJSON       = "config.json"
	CfgSchemaJSON = "config.schema.json"
)

// CONVERTING PARAMS

const ( 
	ignoredIndx = 0
	ignoredFieldsNumber = -1
)

// CONFIG PARAMS
const (
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
