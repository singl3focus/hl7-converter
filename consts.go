package hl7converter

// APPLICATION

// Deprecated: pass an explicit config path to NewConverterParams instead of
// depending on a library-provided sample file location.
const CfgJSON = "config.json"

const CfgSchemaJSON = "config.schema.json"

// CONVERTING PARAMS

const (
	ignoredIndx         = 0
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
