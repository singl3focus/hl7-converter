package hl7converter

import (
	"fmt"
	"sort"
	"math"
	"bufio"
	"bytes"
	"strconv"
	"strings"
)

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

	UsingPositions bool
}

type OptionFunc func(*Converter)

func WithUsingPositions() OptionFunc {
	return func(n *Converter) {
	  n.UsingPositions = true
	}
}

func NewConverter(in, out *Modification, opts ...OptionFunc) (*Converter, error) {
	converter := &Converter{
		Input:  in,
		Output: out,
		LineSplit: GetCustomSplit(in.LineSeparator),
		MsgSource: &Msg{
			Tags: make(map[TagName]SliceFields),
		},
	}

	for _, opt := range opts {
		opt(converter)
	}

	return converter, nil
}

var (
	DefaultValuePointerIndx = 0
	PointerIndx = DefaultValuePointerIndx
)

func (c *Converter) ResetPointerIndx() {
	PointerIndx = DefaultValuePointerIndx
}

/*_______________________________________[PARSE MSG AND EXECUTE OPRIONS SPECIFIED IN config]_______________________________________*/

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
			return nil, NewErrUndefinedInputTag(tag, "ParseMsg func")
		}

		tempTag, tempFields, err := c.handleOptions(tag, fields) // [TODO] UPGRADE OPTIONS
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
// [NOTE]: CALL PANIC
func (c *Converter) handleOptions(tag string, fields []string) (string, []string, error) {
	options := c.Input.TagsInfo.Tags[tag].Options

	newFields := make([]string, len(fields))
	copy(newFields, fields)

	for _, option := range options {
		switch option {
			// example: FN - 31(1 - Tag), len(fields) - 28. That we need add 2 empty fields
		case "autofill":
			diff := (c.Input.TagsInfo.Tags[tag].FieldsNumber - 1) - len(fields)
			for i := 0; i < diff; i++ {
				newFields = append(newFields, "")
			}

		default:
			return tag, newFields, NewErrUndefinedOption(option, tag)
		}
	}

	return tag, newFields, nil
}

/*_______________________________________[GENERAL CONVERT]_______________________________________*/

// Convert
func (c *Converter) Convert(fullMsg []byte) (*Result, error) {
	// _________fill MsgSource in Converter structure
	tags, err := c.ParseMsg(fullMsg)
	if err != nil {
		return nil, err
	}
	c.MsgSource.Tags = tags

	var result *Result
	if c.UsingPositions {
		result, err = c.convertWithPositions()
	} else {
		result, err = c.convertByInput(fullMsg)
	}
	if err != nil {
		return nil, err
	}

	return result, err
}

func (c *Converter) convertByInput(fullMsg []byte) (*Result, error) {
	tagPointerPositions := make(map[string]int) // Tag - Position
	
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

		if _, ok := tagPointerPositions[inputTag]; !ok {
			tagPointerPositions[inputTag] = 0 // start index
		} else {
			tagPointerPositions[inputTag]++
		}
		PointerIndx = tagPointerPositions[inputTag] // DANGER: [TENDER SPOT]

		row, err := c.convertTag(outputTag, &outputTagInfo)
		if err != nil {
			return nil, err
		}

		newRow := NewRow(c.Output.FieldSeparator, row)
		rows = append(rows, newRow)
	}
	
	c.ResetPointerIndx()


	result := NewResult(c.Output.LineSeparator, rows)
	return result, nil
}

// convertWithPositions
//
// - Using tag.Count for change PointerIndx
// - Tags positons are static and are set in the configuration.
func (c *Converter) convertWithPositions() (*Result, error) {
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

	orderedTags := make([]string, 0, len(tempNumbers))
	for _, key := range tempNumbers {
		orderedTags = append(orderedTags, c.Output.TagsInfo.Positions[strconv.Itoa(key)])
	}

	// _________set output count of tags
	for _, outputTag := range orderedTags {
		tag := c.Output.TagsInfo.Tags[outputTag] 

		if s, ok := c.MsgSource.Tags[TagName(c.Output.TagsInfo.Tags[outputTag].Linked)]; !ok {
			tag.Count = 1
		} else {
			tag.Count = len(s)
		}

		c.Output.TagsInfo.Tags[outputTag] = tag
	}
	
	// _________get result
	rows := make([]*Row, 0, 1)
	for _, tag := range orderedTags {
		TagInfo, ok := c.Output.TagsInfo.Tags[tag]
		if !ok {
			return nil, NewErrUndefinedPositionTag(tag)
		}

		for i := 0; i < TagInfo.Count; i++ {
			PointerIndx = i // DANGER: [TENDER SPOT]

			row, err := c.convertTag(tag, &TagInfo)
			if err != nil {
				return nil, err
			}
			
			newRow := NewRow(c.Output.FieldSeparator, row)
			rows = append(rows, newRow)
		}

		c.ResetPointerIndx() // clear pointer indx for next tag
	}

	result := NewResult(c.Output.LineSeparator, rows)
	return result, nil
}

/*_______________________________________[CONVERT SPECIFIED TAG]_______________________________________*/

// convertTag
func (c *Converter) convertTag(outputTagName string, outputTagInfo *Tag) ([]*Field, error) {
	row := strings.Split(outputTagInfo.Tempalate, c.Output.FieldSeparator) // REPEAT BLOCK OF SPLITS

	if outputTagInfo.FieldsNumber != ignoredFieldsNumber {
		if len(row) != outputTagInfo.FieldsNumber || outputTagInfo.FieldsNumber < 1 {
			return nil,
				fmt.Errorf(ErrWrongFieldsNumber, outputTagName, outputTagInfo, len(row))
		}
	}

	outputLine, err := c.assembleOutRow(outputTagInfo, row)
	if err != nil {
		return nil, err
	}

	return outputLine, nil
}

// assembleOutRow creates outLine and fills it. Also inserts component separators in fields with components.
func (c *Converter) assembleOutRow(inTagInfo *Tag, rowData []string) ([]*Field, error) {
	tempLine := make([]*Field, inTagInfo.FieldsNumber) // temp slice was initially filled by default value:""
	for i := range tempLine {
        tempLine[i] = &Field{} // Init empty values for tempLine
    }

	
	tempLine[0].Value = rowData[ignoredIndx] // first position is always placed by Tag

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

			tempLine[fieldPosition].Value = value
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

				tempLine[fieldPosition].Value = value
			} else {
				value, err := c.getDefaultFieldValue(fieldBlocks[1])
				if err != nil {
					return nil, err
				}

				tempLine[fieldPosition].Value = value
			}


		default:
			return nil, NewErrWrongParamCount(fieldValue, OR)
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
				return "", NewErrInvalidLink(str)
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
		return "", NewErrEmptyDefaultValue(str)
	}

	return str, nil
}

func (c *Converter) getValueFromMSGbyLink(link string) (string, error) {
	var value string
	
	elems := strings.Split(link, linkToField) // parse: Tag - Position
	if len(elems) != 2 {
		return "", NewErrInvalidLink(link)
	}

	matchingTag, pos := elems[0], elems[1]
	if matchingTag == "" || pos == "" {
		return "", NewErrInvalidLinkElems(link)
	}

	inTagInfo, ok := c.MsgSource.Tags[TagName(matchingTag)]
	if !ok {
		return "", NewErrUndefinedInputTag(matchingTag, link)
	}

	if PointerIndx > (len(inTagInfo) - 1) {// countOfInputSameTagRows
		return "", NewErrTooBigIndex(PointerIndx, len(inTagInfo) - 1)
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
