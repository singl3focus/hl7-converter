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

type Converter struct {
	// Data parsed from config.
	// goal: find metadata(position, default_value, components_number, linked and data about Tags, Separators)
	// about field so that then get value some field.
	input, output *Modification 

	// For effective split by rows of input message with help "bufio.Scanner"
	lineSplit func(data []byte, atEOF bool)(advance int, token []byte, err error)

	// Convertred structure of input message for fast find a needed field by tag
	inMsg *Msg 

	// Slice with with ordered tags for following the structure of the message
	outOdreredTags []string
}


func NewConverter(passedInput, passedOutput *Modification) (*Converter, error) {
	converter := &Converter{
		input: passedInput,
		output: passedOutput,

		lineSplit: GetCustomSplit(passedInput.LineSeparator),
	}

	return converter, nil
}


func (c *Converter) Convert(fullMsg []byte) (error) {
	_, err := c.ConvertToMSG(fullMsg) // fill inMsg in Converter structure
	if err != nil {
		return err
	}

	_, err = c.GetOutOdreredTags() // fill OutOdreredTags in Converter structure
	if err != nil {
		return err
	}

	for _, tag := range c.outOdreredTags{
		res, err := c.AssembleOutputRowMsg(tag)
		if err != nil {
			return err
		}

		_ = res
	}

	return nil
}



