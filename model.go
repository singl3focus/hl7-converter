package hl7converter

// For parsing metadata of config

type Modification struct {
	ComponentSeparator string `json:"Component_separator"`
	FieldSeparator     string `json:"Field_separator"`
	LineSeparator      string `json:"Line_separator"`

	Types map[string][]string `json:"Types"`

	Tags map[string]Tag `json:"Tags"`
}

type Tag struct {
	Position     int `json:"position"`
	FieldsNumber int `json:"fields_number"`

	Fields map[string]Field `json:"fields"`
}

type Field struct {
	DefaultValue     string   `json:"default_value,omitempty"`    // OPTIONAL
	Position         float64  `json:"position"`                   // MANDATORY
	Linked           []string `json:"linked_fields,omitempty"`    // OPTIONAL
	ComponentsNumber int      `json:"components_count,omitempty"` // OPTIONAL
}

// For parsing input message

type TagName string
type Fields []string

type TagValues map[TagName]Fields

type SliceOfTag []TagValues

type Msg struct {
	Tags map[TagName]SliceOfTag
}
