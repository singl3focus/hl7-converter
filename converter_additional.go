package hl7converter

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
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
	if filepath.Ext(cfgPath) != JsonExtension {
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
		return nil, fmt.Errorf("%w: %w", ErrInvalidConfig, err)
	}

	return &ConverterParams{
		InputModification:  inMod,
		OutputModification: outMod,
	}, nil
}

// IdentifyMsg identifies a message without changing the shared input configuration.
func IdentifyMsg(p *ConverterParams, msg []byte) (string, error) {
	parsed, err := ConvertToMsg(p, msg)
	if err != nil {
		return "", err
	}
	actual := make([]string, 0, len(parsed.Tags))
	for tag := range parsed.Tags {
		actual = append(actual, string(tag))
	}
	sort.Strings(actual)
	match := ""
	for _, name := range sortedKeys(p.InputModification.Types) {
		for _, alternative := range p.InputModification.Types[name] {
			tags := append([]string(nil), alternative...)
			sort.Strings(tags)
			if !equalStrings(actual, tags) {
				continue
			}
			if match != "" && match != name {
				return "", fmt.Errorf("%w: %q and %q", ErrAmbiguousMessageType, match, name)
			}
			match = name
		}
	}
	if match == "" {
		return "", ErrMessageTypeNotFound
	}
	return match, nil
}

// IndetifyMsg is the legacy spelling of IdentifyMsg.
// Deprecated: use IdentifyMsg.
func IndetifyMsg(p *ConverterParams, msg []byte) (string, error) { return IdentifyMsg(p, msg) }

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ConvertToMsg returns parsed message into Msg model using input modification separators.
func ConvertToMsg(p *ConverterParams, fullMsg []byte) (*Msg, error) {
	if p == nil || p.InputModification == nil {
		return nil, fmt.Errorf("%w: nil input modification", ErrInvalidConfig)
	}
	if err := validateSeparators(p.InputModification); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidConfig, err)
	}
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
