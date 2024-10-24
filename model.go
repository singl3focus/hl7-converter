package hl7converter

/*
	For parsing metadata of config
*/

type Modification struct {
	ComponentSeparator    string `json:"component_separator"`
	ComponentArrSeparator string `json:"component_array_separator"`
	FieldSeparator        string `json:"field_separator"`
	LineSeparator         string `json:"line_separator"`

	Types map[string][][]string `json:"types,omitempty"` // [OPTIONAL]

	TagsInfo TagsInfo `json:"tags_info"`
}

type TagsInfo struct {
	Positions map[string]string `json:"positions"`
	Tags      map[string]Tag    `json:"tags"`
}

type Tag struct {
	Count        int    `json:"-"`
	Linked       string `json:"linked"`
	FieldsNumber int    `json:"fields_number"`
	Tempalate    string `json:"template"`

	Options []string `json:"options,omitempty"` // [OPTIONAL]
}

/*
	For parsing input message
*/

type TagName string
type TagFields []string
type SliceFields []TagFields

type Msg struct {
	Tags map[TagName]SliceFields
}
