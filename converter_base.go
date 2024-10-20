package hl7converter

import (
	"fmt"
	"sort"
	"bytes"
	"bufio"
	"slices"
	"strings"
)


// IndetifyMsg
// 
// _______[INFO]_______
// - IndetifyMsg indetify by output modification (field: Types) and compare it with Tags in Msg 
//
// -------[NOTES]-------
// - ИЗМЕНИТЬ ИДЕНТИФИКАЦИЮ ( ДОБАВИТЬ АВТО СПЛИТ MSG? ) 
// 
func indetifyMsg(msg *Msg, modification *Modification) (string, bool) {
	actualTags := make([]string, 0, len(msg.Tags))
	for Tag := range msg.Tags {
		actualTags = append(actualTags, string(Tag))
	}
	sort.Strings(actualTags) // we sort in order to compare tags regardless of the positions of the tags

	for TypeName, Tags := range modification.Types {
		for _, someTags := range Tags{
			sort.Strings(someTags)
	
			if slices.Compare(actualTags, someTags) == 0 {
				return TypeName, true
			}
		}
	}

	return "", false
}


// ConvertToMSG
//
// _______[INFO]_______
// - Func return MSG model for get fields data specified in output 'linked_fields'.
// - Using only input modification
//
// -------[NOTES]-------
// - We can do without copying the structure MSG
//
func ConvertToMSG(p ConverterParams , fullMsg []byte) (*Msg, error) {
	tags := make(map[TagName]SliceFields)

	scanner := bufio.NewScanner(bytes.NewReader(fullMsg))
	scanner.Split(p.LineSplit)

	for scanner.Scan() {
		token := scanner.Text() // [DEV] getting string representation of row
		rowFields := strings.Split(token, p.InMod.FieldSeparator)

		tag, fields := rowFields[0], rowFields[1:]
		if _, ok := p.InMod.TagsInfo.Tags[tag]; !ok {
			return nil, NewErrUndefinedInputTag(tag, "ConvertToMSG func")
		}

		processedTag, processedFields := TagName(tag), TagFields(fields)
		
		if _, ok := tags[processedTag]; ok {
			tags[processedTag] = append(tags[processedTag], processedFields)
		} else {
			tags[processedTag] = make(SliceFields, 0, 1) 
			tags[processedTag] = append(tags[processedTag], processedFields)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("convert input messsge to Msg struct has been unsuccesful")
	}

	return &Msg{Tags: tags}, nil
}