package hl7converter

import (
	"fmt"
	"math"
	"strings"
)


// ConvertMsg
//
// WARNING: message cannot contain a field separator, otherwise the conversion will be incorrect
// NOTE: split logic of separating and remain only logic of 'for'
func ConvertMsg(input, output *Modification, msg []byte) ([]byte, error) {
	msgBlocks := strings.Split(string(msg), input.LineSeparator)
	
	var res []string
	for _, row := range msgBlocks {
		rowFields := strings.Split(row, input.FieldSeparator)
		tag := rowFields[0]
		if len(tag) >= len(row) {
			return nil, fmt.Errorf("identify msg has been unsuccessful") 
		}

		tagInfo, ok := input.Tags[tag]
		if !ok {
			return nil, fmt.Errorf("tag %s not found in input modification %v", tag, input) 
		}


		var matchingTag string // Linked tags hasn't property of json "omitempty"
		for _, linkedTag := range tagInfo.Linked {
			_, ok := output.Tags[linkedTag]
			if ok {
				matchingTag = linkedTag
				break	
			}			
		}

		if matchingTag == "" {
			// return nil, fmt.Errorf("linked tag in tag %s not found in output modification %v", tag, output)

			// NEED NOTIFICATION THAT MSG WAS SKIPPED
			continue 
		}


		outputLine, err := assembleOutputRowMsg(input, output, rowFields, tag, matchingTag)
		if err != nil {
			return nil, err
		}		

		res = append(res, outputLine)
	}

	out := strings.Join(res, output.LineSeparator)

	return []byte(out), nil
}


