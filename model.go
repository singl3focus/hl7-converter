package hl7converter

type Modification struct {
	ComponentSeparator string `json:"Component_separator"`
	FieldSeparator 	   string `json:"Field_separator"`
	LineSeparator 	   string `json:"Line_separator"`

	Tags map[string]Tag `json:"Tags"`
}

type Tag struct {
	Linked 		 []string `json:"linked"`
	FieldsNumber int `json:"fields_number"`

	Fields 		 map[string]Field `json:"fields"`
}

type Field struct {
	DefaultValue string `json:"default_value,omitempty"` 		// OPTIONAL 
	Position 	 float64 `json:"position"` 						// MANDATORY
	Linked 		 []string `json:"linked_fields,omitempty"` 		// OPTIONAL 
	ComponentsNumber int `json:"components_count,omitempty"` 	// OPTIONAL 
}

// _________________________________________

type Msg struct {
	Tags map[string][]string
}