// ConvertToMSG  function
// return MSG model for get fields data specified in output 'linked_fields'.
//
// NOTE: We can do without copying the structure MSG
func (c *Converter) ConvertToMSG(fullMsg []byte) (*Msg, error) {
	tags := make(map[TagName]SliceOfTag)
	
	scanner := bufio.NewScanner(bytes.NewReader(fullMsg))
	scanner.Split(c.lineSplit)
	
	for scanner.Scan() {
		temp := make(TagValues)

		token := scanner.Text() // getting string representation of row
		rowFields := strings.Split(token, c.input.FieldSeparator)
		
		tag, fields := TagName(rowFields[0]), Fields(rowFields[1:])
		if _, ok := c.input.Tags[string(tag)]; !ok {
			return nil, fmt.Errorf(ErrUndefinedInputTag, tag)
		}

		temp[TagName(tag)] = fields

		if _, ok := tags[tag]; ok {
			tags[tag] = append(tags[tag], temp)

		} else {
			tags[tag] = make(SliceOfTag, 0, 10) // capacity is 10 because it's optimal value, which describe average rows of message 
			tags[tag] = append(tags[tag], temp)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("convert input messsge to Msg struct has been unsuccesful")
	}

	c.inMsg.Tags = tags

	return c.inMsg, nil
}

 
// GetCustomSplit function
func GetCustomSplit(sep string) (func(data []byte, atEOF bool) (advance int, token []byte, err error)) {
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


// IDENTIFY BY output Modification (field: Types) and compare it with Tags in Msg
func IndetifyMsg(msg *Msg, input *Modification) string {
	return ""
}


// GetOutOdreredTags function
//
// NOTE: NEED BENCHMARK + We can do without copying the slice
func (c *Converter) GetOutOdreredTags() ([]string, error) {
	orderedTags := make([]string, 0, 5) // capacity is 5 because it's optimal value, which describe average tags of modification 
	
	sotrtingSlice := make([][]string, 0, 5)
	for tagName, tag := range c.output.Tags{
		sotrtingSlice = append(sotrtingSlice, []string{strconv.Itoa(tag.Position), tagName})
	}
	
	sort.Slice(sotrtingSlice, func(i, j int) bool {
			left, err := strconv.Atoi(sotrtingSlice[i][0])
			if err != nil {
				panic(err)
			}
			right, err := strconv.Atoi(sotrtingSlice[j][0])
			if err != nil {
				panic(err)
			}

			return left < right
		})
	
	for _, v := range sotrtingSlice{
		orderedTags = append(orderedTags, v[1])
	}

	c.outOdreredTags = orderedTags

	return c.outOdreredTags, nil
}



// AssembleOutputRowMsg function
//
// Creates outLine and fills it. Also inserts component separators in fields with components.
func (c *Converter) AssembleOutputRowMsg(outTag string) (string, error) {	
	outputFields := c.output.Tags[outTag].Fields

	tempLine := make([]string, c.output.Tags[outTag].FieldsNumber) // temp slice was initially filled by default value:""
	tempLine[0] = outTag // first position is always placed by Tag 


	for fieldName, fieldInfo := range outputFields { // go through the fields of the output structure
		if fieldInfo.ComponentsNumber < 0 {
			return "", fmt.Errorf(ErrNegativeComponentsNumber, fieldName)
		}

		if fieldInfo.ComponentsNumber > 0 { // min. count of components is 2
			if fieldInfo.ComponentsNumber == 1 {
				return "", fmt.Errorf(ErrWrongComponentsNumber, fieldName)
			
			} else { 
				outputPosition := int(fieldInfo.Position) - 1 // we must work with index
				if int(outputPosition) == 0 { // field outputPosition cannot be 0, because 0 position is Tag
					return "", fmt.Errorf(ErrWrongFieldPosition, outputPosition + 1, fieldName)
				}

				if tempLine[outputPosition] == "" { // we must save previous data that's why we checking current value of field 
					tempLine[outputPosition] = strings.Repeat(c.output.ComponentSeparator, fieldInfo.ComponentsNumber - 1) // count separator is ComponentsNumber - 1
				} 
			}
		}

		err := c.setFieldInOutputRow(fieldName, &fieldInfo, tempLine)
		if err != nil {
			return "", err
		}
	}

	// Join tempLine by FieldSeparator
	res := strings.Join(tempLine, c.output.FieldSeparator) // Suggestion: make this optional

	return res, nil 
} 


// setFieldInOutputRow
//
//
func (c *Converter) setFieldInOutputRow(fN string, fI *Field, tL []string) error {
	outputPosition := fI.Position - 1 // we must work with index
	if int(outputPosition) == 0 { // field outputPosition cannot be 0, because 0 position is Tag
		return fmt.Errorf("incorrect position %f by fieldName %s", outputPosition, fN)
	}

	LinkedFields := fI.Linked // get list of avalible field links such as [tag-position]

	switch len(LinkedFields) {
	case 1:
		// Get info about linked Field
		linkedInfo := strings.Split(LinkedFields[0], link_separator) // len must be 2 (tag - 1, fieldName - 2)
		if len(linkedInfo) != 2 {
			return fmt.Errorf("specified field link is incorrect, field %s", fN)
		}
		
		linkedTag, linkedFieldName := linkedInfo[0], linkedInfo[1]

		// Get info about linked Field
		inFieldInfo, ok := in.Tags[linkedTag].Fields[linkedFieldName]

		if ok { // if we have not found the field, we set the default value
			err := setValueToField(in, out, fI, inFieldInfo.Position, tL, fullMsg.Tags[linkedTag])
			if err == nil { // if set value has been unsecseful try set default
				break
			}
		}

		fallthrough // if set value return err

	case 0:
		err := setDefaultValueToField(in, out, fN, fI, tL)
		if err != nil {
			return err
		}
	
	default:
		err := setValueToFieldWithMoreLinks(in, out, fI, tL, fullMsg)
		if err != nil {
			return err
		}
	}

	return nil
}


// setDefaultValueToField
//
//
func setDefaultValueToField(in, out *Modification, fN string, fI *Field, tL []string) error {
	outputPosition := fI.Position - 1 // cannot be equal to 0
	
	defaultValue := fI.DefaultValue
	if defaultValue != "" {
		pos := int(outputPosition) // integer representation of a number
		if isInt(outputPosition) { // Check that postion int or float
			tL[pos] = defaultValue
		} else {
			setFieldComponent(out, tL, pos, getTenth(outputPosition), defaultValue)
		}
	} else {
		return fmt.Errorf("fieldName %s in output modification %v hasn't have a linked_fields Or default value is not specified Or not found in linked_fields", fN, in)
	}

	return nil
}


// setValueToField
//
// NOTE: checking inputPosition is absent
func setValueToField(in, out  *Modification, fI *Field, inPos float64, tL []string, rowFields []string) error {
	outputPosition := fI.Position - 1 // cannot be equal to 0
	inputPosition := inPos - 2 // we subtract 2 because when we split the input msg, we separated those, and also we need to move on to the indexes

	// Check that postion int or float
	if isInt(inputPosition) && isInt(outputPosition) { // if inputPosition and outputPosition Int
		posOu := int(outputPosition)
		posInp := int(inputPosition)

		tL[posOu] = rowFields[posInp]	

	} else if isInt(inputPosition) { // if inputPosition Int
		posInp := int(inputPosition)


		setFieldComponent(out, tL, int(outputPosition), getTenth(outputPosition), rowFields[posInp])

	} else if isInt(outputPosition) { // if outputPosition Int
		posInp := int(inputPosition) // round to less number (3.8 -> 3)
		subposInp := getTenth(inputPosition) // position in field 
		posOu := int(outputPosition)

		val, err := getFieldComponent(in, rowFields, posInp, subposInp)
		if err != nil || val == "" {
			return err
		}

		tL[posOu] = val

	} else { // if inputPosition and outputPosition NOT Int
		posInp := int(inputPosition) // round to less number (11.2 -> 11)
		subposInp := getTenth(inputPosition) // position in field 

		val, err := getFieldComponent(in, rowFields, posInp, subposInp)
		if err != nil || val == "" {
			return err
		}

		setFieldComponent(out, tL, int(outputPosition), getTenth(outputPosition), val)
	}

	return nil
}


// setValueToFieldWithMoreLinks
func setValueToFieldWithMoreLinks(in, out *Modification, fI *Field,  tL []string, fullMsg *Msg) error {
	outputPosition := fI.Position - 1 // cannot be equal to 0
	linkedFields := fI.Linked
	lenLinked := len(linkedFields) // ADD: 

	if isInt(outputPosition) {
		line := ""
		for i, inputField := range linkedFields { // Len Linked fields more than 1

			linkedInfo := strings.Split(inputField, link_separator) // len must be 2 (tag - 1, fieldName - 2)
			if len(linkedInfo) != 2 {
				return fmt.Errorf("specified field link is incorrect, linked_field %s", inputField)
			}
			
			linkedTag, linkedFieldName := linkedInfo[0], linkedInfo[1]


			inpPos := in.Tags[linkedTag].Fields[linkedFieldName].Position // only for check
			inpPosInt := int(inpPos) - 2 // represantation of index, for putting

			if isInt(inpPos) {
				line += fullMsg.Tags[linkedTag][inpPosInt]

			} else {
				subposInp := getTenth(inpPos)
				val, err := getFieldComponent(in, fullMsg.Tags[linkedTag], inpPosInt, subposInp)
				if err != nil || val == "" {
					return err
				}

				line += val
			}
			

			if i != (lenLinked - 1) { // The line should not end by separator
				line += out.ComponentSeparator
			}  
		}

		tL[int(outputPosition)] = line

	} else {
		line := ""
		for i, inputField := range linkedFields { // Len Linked fields more than 1

			linkedInfo := strings.Split(inputField, link_separator) // len must be 2 (tag - 1, fieldName - 2)
			if len(linkedInfo) != 2 {
				return fmt.Errorf("specified field link is incorrect, linked_field %s", inputField)
			}
			
			linkedTag, linkedFieldName := linkedInfo[0], linkedInfo[1]


			inpPos := in.Tags[linkedTag].Fields[linkedFieldName].Position // only for check
			inpPosInt := int(inpPos) - 2 // represantation of index, for putting

			if isInt(inpPos) {
				line += fullMsg.Tags[linkedTag][inpPosInt]

			} else {
				subposInp := getTenth(inpPos)
				val, err := getFieldComponent(in, fullMsg.Tags[linkedTag], inpPosInt, subposInp)
				if err != nil || val == "" {
					return err
				}

				line += val
			}
			

			if i != (lenLinked - 1) { // The line should not end by separator
				line += out.ComponentSeparator
			}  
		}

		setFieldComponent(out, tL, int(outputPosition), getTenth(outputPosition), line)
	}

	return nil
}


// setFieldComponent
//
// NOTE: be careful with subpos
func setFieldComponent(out *Modification, tL []string, pos, subPos int, value string) {
	fieldComponents := strings.Split(tL[pos], out.ComponentSeparator)
	fieldComponents[subPos - 1] = value

	res := strings.Join(fieldComponents, out.ComponentSeparator)
	tL[pos] = res
}


// getFieldComponent
func getFieldComponent(in *Modification, rowFields []string, posInp, subposInp int) (string, error) {
	val := strings.Split(rowFields[posInp], in.ComponentSeparator)[subposInp - 1] // Be careful with subpos
	if len(val) >= len(rowFields[posInp]) {
		return "", fmt.Errorf("incorrect field %s, the field component could not be pulled out", rowFields[posInp])
	}

	return val, nil
}



// isInt return that number(float64) is Int or not
func isInt(numb float64) bool {
	return math.Mod(numb, 1.0) == 0
}

// getTenth return tenth of number(float64) 
func getTenth(numb float64) int {
	x := math.Round(numb * 100) / 100
	return int(x * 10) % 10
}
