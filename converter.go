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
		Tags: make(map[TagName]SliceFields),
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
	DefaultValuePointerIndx = 0
	PointerIndx = DefaultValuePointerIndx
)

func (c *Converter) ResetParams() {
	PointerIndx = DefaultValuePointerIndx
}

// GetCustomSplit
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

// ParseMsg return 'map[TagName]SliceOfTag' model for get fields data specified in output modification.
func (c *Converter) ParseMsg(fullMsg []byte) (map[TagName]SliceFields, error) {
	tags := make(map[TagName]SliceFields)

	scanner := bufio.NewScanner(bytes.NewReader(fullMsg))
	scanner.Split(c.LineSplit)

	for scanner.Scan() {
		token := scanner.Text() // [DEV] getting string representation of row
		rowFields := strings.Split(token, c.Input.FieldSeparator)

		tag, fields := rowFields[0], rowFields[1:]
		if _, ok := c.Input.TagsInfo.Tags[tag]; !ok {
			return nil, fmt.Errorf(ErrUndefinedInputTag, tag, "ParseMsg func")
		}

		tempTag, tempFields := c.handleOptions(tag, fields) // [TODO] UPGRADE OPTIONS
		processedTag, processedFields := TagName(tempTag), Fields(tempFields)

		if _, ok := tags[processedTag]; ok {
			tags[processedTag] = append(tags[processedTag], processedFields)
		} else {
			tags[processedTag] = make(SliceFields, 0, 1)
			tags[processedTag] = append(tags[processedTag], processedFields)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, ErrInvalidParseMsg
	}

	return tags, nil
}

// handleOptions
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

// Convert
func (c *Converter) Convert(fullMsg []byte) ([][]string, error) {

	// _________fill MsgSource in Converter structure
	tags, err := c.ParseMsg(fullMsg)
	if err != nil {
		return nil, err
	}
	c.MsgSource.Tags = tags

	
	// _________prepare 'Ordered Slice of Tags' for convert
	tempNumbers := make([]int, 0, 1)
	for key, _ := range c.Output.TagsInfo.Positions {
		k, err := strconv.Atoi(key)
		if err != nil {
			return nil, err
		}
		tempNumbers = append(tempNumbers, k)
	}
	sort.Ints(tempNumbers)

	OutputTags := make([]string, 0, len(tempNumbers))
	for _, key := range tempNumbers {
		OutputTags = append(OutputTags, c.Output.TagsInfo.Positions[strconv.Itoa(key)])
	}

	// _________set output count of tags  
	for _, outputTag := range OutputTags {
		tag := c.Output.TagsInfo.Tags[outputTag] 

		if s, ok := tags[TagName(c.Output.TagsInfo.Tags[outputTag].Linked)]; !ok {
			tag.Count = 1
		} else {
			tag.Count = len(s)
		}

		c.Output.TagsInfo.Tags[outputTag] = tag
	}
 
	return c.convert(OutputTags)
}

func (c *Converter) convert(orderedTags []string) ([][]string, error) {
	splitedRows := make([][]string, 0, 1)
	for _, tag := range orderedTags {
		TagInfo, ok := c.Output.TagsInfo.Tags[tag]
		if !ok {
			return nil, fmt.Errorf(ErrUndefinedPositionTag, tag)
		}

		for i := 0; i < TagInfo.Count; i++ {
			PointerIndx = i

			row, err := c.convertTag(tag, &TagInfo)
			if err != nil {
				return nil, err
			}

			splitedRows = append(splitedRows, row)
		}

		c.ResetParams() // clear pointer indx for next tag
	}

	return splitedRows, nil
}

// convertTag
func (c *Converter) convertTag(TagName string, TagInfo *Tag) ([]string, error) {
	row := strings.Split(TagInfo.Tempalate, c.Output.FieldSeparator) // REPEAT BLOCK OF SPLITS

	if TagInfo.FieldsNumber != ignoredFieldsNumber {
		if len(row) != TagInfo.FieldsNumber || TagInfo.FieldsNumber < 1 {
			return nil,
				fmt.Errorf(ErrWrongFieldsNumber, TagName, TagInfo, len(row))
		}
	}

	outputLine, err := c.assembleOutRow(TagInfo, row)
	if err != nil {
		return nil, err
	}

	return outputLine, nil
}

// assembleOutRow creates outLine and fills it. Also inserts component separators in fields with components.
func (c *Converter) assembleOutRow(inTagInfo *Tag, rowData []string) ([]string, error) {
	tempLine := make([]string, inTagInfo.FieldsNumber) // temp slice was initially filled by default value:""

	tempLine[0] = rowData[ignoredIndx] // first position is always placed by Tag

	// [DEV] - fieldPosition started from '0' not from 'ignoredIndx+1'
	for fieldPosition, fieldValue := range rowData[ignoredIndx+1:] {
		if fieldValue == "" {
			continue
		}

		fieldPosition++ // because we started fro, 'ignoredIndx+1'

		fieldBlocks := strings.Split(fieldValue, OR)

		switch len(fieldBlocks) {
		case 1: // WITHOUT OPPORTUNITY GETTING DEFAULT_VALUE
			mask, err := c.TempalateParse(fieldBlocks[0])
			if err != nil {
				return nil, err
			}

			value, err := c.getFieldValue(mask, fieldBlocks[0])
			if err != nil {
				return nil, err
			}

			tempLine[fieldPosition] = value
		case 2: // MUST BE TEMPLATE OR DEFAULT_VALUE
			if fieldBlocks[0] != "" {
				mask, err := c.TempalateParse(fieldBlocks[0])
				if err != nil {
					return nil, err
				}

				value, err := c.getFieldValue(mask, fieldBlocks[0])
				if err != nil {
					return nil, err
				}

				tempLine[fieldPosition] = value
			} else {
				value, err := c.getDefaultFieldValue(fieldBlocks[1])
				if err != nil {
					return nil, err
				}

				tempLine[fieldPosition] = value
			}


		default:
			return nil, fmt.Errorf(ErrWrongParamCount, fieldValue, OR)
		}
	}

	return tempLine, nil
}

func (c *Converter) TempalateParse(str string) ([]int, error) {
	mask := make([]int, 0, len(str)) // example: [1,1,1,1,0,0,0,1,1,1], 1 - Symbol, 0 - Link
	stLinkIndx, endLinkIndx := 0, 0

	for i, v := range str {
		if string(v) == linkElemSt {
			stLinkIndx = i 
		} else if string(v) == linkElemEnd {
			endLinkIndx = i
		}

		if endLinkIndx > stLinkIndx {
			for j := stLinkIndx; j < endLinkIndx; j++ {
				mask[j] = itLink
			}
			mask = append(mask, itLink)
			stLinkIndx, endLinkIndx = 0, 0
		} else {
			mask = append(mask, itSymbol)
		}
	}

	return mask, nil
}

// getFieldValue is distributing of fields by linked fields number
func (c *Converter) getFieldValue(mask []int, str string) (string, error) {
	var output string

	for i := 0; i < len(str); i++ {
		if mask[i] == itSymbol {
			output += string(str[i])
		
		} else if mask[i] == itLink {
			var link string // // parse: Tag-Position (Without '<', '>')
			for j := i; j < len(str); j++ {
				if (mask[j] == itSymbol){
					link = str[ i+1 : j-1 ]		
					i = j-1 // subtract '1' because next step will be increment
					break
				} else if (j == len(str)-1) {
					link = str[ i+1 : j ]		
					i = j // cycle must be ended
					break
				}
			}

			if link == "" {
				return "", fmt.Errorf(ErrInvalidLink, str)
			} 

			value, err := c.getValueFromMSGbyLink(link)
			if err != nil {
				return "", err
			}

			output += value
		}
	}	

	return output, nil
}

// getDefaultFieldValue
func (c *Converter) getDefaultFieldValue(str string) (string, error) {
	if str == "" {
		return "", fmt.Errorf(ErrEmptyDefaultValue, str)
	}

	return str, nil
}

func (c *Converter) getValueFromMSGbyLink(link string) (string, error) {
	var value string
	
	elems := strings.Split(link, linkToField) // parse: Tag - Position
	if len(elems) != 2 {
		return "", fmt.Errorf(ErrInvalidLink, link)
	}

	matchingTag, pos := elems[0], elems[1]
	if matchingTag == "" || pos == "" {
		return "", fmt.Errorf(ErrInvalidLinkElems, link)
	}

	inTagInfo, ok := c.MsgSource.Tags[TagName(matchingTag)]
	if !ok {
		return "", fmt.Errorf(ErrUndefinedInputTag, matchingTag, link)
	}

	if PointerIndx > (len(inTagInfo) - 1) {// countOfInputSameTagRows
		return "", fmt.Errorf(ErrTooBigIndex, PointerIndx, len(inTagInfo) - 1)
	}

	position, err := strconv.ParseFloat(pos, 64)
	if err != nil {
		return "", err
	}

	// [NOTE]: Be carefull with indexes
	differenceIndexes := 2

	if isInt(position) {
		value = inTagInfo[PointerIndx][int(position) - differenceIndexes] // example: link(MSH-2), (inTagInfo - MSH: [[A, B, C]]), 
	} else {
		fieldPosIndx, componentPosIndx := int(position) - differenceIndexes, getTenth(position) - 1
		
		fieldValue := inTagInfo[PointerIndx][fieldPosIndx]

		components := strings.Split(fieldValue, c.Input.ComponentSeparator)
		if len(components) == 1 {
			return "", fmt.Errorf(ErrWrongComponentsNumber, fieldValue, link)
		}

		if componentPosIndx > (len(components) - 1) {
			return "", fmt.Errorf(ErrWrongComponentLink, link, componentPosIndx + 1, len(components), matchingTag)
		}

		value = components[componentPosIndx]
	}

	return value, nil
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
