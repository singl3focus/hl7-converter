package hl7converter

/*
	For parsing input message
*/

type TagName string
type TagFields []string
type SliceFields []TagFields

type Msg struct {
	Tags map[TagName]SliceFields
}
