package hl7converter

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// Converter
type Converter struct {
	/*
		Data parsed from config.
		Goal: find metadata(position, default_value, components_number, linked and data about Tags, Separators)
		about field so that then get value some field.
	*/
	Input, Output *Modification

	// For effective split by rows of input message with help "bufio.Scanner"
	LineSplit func(data []byte, atEOF bool) (advance int, token []byte, err error)

	// Convertred structure of input message for fast find a needed field by tag
	MsgSource *Msg
}

func NewConverter(in, out *Modification) (*Converter, error) {
	msg := &Msg{
		Tags: make(map[TagName]SliceOfTag),
	}

	converter := &Converter{
		Input:  in,
		Output: out,

		LineSplit: GetCustomSplit(in.LineSeparator),

		MsgSource: msg,
	}

	return converter, nil
}

var (
	defaultPointerToIndexInMultiTag = 0
	defaultPointerToTag             = ""

	pointerToIndexInMultiTag = defaultPointerToIndexInMultiTag
	pointerToTag             = defaultPointerToTag

	PointerIndx = 0
)

func GetCustomSplit(sep string) func(data []byte, atEOF bool) (advance int, token []byte, err error) {
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

func (c *Converter) handleOptions(tag string, fields []string) (string, []string) {
	options := c.Input.TagsInfo.Tags[tag].Options

	newFields := fields

	for _, option := range options {
		switch option {

		// example: FN - 31(1 - Tag), len(fields) - 28/ That we need add 2 empty fields
		case "autofill":
			diff := (c.Input.TagsInfo.Tags[tag].FieldsNumber - 1) - len(fields)
			for i := 0; i < diff; i++ {
				newFields = append(newFields, "")
			}

		default:
			panic(fmt.Sprintf(ErrUndefinedOption, option, tag))
		}
	}

	return tag, newFields
}

// ParseMsg return 'map[TagName]SliceOfTag' model for get fields data specified in output modification.
func (c *Converter) ParseMsg(fullMsg []byte) (map[TagName]SliceOfTag, error) {
	tags := make(map[TagName]SliceOfTag)

	scanner := bufio.NewScanner(bytes.NewReader(fullMsg))
	scanner.Split(c.LineSplit)

	for scanner.Scan() {
		temp := make(TagValues)

		token := scanner.Text() // [DEV] getting string representation of row
		rowFields := strings.Split(token, c.Input.FieldSeparator)

		tag, fields := rowFields[0], rowFields[1:]
		if _, ok := c.Input.TagsInfo.Tags[tag]; !ok {
			return nil, fmt.Errorf(ErrUndefinedInputTag, tag)
		}

		tempTag, tempFields := c.handleOptions(tag, fields) // [TODO] UPGRADE OPTIONS
		processedTag, processedFields := TagName(tempTag), Fields(tempFields)

		temp[processedTag] = processedFields

		if _, ok := tags[processedTag]; ok {
			tags[processedTag] = append(tags[processedTag], temp)

		} else {
			tags[processedTag] = make(SliceOfTag, 0, 1)
			tags[processedTag] = append(tags[processedTag], temp)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, ErrInvalidParseMsg
	}

	return tags, nil
}

func (c *Converter) Convert(fullMsg []byte) ([][]string, error) {

	// _________fill MsgSource in Converter structure
	tags, err := c.ParseMsg(fullMsg)
	if err != nil {
		return nil, err
	}
	c.MsgSource.Tags = tags

	// _________prepare 'Ordered Slice of Tags' for convert
	tempNumbers := make([]string, 0, 1)
	for key, _ := range c.Output.TagsInfo.Positions {
		tempNumbers = append(tempNumbers, key)
	}
	sort.Strings(tempNumbers)

	OutputTags := make([]string, 0, len(tempNumbers))
	for _, key := range tempNumbers {
		OutputTags = append(OutputTags, c.Output.TagsInfo.Positions[key])
	}

	return c.convert(OutputTags)
}

func (c *Converter) convert(orderedTags []string) ([][]string, error) {
	defer c.ResetParams()

	splitedRows := make([][]string, 0, 1)
	for _, tag := range orderedTags {
		TagInfo, ok := c.Output.TagsInfo.Tags[tag]
		if !ok {
			return nil, fmt.Errorf(ErrUndefinedPositionTag, tag)
		}

		for i := 1; i <= TagInfo.Count; i++ {
			PointerIndx = i - 1

			row, err := c.convertTag(tag, &TagInfo)
			if err != nil {
				return nil, err
			}

			splitedRows = append(splitedRows, row)
		}
	}

	return splitedRows, nil
}

// convertTag
func (c *Converter) convertTag(TagName string, TagInfo *Tag) ([]string, error) {
	row := strings.Split(TagInfo.Tempalate, c.Output.FieldSeparator) // REPEAT BLOCK OF SPLITS

	if len(row) != TagInfo.FieldsNumber {
		return nil,
			fmt.Errorf(ErrWrongFieldsNumber, TagName, TagInfo, len(row), c.Output.TagsInfo.Tags[TagName].FieldsNumber)
	}

	outputLine, err := c.assembleOutRow(TagName, TagInfo.Linked, TagInfo, row)
	if err != nil {
		return nil, err
	}

	// // IMPORTANT ELEMENT TO MOVE POINTER FOR MULTI TAG
	// err := c.MovePointerIndx(tag) //InputTag
	// if err != nil {
	// 	return nil, err
	// }

	return outputLine, nil
}

// assembleOutRow creates outLine and fills it. Also inserts component separators in fields with components.
func (c *Converter) assembleOutRow(inTag, linkedTag string, inTagInfo *Tag, rowData []string) ([]string, error) {
	tempLine := make([]string, inTagInfo.FieldsNumber) // temp slice was initially filled by default value:""

	tempLine[0] = rowData[ignoredIndx] // first position is always placed by Tag

	for fieldPosition, fieldValue := range rowData[ignoredIndx+1:] {
		if fieldValue == "" {
			continue
		}

		fieldBlocks := strings.Split(fieldValue, OR)

		switch len(fieldBlocks) { // MUST BE LINK
		case 1: // WITHOUT OPPORTUNITY GETTING DEFAULT_VALUE
		case 2: // MUST BE TEMPLATE AND DEFAULT_VALUE
		default:
			
		}

		// err := c.setFieldInOutputRow(fieldName, &fieldInfo, tempLine)
		// if err != nil {
		// 	return nil, err
		// }
	}

	return tempLine, nil
}

func (c *Converter) MovePointerIndx(tag string) error {
	tagsList, ok := c.MsgSource.Tags[TagName(tag)]
	if !ok {
		return fmt.Errorf(ErrInputTagMSGNotFound, tag, c.MsgSource)
	}
	if len(tagsList) == 0 {
		return fmt.Errorf(ErrWrongSliceOfTag, tag)
	}
	if len(tagsList) == 1 {
		return nil // ignore if tagsList is 1
	}

	pointerToIndexInMultiTag++
	return nil
}

// WARNING: Important element
func (c *Converter) GetPointerIndx(inputTag string) (int, error) {
	tagsList, ok := c.MsgSource.Tags[TagName(inputTag)]
	if !ok {
		return 0, fmt.Errorf(ErrInputTagMSGNotFound, inputTag, c.MsgSource)
	}

	if len(tagsList) == 0 {
		return 0, fmt.Errorf(ErrWrongSliceOfTag, inputTag)
	}
	if len(tagsList) == 1 {
		return 0, nil
	}

	// len(tagsList) >= 2
	if pointerToTag == "" {
		pointerToTag = inputTag // It's means that we meet the multi tag
	} else if pointerToTag != inputTag {
		return 0, fmt.Errorf(ErrManyMultiTags, pointerToTag, inputTag)
	}

	// we reduce pointer on 1 point because pointer has already moved in the code above .
	return pointerToIndexInMultiTag - 1, nil
}

func (c *Converter) ResetParams() {
	pointerToIndexInMultiTag = defaultPointerToIndexInMultiTag
	pointerToTag = defaultPointerToTag
}

// setFieldInOutputRow is distributing of fields by linked fields number
func (c *Converter) setFieldInOutputRow(fN string, fI *Field, tL []string) error {
	outputPosition := fI.Position - 1 // we must work with index
	if int(outputPosition) == 0 {     // field outputPosition cannot be 0, because 0 position is Tag
		return fmt.Errorf(ErrWrongFieldPosition, fN, c.Output)
	}

	LinkedFields := fI.Linked // get list of avalible field links such as [tag-position]
	switch len(LinkedFields) {
	case 1:
		err := c.setValueToFieldWithOneLink(fN, fI, tL)
		if err == nil {
			return nil
		}

		fallthrough // if setValueToFieldWithOneLink has been unsuccessful, try set default value

	case 0:
		err := c.setDefaultValueToField(fN, fI, tL)
		if err != nil {
			return err
		}

	default:
		err := c.setValueToFieldWithMoreLinks(fN, fI, tL)
		if err != nil {
			return err
		}
	}

	return nil
}

// setDefaultValueToField
func (c *Converter) setDefaultValueToField(fN string, fI *Field, tL []string) error {
	outputPosition := fI.Position - 1 // cannot be equal to 0

	defaultValue := fI.DefaultValue
	if defaultValue != "" {
		pos := int(outputPosition) // integer representation of a number

		if isInt(outputPosition) { // Check that postion int or float
			tL[pos] = defaultValue
		} else {
			c.setFieldComponent(tL, pos, getTenth(outputPosition), defaultValue)
		}
	} else {
		return fmt.Errorf(ErrWrongFieldMetadata, fN, c.Input)
	}

	return nil
}

// setValueToField
//
// param::rowFields - it's slice contains only fields without tag
//
// NOTE: checking inputPosition is absent
func (c *Converter) setValueToField(fI *Field, inPos float64, tL []string, rowFields []string) error {
	outputPosition := fI.Position - 1 // cannot be equal to 0
	inputPosition := inPos - 2        // we subtract 2 because when we split the input msg, we separated those, and also we need to move on to the indexes

	// Check that postion int or float
	if isInt(inputPosition) && isInt(outputPosition) { // if inputPosition and outputPosition Int
		posOu := int(outputPosition)
		posInp := int(inputPosition)

		tL[posOu] = rowFields[posInp]

	} else if isInt(inputPosition) { // if inputPosition Int
		posInp := int(inputPosition)

		c.setFieldComponent(tL, int(outputPosition), getTenth(outputPosition), rowFields[posInp])

	} else if isInt(outputPosition) { // if outputPosition Int
		posInp := int(inputPosition)         // round to less number (3.8 -> 3)
		subposInp := getTenth(inputPosition) // position in field
		posOu := int(outputPosition)

		val, err := c.getFieldComponent(rowFields, posInp, subposInp)
		if err != nil || val == "" {
			return err
		}

		tL[posOu] = val

	} else { // if inputPosition and outputPosition NOT Int
		posInp := int(inputPosition)         // round to less number (11.2 -> 11)
		subposInp := getTenth(inputPosition) // position in field

		val, err := c.getFieldComponent(rowFields, posInp, subposInp)
		if err != nil || val == "" {
			return err
		}

		c.setFieldComponent(tL, int(outputPosition), getTenth(outputPosition), val)
	}

	return nil
}

func (c *Converter) setValueToFieldWithOneLink(fN string, fI *Field, tL []string) error {
	// Get info about linked Field
	linkedInfo := strings.Split(fI.Linked[0], linkToFieldSeparator) // len must be 2 (tag - 1, fieldPosition - 2)
	if len(linkedInfo) != 2 {
		return fmt.Errorf(ErrWrongFieldLink, fN)
	}

	linkedTag, linkedFieldPosition := linkedInfo[0], linkedInfo[1]

	indxToRow, err := c.GetPointerIndx(linkedTag)
	if err != nil {
		return err
	}

	position, err := strconv.ParseFloat(linkedFieldPosition, 64)
	if err != nil {
		return err
	}

	err = c.setValueToField(fI, position, tL, (c.MsgSource.Tags[TagName(linkedTag)][indxToRow])[TagName(linkedTag)])
	if err != nil {
		return err
	}

	return nil
}

// setValueToFieldWithMoreLinks
func (c *Converter) setValueToFieldWithMoreLinks(fN string, fI *Field, tL []string) error {
	outputPosition := fI.Position - 1 // cannot be equal to 0

	linkedFields := fI.Linked
	lenLinked := len(linkedFields)

	line := ""
	for i, inputLinkField := range linkedFields { // MANDATORY: len linked fields more than 1
		linkedInfo := strings.Split(inputLinkField, linkToFieldSeparator) // len must be 2 (tag - 1, fieldPosition - 2)
		if len(linkedInfo) != 2 {
			return fmt.Errorf(ErrWrongFieldLink, fN)
		}

		linkedTag, linkedFieldPosition := linkedInfo[0], linkedInfo[1]

		indxToRow, err := c.GetPointerIndx(linkedTag) // WRONG CYCLE CALLS OF GET POINTER
		if err != nil {
			return err
		}

		inpPos, err := strconv.ParseFloat(linkedFieldPosition, 64) // only for check
		if err != nil {
			return err
		}
		inpPosInt := int(inpPos) - 2 // represantation of index, for get value from rowFields by tag

		if isInt(inpPos) {
			line += ((c.MsgSource.Tags[TagName(linkedTag)][indxToRow])[TagName(linkedTag)])[inpPosInt]

		} else {
			subposInp := getTenth(inpPos)
			val, err := c.getFieldComponent((c.MsgSource.Tags[TagName(linkedTag)][indxToRow])[TagName(linkedTag)], inpPosInt, subposInp)
			if err != nil || val == "" {
				return err
			}

			line += val
		}

		if i != (lenLinked - 1) { // The line should not end by separator
			line += c.Output.ComponentSeparator
		}
	}

	if isInt(outputPosition) {
		tL[int(outputPosition)] = line
	} else {
		c.setFieldComponent(tL, int(outputPosition), getTenth(outputPosition), line)
	}

	return nil
}

// setFieldComponent
//
// NOTE: be careful with subpos
func (c *Converter) setFieldComponent(tL []string, pos, subPos int, value string) {
	fieldComponents := strings.Split(tL[pos], c.Output.ComponentSeparator)
	fieldComponents[subPos-1] = value

	res := strings.Join(fieldComponents, c.Output.ComponentSeparator)
	tL[pos] = res
}

// getFieldComponent
//
// NOTE: be careful with subpos
func (c *Converter) getFieldComponent(rowFields []string, posInp, subposInp int) (string, error) {
	val := strings.Split(rowFields[posInp], c.Input.ComponentSeparator)[subposInp-1]
	if len(val) >= len(rowFields[posInp]) { // => we must compare with '=='
		return "", fmt.Errorf(ErrWrongComponentSplit, rowFields[posInp])
	}

	return val, nil
}

// isInt return that number(float64) is Int or not
func isInt(numb float64) bool {
	return math.Mod(numb, 1.0) == 0
}

// getTenth return tenth of number(float64)
func getTenth(numb float64) int {
	x := math.Round(numb*100) / 100
	return int(x*10) % 10
}
