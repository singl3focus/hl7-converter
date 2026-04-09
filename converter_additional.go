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
	ErrInvalidJsonExtension = errors.New("convert params error: path does not contains extension 'json'")

	ErrUndefinedInputTag = errors.New("convert to msg error: undefined input tag")
)

func NewErrUndefinedInputTag(tag, someinfo string) error {
	return &Error{
		Err:            ErrUndefinedInputTag,
		AdditionalInfo: fmt.Sprintf("tag %s additional info %s", tag, someinfo),
	}
}

// ConverterParams carries parsed input/output modifications used to build Converter.
type ConverterParams struct {
	InputModification  *Modification
	OutputModification *Modification
}

// NewConverterParams reads config file and returns params for specified input/output blocks.
// Config path must end with JsonExtension.
func NewConverterParams(cfgPath, cfgInBlockName, cfgOutBlockName string) (*ConverterParams, error) {
	if !strings.Contains(cfgPath, JsonExtension) {
		return nil, NewError(ErrInvalidJsonExtension, true, fmt.Sprintf("path %s", cfgPath))
	}

	inMod, err := ReadJSONConfigBlock(cfgPath, cfgInBlockName)
	if err != nil {
		return nil, err
	}

	outMod, err := ReadJSONConfigBlock(cfgPath, cfgOutBlockName)
	if err != nil {
		return nil, err
	}

	if err := validateConversionPair(inMod, outMod); err != nil {
		return nil, err
	}

	return &ConverterParams{
		InputModification:  inMod,
		OutputModification: outMod,
	}, nil
}

// IndetifyMsg detects message type using Types of input modification and actual parsed tags.
func IndetifyMsg(p *ConverterParams, msg []byte) (string, error) {
	MSG, err := ConvertToMsg(p, msg)
	if err != nil {
		return "", err
	}

	msgType, ok := identifyMsg(MSG, p.InputModification)
	if !ok {
		return "", fmt.Errorf("undefined type, msg: %v", MSG)
	}

	return msgType, nil
}

// identifyMsg compares actual tags with configured Types.
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

// ConvertToMsg returns parsed message into Msg model using input modification separators.
func ConvertToMsg(p *ConverterParams, fullMsg []byte) (*Msg, error) {
	tags := make(map[TagName]SliceFields)

	scanner := bufio.NewScanner(bytes.NewReader(fullMsg))
	scanner.Split(GetCustomSplit(p.InputModification.LineSeparator))

	for scanner.Scan() {
		token := scanner.Text() // getting string representation of row
		rowFields := strings.Split(token, p.InputModification.FieldSeparator)
		if len(rowFields) < 2 {
			return nil, NewError(ErrParseFailure, true,
				fmt.Sprintf("line: %s, splitter: %s, fields count less than 2", token, p.InputModification.FieldSeparator))
		}

		tag, fields := rowFields[0], rowFields[1:]
		if _, ok := p.InputModification.TagsInfo.Tags[tag]; !ok {
			return nil, NewError(ErrUndefinedInputTag, true, fmt.Sprintf("tag %s", tag))
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
		return nil, NewError(ErrUndefinedScannerFailure, true)
	}

	return &Msg{Tags: tags}, nil
}
