package hl7converter

// TagName is a message tag identifier.
type TagName string

// TagFields is a slice of fields for a single tag row.
type TagFields []string

// SliceFields holds multiple occurrences of a tag.
type SliceFields []TagFields

// Msg stores parsed message grouped by tag.
type Msg struct {
	Tags map[TagName]SliceFields
}
