package hl7converter

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"
)

type Converter struct {
	// Data parsed from config.
	// Goal: find metadata(position, default_value, components_number, linked and data about Tags, Separators)
	// about field so that then get value some field.
	Input, Output *Modification

	// For effective split by rows of input message with help "bufio.Scanner"
	LineSplit bufio.SplitFunc

	// Convertred structure of input message for fast find a needed field by tag
	MsgSource *Msg

	UsingPositions bool
	UsingAliases bool

	// Pointer to Tag Index in the MSG (for rows with same tags)
	pointerIndx int
}

type OptionFunc func(*Converter)

func WithUsingPositions() OptionFunc {
	return func(n *Converter) {
	  n.UsingPositions = true
	}
}

func WithUsingAliases() OptionFunc {
	return func(n *Converter) {
	  n.UsingAliases = true
	}
}

func NewConverter(p *ConverterParams, opts ...OptionFunc) (*Converter, error) {
	converter := &Converter{
		Input:  p.InMod,
		Output: p.OutMod,
		LineSplit: GetCustomSplit(p.InMod.LineSeparator),
		MsgSource: &Msg{
			Tags: make(map[TagName]SliceFields),
		},

		pointerIndx: defaultValuePointerIndx,
	}

	for _, opt := range opts {
		opt(converter)
	}

	return converter, nil
}

var (
	defaultValuePointerIndx = 0 // TODO: add opporunity of parallel using converter (be careful with pointerIndx) 
	// pointerIndx = defaultValuePointerIndx // TODO: move to Converter!
)

// TODO: add opporunity of parallel using converter (be careful with pointerIndx) 

func (c *Converter) resetPointerIndx() {
	c.pointerIndx = defaultValuePointerIndx
}

func (c *Converter) resetState() {
	c.resetPointerIndx()
}

/*_______________________________________[PARSE MSG AND EXECUTE OPRIONS SPECIFIED IN config]_______________________________________*/

// GetCustomSplit
// TODO: comment
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
func (c *Converter) handleOptions(tag string, fields []string) (string, []string, error) {
	options := c.Input.TagsInfo.Tags[tag].Options

	newFields := make([]string, len(fields))
	copy(newFields, fields)

	for _, option := range options {
		switch option {
		case "autofill":
			// [EXAMPLE]: FN - 31 (1 - Tag), len(fields) - 28. That we need add 2 empty fields

			diff := (c.Input.TagsInfo.Tags[tag].FieldsNumber - 1) - len(fields)
			
			for i := 0; i < diff; i++ { // [DEV] not use {for range diff} in order to easy support win7 version
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
// TODO: errors
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
	c.MsgSource.Tags = tags

	c.resetState()

	if c.UsingPositions {
		result, err = c.convertWithPositions()
	} else {
		result, err = c.convertByInput(fullMsg)
	}
	if err != nil {
		return nil, err
	}

	if c.UsingAliases {
		if err = result.ApplyAliases(c.Output.Aliases) ; err != nil {
			return nil, err
		}
	}

	return result, err
}

func (c *Converter) convertByInput(fullMsg []byte) (*Result, error) {
	defer c.resetPointerIndx()

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
		c.pointerIndx = tagPointerPositions[inputTag] // * DANGER: [TENDER SPOT]

		row, err := c.convertTag(outputTag, &outputTagInfo)
		if err != nil {
			return nil, err
		}

		rows = append(rows, NewRow(c.Output.FieldSeparator, row))
	}

	// c.resetPointerIndx() // TODO: check it, how about multiple msgs
	
	return NewResult(c.Output.LineSeparator, rows), nil
}

// convertWithPositions
//
// - Using tag.Count for change pointerIndx
// - Tags positons are static and are set in the configuration.
func (c *Converter) convertWithPositions() (*Result, error) {
	orderedTags, err := c.Output.OrderedPositionTags()
	if err != nil {
		return nil, err
	}

	// set output count of tags
	for _, outputTag := range orderedTags {
		tag := c.Output.TagsInfo.Tags[outputTag] 

		if s, ok := c.MsgSource.Tags[TagName(c.Output.TagsInfo.Tags[outputTag].Linked)]; !ok {
			tag.Count = 1
		} else {
			tag.Count = len(s)
		}

		c.Output.TagsInfo.Tags[outputTag] = tag
	}
	
	// get result
	rows := make([]*Row, 0, 1)
	for _, tag := range orderedTags {
		TagInfo, ok := c.Output.TagsInfo.Tags[tag]
		if !ok {
			return nil, NewErrUndefinedPositionTag(tag)
		}

		for i := 0; i < TagInfo.Count; i++ {
			c.pointerIndx = i // * DANGER: [TENDER SPOT]

			row, err := c.convertTag(tag, &TagInfo)
			if err != nil {
				return nil, err
			}
			
			rows = append(rows, NewRow(c.Output.FieldSeparator, row))
		}

		c.resetPointerIndx() // clear pointer indx for next tag
	}

	return NewResult(c.Output.LineSeparator, rows), nil
}

/*_______________________________________[CONVERT SPECIFIED TAG]_______________________________________*/

// convertTag
func (c *Converter) convertTag(outputTag string, outputTagInfo *Tag) ([]*Field, error) {
	row := strings.Split(outputTagInfo.Tempalate, c.Output.FieldSeparator)

	if outputTagInfo.FieldsNumber != ignoredFieldsNumber {
		if len(row) != outputTagInfo.FieldsNumber || outputTagInfo.FieldsNumber < 1 {
			return nil, NewErrWrongFieldsNumber(outputTag, outputTagInfo, len(row))
		}
	}

	outputLine, err := c.assembleOutRow(outputTagInfo, row)
	if err != nil {
		return nil, err
	}

	return outputLine, nil
}

// assembleOutRow creates out line and fills it. Also inserts component separators in fields with components.
func (c *Converter) assembleOutRow(inTagInfo *Tag, rowData []string) ([]*Field, error) {
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

			value, err := c.getFieldValue(mask, fieldBlocks[0])
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

				value, err := c.getFieldValue(mask, fieldBlocks[0])
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
func (c *Converter) getFieldValue(mask []int, str string) (string, error) {
	var builder strings.Builder

	for i := 0; i < len(str); i++ {
		if mask[i] == itSymbol {
			builder.WriteByte(str[i])
		} else if mask[i] == itLink {
			var link string // parse: Tag-Position (Without '<', '>')
			
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

	if c.pointerIndx > (len(inTagInfo) - 1) { // countOfInputSameTagRows
		return "", NewErrTooBigIndex(c.pointerIndx, len(inTagInfo) - 1)
	}

	position, err := strconv.ParseFloat(pos, 64)
	if err != nil {
		return "", err
	}

	// [NOTE]: Be carefull with indexes
	differenceIndexes := 2

	if isInt(position) {
		value = inTagInfo[c.pointerIndx][int(position) - differenceIndexes] // example: link(MSH-2), (inTagInfo - MSH: [[A, B, C]]), 
	} else {
		fieldPosIndx, componentPosIndx := int(position) - differenceIndexes, getTenth(position) - 1
		
		fieldValue := inTagInfo[c.pointerIndx][fieldPosIndx]

		components := strings.Split(fieldValue, c.Input.ComponentSeparator)
		if len(components) == 1 {
			return "", NewErrWrongComponentsNumber(fieldValue, link)
		}

		if componentPosIndx > (len(components) - 1) {
			return "", NewErrWrongComponentLink(link, componentPosIndx + 1, len(components), matchingTag)
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