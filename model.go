package hl7converter

// For parsing metadata of config

type Modification struct {
	ComponentSeparator string `json:"component_separator" yaml:"component_separator"`
	FieldSeparator     string `json:"field_separator" yaml:"field_separator"`
	LineSeparator      string `json:"line_separator" yaml:"line_separator"`

	Types map[string][][]string `json:"types,omitempty" yaml:"types"`

	Tags map[string]Tag `json:"tags" yaml:"tags"`
}

type Tag struct {
	Linked       []string `json:"linked,omitempty" yaml:"linked"`
	Options      []string `json:"options,omitempty" yaml:"options"`
	FieldsNumber int      `json:"fields_number" yaml:"fields_number"`

	Fields map[string]Field `json:"fields" yaml:"fields"`
}

type Field struct {
	DefaultValue     string   `json:"default_value,omitempty" yaml:"default_value"`       // OPTIONAL
	Position         float64  `json:"position" yaml:"position"`                           // MANDATORY
	Linked           []string `json:"linked_fields,omitempty" yaml:"linked_fields"`       // OPTIONAL
	ComponentsNumber int      `json:"components_count,omitempty" yaml:"components_count"` // OPTIONAL
}

// For parsing input message

type TagName string
type Fields []string
type TagValues map[TagName]Fields

type SliceOfTag []TagValues

type Msg struct {
	Tags map[TagName]SliceOfTag
}
