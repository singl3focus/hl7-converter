package hl7converter

import (
	"fmt"
	"sort"
	"bytes"
	"bufio"
	"slices"
	"strings"
)

type BaseConverter struct {

}

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
	tags := make(map[TagName]SliceOfTag)

	scanner := bufio.NewScanner(bytes.NewReader(fullMsg))
	scanner.Split(p.lineSplit)

	for scanner.Scan() {
		temp := make(TagValues)

		token := scanner.Text() // [DEV] getting string representation of row
		rowFields := strings.Split(token, p.inMod.FieldSeparator)

		tag, fields := rowFields[0], rowFields[1:]
		if _, ok := p.inMod.Tags[tag]; !ok {
			return nil, fmt.Errorf(ErrUndefinedInputTag, tag)
		}

		processedTag, processedFields := TagName(tag), Fields(fields)
		
		temp[processedTag] = processedFields

		if _, ok := tags[processedTag]; ok {
			tags[processedTag] = append(tags[processedTag], temp)

		} else {
			tags[processedTag] = make(SliceOfTag, 0, 10) // [MAGIC] note - capacity is 10 because it's optimal value, which describe average rows of message
			tags[processedTag] = append(tags[processedTag], temp)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("convert input messsge to Msg struct has been unsuccesful")
	}

	return &Msg{Tags: tags}, nil
}