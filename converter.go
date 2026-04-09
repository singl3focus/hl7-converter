package hl7converter

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// Package-level converting errors that are part of the public API.
var ( // TODO: move errors to origin place
	FatalErrOfConverting     = errors.New("convert error: parse input message to struct has been unsuccesful")
	ErrInvalidParseMsg       = errors.New("convert error: parse input message to struct has been unsuccesful")
	ErrInputTagNotFound      = errors.New("convert error: input tag not found")
	ErrOutputTagNotFound     = errors.New("convert error: linked tags in tag not found in output modification") // todo: clarify - are we can have linked one tag or many
	ErrUndefinedOption       = errors.New("convert error: undefined option by tag")
	ErrUndefinedPositionTag  = errors.New("convert error: tag has position, but it's not found in tags")
	ErrInvalidLink           = errors.New("convert error: invalid link")
	ErrInvalidLinkElems      = errors.New("convert error: invalid link, some elems empty")
	ErrWrongParamCount       = errors.New("convert error: you can use only one special symbol in field")
	ErrEmptyDefaultValue     = errors.New("convert error: field has empty default value")
	ErrTooBigIndex           = errors.New("convert error: index of output rows more than max index of input rows")
	ErrWrongFieldsNumber     = errors.New("convert error: tag has invalid specified count of fields number")
	ErrWrongComponentsNumber = errors.New("convert error: component not found but link is exist")
	ErrWrongComponentLink    = errors.New("convert error: link component position more than max components count of input row with tag")
)

func NewFatalErrOfConverting(r any) error {
	return &Error{
		Err:            FatalErrOfConverting,
		AdditionalInfo: fmt.Sprintf("recovered %+v", r),
	}
}

func NewErrInvalidParseMsg(err string) error {
	return &Error{
		Err:            ErrInvalidParseMsg,
		AdditionalInfo: fmt.Sprintf("system error %s", err),
	}
}

func NewErrInputTagNotFound(row string) error {
	return &Error{
		Err:            ErrInputTagNotFound,
		AdditionalInfo: fmt.Sprintf("row %s", row),
	}
}

func NewErrOutputTagNotFound(tag string) error {
	return &Error{
		Err:            ErrOutputTagNotFound,
		AdditionalInfo: fmt.Sprintf("input tag %s", tag),
	}
}

func NewErrUndefinedOption(option, tag string) error {
	return &Error{
		Err:            ErrUndefinedOption,
		AdditionalInfo: fmt.Sprintf("option %s tag %s, available %s", option, tag, supportedOptionsSummary()),
	}
}

func NewErrUndefinedPositionTag(tag string) error {
	return &Error{
		Err:            ErrUndefinedPositionTag,
		AdditionalInfo: fmt.Sprintf("tag %s", tag),
	}
}

func NewErrInvalidLink(link string) error {
	return &Error{
		Err:            ErrInvalidLink,
		AdditionalInfo: fmt.Sprintf("link %s", link),
	}
}

func NewErrInvalidLinkElems(link string) error {
	return &Error{
		Err:            ErrInvalidLinkElems,
		AdditionalInfo: fmt.Sprintf("link %s", link),
	}
}

func NewErrWrongParamCount(field, param string) error {
	return &Error{
		Err:            ErrWrongParamCount,
		AdditionalInfo: fmt.Sprintf("field %s param %s", field, param),
	}
}

func NewErrEmptyDefaultValue(field string) error {
	return &Error{
		Err:            ErrEmptyDefaultValue,
		AdditionalInfo: fmt.Sprintf("field %s", field),
	}
}

func NewErrTooBigIndex(idx, maxIdx int) error {
	return &Error{
		Err:            ErrTooBigIndex,
		AdditionalInfo: fmt.Sprintf("index %d maxIndex %d", idx, maxIdx),
	}
}

func NewErrWrongFieldsNumber(tag string, tagstruct *Tag, currentFieldsNumb int) error {
	return &Error{
		Err:            ErrWrongFieldsNumber,
		AdditionalInfo: fmt.Sprintf("tagName %s tagSturcture %+v current fields has %d", tag, tagstruct, currentFieldsNumb),
	}
}

func NewErrWrongComponentsNumber(inputdata, link string) error {
	return &Error{
		Err:            ErrWrongComponentsNumber,
		AdditionalInfo: fmt.Sprintf("line %s link %s", inputdata, link),
	}
}

func NewErrWrongComponentLink(link string, compPos, compCount int, inputTag string) error {
	return &Error{
		Err:            ErrWrongComponentLink,
		AdditionalInfo: fmt.Sprintf("tag %s link %s componentPosition %d componentCount %d", inputTag, link, compPos, compCount),
	}
}

