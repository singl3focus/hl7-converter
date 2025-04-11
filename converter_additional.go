package hl7converter

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
)

var (
	ErrInvalidJsonExtension = errors.New("config error: path doesn't contains extension 'json'")

	ErrNilModification = errors.New("config error: modification was incorrectly read from the file because it's empty")
)

// ConverterParams
type ConverterParams struct {
	InMod, OutMod *Modification
}

func NewConverterParams(cfgPath, cfgInBlockName, cfgOutBlockName string) (*ConverterParams, error) {
	if !strings.Contains(cfgPath, ".json") {
		return nil, NewError(ErrInvalidJsonExtension, fmt.Sprintf("path %s", cfgPath))
	}

	inMod, err := ReadJSONConfigBlock(cfgPath, cfgInBlockName)
	if err != nil {
		return nil, err
	}
	if inMod == nil {
		return nil, NewError(ErrNilModification, fmt.Sprintf("modification %s path %s", cfgInBlockName, cfgPath))
	}

	outMod, err := ReadJSONConfigBlock(cfgPath, cfgOutBlockName)
	if err != nil {
		return nil, err
	}
	
	if outMod == nil {
		return nil, NewError(ErrNilModification, fmt.Sprintf("modification %s path %s", cfgOutBlockName, cfgPath))
	}

	return &ConverterParams{
		InMod:  inMod,
		OutMod: outMod,
	}, nil
}

// IndetifyMsg
func IndetifyMsg(p *ConverterParams, msg []byte) (string, error) {
	MSG, err := ConvertToMSG(p, msg)
	if err != nil {
		return "", err
	}

	msgType, ok := identifyMsg(MSG, p.InMod)
	if !ok {
		return "", fmt.Errorf("undefined type, msg: %v", MSG)
	}

	return msgType, nil
}

// identifyMsg indetify by output modification (field: Types) and compare it with Tags in Msg
//
// -------[NOTES]-------
// - ИЗМЕНИТЬ ИДЕНТИФИКАЦИЮ ( ДОБАВИТЬ АВТО СПЛИТ MSG? )
func identifyMsg(msg *Msg, modification *Modification) (string, bool) {
	actualTags := make([]string, 0, len(msg.Tags))
	for t := range msg.Tags {
		actualTags = append(actualTags, string(t))
	}
	sort.Strings(actualTags) // we sort in order to compare tags regardless of the positions of the tags

	for TypeName, Tags := range modification.Types {
		for _, someTags := range Tags {
			sort.Strings(someTags)

			if slices.Compare(actualTags, someTags) == 0 {
				return TypeName, true
			}
		}
	}

	return "", false
}

// ConvertToMSG c return MSG model for get fields data specified in output 'linked_fields'.
// - Using only input modification
//
// -------[NOTES]-------
// - We can do without copying the structure MSG
func ConvertToMSG(p *ConverterParams, fullMsg []byte) (*Msg, error) {
	tags := make(map[TagName]SliceFields)

	scanner := bufio.NewScanner(bytes.NewReader(fullMsg))
	scanner.Split(GetCustomSplit(p.InMod.LineSeparator))

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
