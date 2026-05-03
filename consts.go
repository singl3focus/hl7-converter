package hl7converter

// APPLICATION

// Deprecated: pass an explicit config path to NewConverterParams instead of
// depending on a library-provided sample file location. The path is preserved
// only for backward compatibility and points to the relocated sample fixture.
const CfgJSON = "examples/config.json"

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