// Converter transforms input message according to config modifications.
// A single Converter instance is safe to reuse across concurrent Convert calls.
type Converter struct {
	// Data parsed from config.
	// Goal: find metadata(position, default_value, components_number, linked and data about Tags, Separators)
	// about field so that then get value some field.
	Input, Output *Modification

	// For effective split by rows of input message with help "bufio.Scanner"
	LineSplit bufio.SplitFunc

	UsingPositions bool
	UsingAliases   bool
}

// OptionFunc configures Converter construction.
type OptionFunc func(*Converter)

// WithUsingPositions enables positional output generation using Output.TagsInfo.Positions.
func WithUsingPositions() OptionFunc {
	return func(n *Converter) {
		n.UsingPositions = true
	}
}

// WithUsingAliases applies aliases after conversion using Output.Aliases.
func WithUsingAliases() OptionFunc {
	return func(n *Converter) {
		n.UsingAliases = true
	}
}

// NewConverter builds Converter from params and options.
// Input/Output modifications and separators are taken from params; Converter is ready for one-shot or repeated serial use.
func NewConverter(p *ConverterParams, opts ...OptionFunc) (*Converter, error) {
	converter := &Converter{
		Input:     p.InputModification,
		Output:    p.OutputModification,
		LineSplit: GetCustomSplit(p.InputModification.LineSeparator),
	}

	for _, opt := range opts {
		opt(converter)
	}

	return converter, nil
}

type convertCallState struct {
	msgSource      *Msg
	tagPointerByIn map[string]int
	currentIndex   int
	outputTagCount map[string]int
}

func newConvertCallState(tags map[TagName]SliceFields) *convertCallState {
	return &convertCallState{
		msgSource:      &Msg{Tags: tags},
		tagPointerByIn: make(map[string]int),
		outputTagCount: make(map[string]int),
	}
}

func (s *convertCallState) nextInputIndex(tag string) int {
	idx := s.tagPointerByIn[tag]
	s.tagPointerByIn[tag] = idx + 1

	return idx
}

func (s *convertCallState) countForOutputTag(outputTag string, output *Modification) int {
	if count, ok := s.outputTagCount[outputTag]; ok {
		return count
	}

	linkedTag := output.TagsInfo.Tags[outputTag].Linked
	count := 1
	if linkedRows, ok := s.msgSource.Tags[TagName(linkedTag)]; ok && len(linkedRows) > 0 {
		count = len(linkedRows)
	}

	s.outputTagCount[outputTag] = count

	return count
}

/*_______________________________________[PARSE MSG AND EXECUTE OPRIONS SPECIFIED IN config]_______________________________________*/

// GetCustomSplit returns a bufio.SplitFunc that splits by a custom line separator.
func GetCustomSplit(sep string) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if i := bytes.Index(data, []byte(sep)); i >= 0 {
			return i + len(sep), data[0:i], nil
		}

		if atEOF {
			return len(data), data, nil
		}

		return 0, nil, nil
	}
}

// ParseMsg returns parsed tags from input message according to Input modification.
// It validates tags against Input.TagsInfo.Tags and applies tag options.
func (c *Converter) ParseMsg(fullMsg []byte) (map[TagName]SliceFields, error) {
	tags := make(map[TagName]SliceFields)

	scanner := bufio.NewScanner(bytes.NewReader(fullMsg))
	scanner.Split(c.LineSplit)

	for scanner.Scan() {
		token := scanner.Text() // [DEV] getting string representation of row
		rowFields := strings.Split(token, c.Input.FieldSeparator)

		tag, fields := rowFields[0], rowFields[1:]
		if _, ok := c.Input.TagsInfo.Tags[tag]; !ok {
			return nil, NewErrUndefinedInputTag(tag, "ParseMsg func")
		}

		tempTag, tempFields, err := c.handleOptions(tag, fields)
		if err != nil {
			return nil, err
		}
		processedTag, processedFields := TagName(tempTag), TagFields(tempFields)

		if _, ok := tags[processedTag]; ok {
			tags[processedTag] = append(tags[processedTag], processedFields)
		} else {
			tags[processedTag] = make(SliceFields, 0, 1)
			tags[processedTag] = append(tags[processedTag], processedFields)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, NewErrInvalidParseMsg(err.Error())
	}

	return tags, nil
}

// handleOptions
func (c *Converter) handleOptions(tag string, fields []string) (string, []string, error) {
	newFields, err := applyTagOptions(tag, fields, c.Input.TagsInfo.Tags[tag])
	if err != nil {
		return tag, nil, err
	}

	return tag, newFields, nil
}

/*_______________________________________[GENERAL CONVERT]_______________________________________*/

// Convert executes conversion from input message to Result according to Output modification.
// One Converter instance may be used by multiple goroutines concurrently.
func (c *Converter) Convert(fullMsg []byte) (result *Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = NewFatalErrOfConverting(r)
		}
	}()

	tags, err := c.ParseMsg(fullMsg)
	if err != nil {
		return nil, err
	}
	state := newConvertCallState(tags)

	if c.UsingPositions {
		result, err = c.convertWithPositions(state)
	} else {
		result, err = c.convertByInput(fullMsg, state)
	}
	if err != nil {
		return nil, err
	}

	if c.UsingAliases {
		aliases := make(Aliases, len(c.Output.Aliases))
		for name, link := range c.Output.Aliases {
			aliases[name] = link
		}

		if err = result.ApplyAliases(aliases); err != nil {
			return nil, err
		}
	}

	return result, err
}