// assembleOutputRowMsg
//
//
func assembleOutputRowMsg(input, output *Modification, rowFields []string, inputTag, matchingTag string) (string, error) {	
	outputFields := output.Tags[matchingTag].Fields

	tempLine := make([]string, output.Tags[matchingTag].FieldsNumber) // temp slice was initially filled by default value:""
	tempLine[0] = matchingTag // first position is always placed by Tag 

	for fieldName, fieldInfo := range outputFields { // go through the fields of the output structure

		if fieldInfo.ComponentsNumber > 0 { // min. count of components is 2
			if fieldInfo.ComponentsNumber == 1 {
				return "", fmt.Errorf("Commponent count can be equal to 0 or more than 1, else it's field hasn't have components")
			} else {
				outputPosition := int(fieldInfo.Position) - 1 
				if int(outputPosition) == 0 { // field outputPosition cannot be 0, because 0 position is Tag
					return "", fmt.Errorf("incorrect position %d by fieldName %s", outputPosition, fieldName)
				}

				if tempLine[outputPosition] == "" { // we must save previous data
					tempLine[outputPosition] = strings.Repeat(output.ComponentSeparator, fieldInfo.ComponentsNumber - 1) // count separator is ComponentsNumber - 1
				} 
			}
		}

		err := setFieldInOutputRow(input, output, fieldName, &fieldInfo, tempLine, rowFields, inputTag)
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
func setFieldInOutputRow(in, out *Modification, fN string, fI *Field, tL, rF []string, inT string) error {
	inputFields := in.Tags[inT].Fields

	outputPosition := fI.Position - 1 
	if int(outputPosition) == 0 { // field outputPosition cannot be 0, because 0 position is Tag
		return fmt.Errorf("incorrect position %f by fieldName %s", outputPosition, fN)
	}

	LinkedFields := fI.Linked // get list of avalible field links

	if len(LinkedFields) <= 1 { // If count linked fields less than or equal to 1 

		if len(LinkedFields) == 0 { // Linked Fields not specified
			err := setDefaultValueToField(in, out, fN, fI, tL)
			if err != nil {
				return err
			}

		} else {
			rowInfo, ok := inputFields[LinkedFields[0]]

			if !ok { // if we have not found the field, we set the default value
				err := setDefaultValueToField(in, out, fN, fI, tL)
				if err != nil {
					return err
				}

			} else { // if we have found the field, we set it
				err := setValueToField(in, out, fI, &rowInfo, tL, rF)

				if err != nil { // if set value has been unsecseful try set default
					err := setDefaultValueToField(in, out, fN, fI, tL)
					if err != nil {
						return err
					}
				}
			}
		}

	} else { // If count linked fields is more than 1
		err := setValueToFieldWithMoreLinks(in, out, fI, inputFields, tL, rF)
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
		return fmt.Errorf("fieldName %s not found in input modification %v and default value is not specified", fN, in)
	}

	return nil
}


// setValueToField
//
//
func setValueToField(in, out  *Modification, fI, rI *Field, tL, rF []string) error {
	outputPosition := fI.Position - 1 // cannot be equal to 0
	inputPosition := rI.Position - 1 // was there a check?????

	// Check that postion int or float
	if isInt(inputPosition) && isInt(outputPosition) { // if inputPosition and outputPosition Int
		posOu := int(outputPosition)
		posInp := int(inputPosition)

		tL[posOu] = rF[posInp]	

	} else if isInt(inputPosition) { // if inputPosition Int
		posInp := int(inputPosition)


		setFieldComponent(out, tL, int(outputPosition), getTenth(outputPosition), rF[posInp])

	} else if isInt(outputPosition) { // if outputPosition Int
		posInp := int(inputPosition) // round to less number (3.8 -> 3)
		subposInp := getTenth(inputPosition) // position in field 
		posOu := int(outputPosition)

		val, err := getFieldComponent(in, rF, posInp, subposInp)
		if err != nil || val == "" {
			return err
		}

		tL[posOu] = val

	} else { // if inputPosition and outputPosition NOT Int
		posInp := int(inputPosition) // round to less number (11.2 -> 11)
		subposInp := getTenth(inputPosition) // position in field 

		val, err := getFieldComponent(in, rF, posInp, subposInp)
		if err != nil || val == "" {
			return err
		}

		setFieldComponent(out, tL, int(outputPosition), getTenth(outputPosition), val)
	}

	return nil
}


// setValueToFieldWithMoreLinks
func setValueToFieldWithMoreLinks(in, out *Modification, fI *Field, iF map[string]Field,  tL, rF []string) error {
	outputPosition := fI.Position - 1 // cannot be equal to 0
	linkedFields := fI.Linked
	lenLinked := len(linkedFields)

	if isInt(outputPosition) {
		line := ""
		for i, inputField := range linkedFields { // Len Linked fields more than 1
			inpPos := iF[inputField].Position // only for check

			inpPosInt := int(iF[inputField].Position) - 1 // represantation of index, for putting
			if isInt(inpPos) {
				line += rF[inpPosInt]

			} else {
				subposInp := getTenth(inpPos)
				val, err := getFieldComponent(in, rF, inpPosInt, subposInp)
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
			inpPos := iF[inputField].Position // only for check

			inpPosInt := int(iF[inputField].Position) - 1 // represantation of index, for putting
			if isInt(inpPos) {
				line += rF[inpPosInt]

			} else {
				subposInp := getTenth(inpPos)
				val, err := getFieldComponent(in, rF, inpPosInt, subposInp)
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
func getFieldComponent(in *Modification, rF []string, posInp, subposInp int) (string, error) {
	val := strings.Split(rF[posInp], in.ComponentSeparator)[subposInp - 1] // Be careful with subpos
	if len(val) >= len(rF[posInp]) {
		return "", fmt.Errorf("incorrect field %s, the field component could not be pulled out", rF[posInp])
	}

	return val, nil
}



// isInt return that number(float64) is Int or not
func isInt(numb float64) bool {
	return math.Mod(numb, 1.0) == 0
}

// getTenth return tenth of number(float64) 
func getTenth(numb float64) int {
	return int(numb * 10) % 10
}

