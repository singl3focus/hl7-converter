package hl7converter

import (
	"fmt"
	"math"
	"bytes"
	"bufio"
	"strings"
)

const (
	link_separator = "."
)


// ConvertToMSG 
// return MSG model for get fields value specified in output 'linked_fields'
//
// WARNING: message cannot contain a same tags because for access to value use map (map cannot contain same key)
func ConvertToMSG(input *Modification, fullMsg []byte, customSplit func(data []byte, atEOF bool)(advance int, token []byte, err error)) (*Msg, error) {
	var msg Msg
	
	scanner := bufio.NewScanner(bytes.NewReader(fullMsg))
	scanner.Split(customSplit)
	
	temp := make(map[string][]string)
	for scanner.Scan() {
		token := scanner.Text()
		rowFields := strings.Split(token, input.FieldSeparator)
		
		tag, fields := rowFields[0], rowFields[1:]
		temp[tag] = fields
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("convert msg to Msg struct has been unsuccesful")
	}

	msg.Tags = temp

	return &msg, nil
}


// ConvertRow
// return converted line or "skip" if specified linkedTag is "none"
//
// WARNING: message cannot contain a field separator, otherwise the conversion will be incorrect
func ConvertRow(input, output *Modification, msg string, fullMsg *Msg) (string, error) {
	currentRowFields := strings.Split(msg, input.FieldSeparator)

	tag := currentRowFields[0]
	if len(tag) >= len(currentRowFields) {
		return "", fmt.Errorf("identify msg and split input msg has been unsuccessful") 
	}

	tagInfo, ok := input.Tags[tag]
	if !ok {
		return "", fmt.Errorf("tag %s not found in input modification %v", tag, input) 
	}


	var matchingTag string
	for _, linkedTag := range tagInfo.Linked { // the first tag get will be taken for output
		if linkedTag == "none" {
			return "skip", nil 
		}

		if _, ok := output.Tags[linkedTag]; ok {
			matchingTag = linkedTag
			break
		}		
	}

	if matchingTag == "" {
		return "", fmt.Errorf("linked tag in tag %s not found in output modification %v", tag, output)
	}


	outputLine, err := assembleOutputRowMsg(input, output, matchingTag, fullMsg)
	if err != nil {
		return "", err
	}		


	return outputLine, nil
}


// assembleOutputRowMsg
//
//
func assembleOutputRowMsg(input, output *Modification, matchingTag string, fullMsg *Msg) (string, error) {	
	outputFields := output.Tags[matchingTag].Fields

	tempLine := make([]string, output.Tags[matchingTag].FieldsNumber) // temp slice was initially filled by default value:""
	tempLine[0] = matchingTag // first position is always placed by Tag 

	for fieldName, fieldInfo := range outputFields { // go through the fields of the output structure

		if fieldInfo.ComponentsNumber > 0 { // min. count of components is 2
			
			if fieldInfo.ComponentsNumber == 1 {
				return "", fmt.Errorf("Commponent count can be equal to 0 or more than 1, else it's field hasn't have components")
			} else {
				outputPosition := int(fieldInfo.Position) - 1 // we must work with index
				if int(outputPosition) == 0 { // field outputPosition cannot be 0, because 0 position is Tag
					return "", fmt.Errorf("incorrect position %d by fieldName %s", outputPosition, fieldName)
				}

				if tempLine[outputPosition] == "" { // we must save previous data 
					tempLine[outputPosition] = strings.Repeat(output.ComponentSeparator, fieldInfo.ComponentsNumber - 1) // count separator is ComponentsNumber - 1
				} 
			}
		}

		err := setFieldInOutputRow(input, output, fieldName, &fieldInfo, tempLine, fullMsg)
		if err != nil {
			return "", err
		}
	}

	// Join tempLine by FieldSeparator
	res := strings.Join(tempLine, output.FieldSeparator)

	return res, nil 
} 


// setFieldInOutputRow
//
//
func setFieldInOutputRow(in, out *Modification, fN string, fI *Field, tL []string, fullMsg *Msg) error {

	outputPosition := fI.Position - 1 // we must work with index
	if int(outputPosition) == 0 { // field outputPosition cannot be 0, because 0 position is Tag
		return fmt.Errorf("incorrect position %f by fieldName %s", outputPosition, fN)
	}

	LinkedFields := fI.Linked // get list of avalible field links

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