func (c *Converter) convertByInput(fullMsg []byte, state *convertCallState) (*Result, error) {
	scanner := bufio.NewScanner(bytes.NewReader(fullMsg))
	scanner.Split(c.LineSplit)

	rows := make([]*Row, 0, 1)
	for scanner.Scan() {
		inputRow := scanner.Text()

		var inputTag string
		for i, ch := range inputRow {
			if string(ch) == c.Input.FieldSeparator {
				inputTag = inputRow[:i]
				if inputTag == "" {
					return nil, NewErrInputTagNotFound(inputRow)
				}
				break
			}
		}

		if _, exist := c.Input.TagsInfo.Tags[inputTag]; !exist {
			return nil, NewErrUndefinedInputTag(inputTag, "not found in input modification")
		}

		outputTag := c.Input.TagsInfo.Tags[inputTag].Linked
		outputTagInfo, exist := c.Output.TagsInfo.Tags[outputTag]
		if !exist {
			return nil, NewErrOutputTagNotFound(inputTag)
		}

		state.currentIndex = state.nextInputIndex(inputTag)

		row, err := c.convertTag(state, outputTag, &outputTagInfo)
		if err != nil {
			return nil, err
		}

		rows = append(rows, NewRow(c.Output.FieldSeparator, row))
	}

	return NewResult(c.Output.LineSeparator, rows), nil
}

// convertWithPositions generates output rows strictly by Output.TagsInfo.Positions.
func (c *Converter) convertWithPositions(state *convertCallState) (*Result, error) {
	orderedTags, err := c.Output.OrderedPositionTags()
	if err != nil {
		return nil, err
	}

	rows := make([]*Row, 0, 1)
	for _, tag := range orderedTags {
		TagInfo, ok := c.Output.TagsInfo.Tags[tag]
		if !ok {
			return nil, NewErrUndefinedPositionTag(tag)
		}

		for i := 0; i < state.countForOutputTag(tag, c.Output); i++ {
			state.currentIndex = i

			row, err := c.convertTag(state, tag, &TagInfo)
			if err != nil {
				return nil, err
			}

			rows = append(rows, NewRow(c.Output.FieldSeparator, row))
		}
	}

	return NewResult(c.Output.LineSeparator, rows), nil
}

/*_______________________________________[CONVERT SPECIFIED TAG]_______________________________________*/

// convertTag
func (c *Converter) convertTag(state *convertCallState, outputTag string, outputTagInfo *Tag) ([]*Field, error) {
	row := strings.Split(outputTagInfo.Tempalate, c.Output.FieldSeparator)

	if outputTagInfo.FieldsNumber != ignoredFieldsNumber {
		if len(row) != outputTagInfo.FieldsNumber || outputTagInfo.FieldsNumber < 1 {
			return nil, NewErrWrongFieldsNumber(outputTag, outputTagInfo, len(row))
		}
	}

	outputLine, err := c.assembleOutRow(state, outputTagInfo, row)
	if err != nil {
		return nil, err
	}

	return outputLine, nil
}

// assembleOutRow creates out line and fills it. Also inserts component separators in fields with components.
func (c *Converter) assembleOutRow(state *convertCallState, inTagInfo *Tag, rowData []string) ([]*Field, error) {
	line := make([]*Field, inTagInfo.FieldsNumber) // temp slice was initially filled by default value:""
	for i := range line {
		line[i] = &Field{} // Init empty values for line
	}

	// first position is always placed by Tag
	line[0] = NewField(rowData[ignoredIndx], c.Output.ComponentSeparator, c.Output.ComponentArrSeparator)

	// [DEV] fieldPosition started from '0' not from 'ignoredIndx+1'
	for fieldPosition, fieldValue := range rowData[ignoredIndx+1:] {
		if fieldValue == "" {
			continue
		}

		// it's not inc global var, just represent because we started from 'ignoredIndx+1'
		fieldPosition++

		fieldBlocks := strings.Split(fieldValue, OR)
		switch len(fieldBlocks) {
		case 1: // WITHOUT OPPORTUNITY SET DEFAULT_VALUE
			mask, err := TempalateParse(fieldBlocks[0])
			if err != nil {
				return nil, err
			}

			value, err := c.getFieldValue(state, mask, fieldBlocks[0])
			if err != nil {
				return nil, err
			}

			line[fieldPosition] = NewField(value, c.Output.ComponentSeparator, c.Output.ComponentArrSeparator)
		case 2: // MUST BE TEMPLATE OR DEFAULT_VALUE
			if fieldBlocks[0] != "" {
				mask, err := TempalateParse(fieldBlocks[0])
				if err != nil {
					return nil, err
				}

				value, err := c.getFieldValue(state, mask, fieldBlocks[0])
				if err != nil {
					return nil, err
				}

				line[fieldPosition] = NewField(value, c.Output.ComponentSeparator, c.Output.ComponentArrSeparator)
			} else {
				value, err := c.getDefaultFieldValue(fieldBlocks[1])
				if err != nil {
					return nil, err
				}

				line[fieldPosition] = NewField(value, c.Output.ComponentSeparator, c.Output.ComponentArrSeparator)
			}
		default:
			return nil, NewErrWrongParamCount(fieldValue, OR)
		}
	}

	return line, nil
}

// getFieldValue is distributing of fields by linked fields number
func (c *Converter) getFieldValue(state *convertCallState, mask []int, str string) (string, error) {
	var builder strings.Builder

	for i := 0; i < len(str); i++ {
		if mask[i] == itSymbol {
			builder.WriteByte(str[i])
		} else if mask[i] == itLink {
			var link string // parse: Tag-Position (Without '<', '>')

			for j := i; j < len(str); j++ {
				if mask[j] == itSymbol {
					link = str[i+1 : j-1]
					i = j - 1 // subtract '1' because next step will be increment
					break
				} else if j == len(str)-1 {
					link = str[i+1 : j]
					i = j // cycle must be ended
					break
				}
			}

			if link == "" {
				return "", NewErrInvalidLink(str)
			}

			value, err := c.getValueFromMSGbyLink(state, link)
			if err != nil {
				return "", err
			}

			builder.WriteString(value)
		}
	}

	return builder.String(), nil
}

// getDefaultFieldValue
func (c *Converter) getDefaultFieldValue(str string) (string, error) {
	if str == "" {
		return "", NewErrEmptyDefaultValue(str)
	}

	return str, nil
}

func (c *Converter) getValueFromMSGbyLink(state *convertCallState, link string) (string, error) {
	var value string

	ref, err := parseLinkRef(link)
	if err != nil {
		return "", err
	}

	inTagInfo, ok := state.msgSource.Tags[TagName(ref.Tag)]
	if !ok {
		return "", NewErrUndefinedInputTag(ref.Tag, link)
	}

	if state.currentIndex > (len(inTagInfo) - 1) {
		return "", NewErrTooBigIndex(state.currentIndex, len(inTagInfo)-1)
	}

	fieldIndex, err := fieldIndexFromPosition(ref.Raw, ref.PositionValue)
	if err != nil {
		return "", err
	}
	if fieldIndex < 0 || fieldIndex >= len(inTagInfo[state.currentIndex]) {
		return "", NewErrInvalidLink(link)
	}

	if isInt(ref.PositionValue) {
		value = inTagInfo[state.currentIndex][fieldIndex]
	} else {
		componentPosIndx, err := componentIndexFromPosition(ref.Raw, ref.PositionValue)
		if err != nil {
			return "", err
		}

		fieldValue := inTagInfo[state.currentIndex][fieldIndex]

		components := strings.Split(fieldValue, c.Input.ComponentSeparator)
		if len(components) == 1 {
			return "", NewErrWrongComponentsNumber(fieldValue, link)
		}

		if componentPosIndx > (len(components) - 1) {
			return "", NewErrWrongComponentLink(link, componentPosIndx+1, len(components), ref.Tag)
		}

		value = components[componentPosIndx]
	}

	return value, nil
}

/*_______________________________________[PARSE INPUT TO RESULT]_______________________________________*/

func (c *Converter) ParseInput(msg []byte) (*Result, error) {
	modification := c.Input

	scanner := bufio.NewScanner(bytes.NewReader(msg))
	scanner.Split(c.LineSplit)

	rows := make([]*Row, 0, 1)
	for scanner.Scan() {
		inputRow := strings.Split(scanner.Text(), modification.FieldSeparator)

		fields := make([]*Field, 0, 1)
		for _, f := range inputRow {
			fields = append(fields, NewField(f, modification.ComponentSeparator, modification.ComponentArrSeparator))
		}

		rows = append(rows, NewRow(modification.FieldSeparator, fields))
	}

	return NewResult(modification.LineSeparator, rows), nil
}
