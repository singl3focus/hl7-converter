package hl7converter

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// ConvertMsg
func ConvertMsg(input, output *Modification, msg []byte) ([]byte, error) {
	msgBlocks := strings.Split(string(msg), input.LineSeparator)
	
	var res []string
	for _, row := range msgBlocks {
		rowFields := strings.Split(row, input.FieldSeparator) // uneffective
		tag := rowFields[0]
		if len(tag) >= len(row) {
			return nil, fmt.Errorf("identify msg has been unsuccessful") 
		}

		tagInfo, ok := input.Tags[tag]
		if !ok {
			return nil, fmt.Errorf("tag %s not found in input modification %v", tag, input) 
		}

		var matchingTag string
		for _, linkedTag := range tagInfo.Linked {
			_, ok := output.Tags[linkedTag]
			if ok {
				matchingTag = linkedTag
				break	
			}			
		}

		if matchingTag == "" {
			// return nil, fmt.Errorf("linked tag in tag %s not found in output modification %v", tag, output)
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


func assembleOutputRowMsg(input, output *Modification, rowFields []string, inputTag, matchingTag string) (string, error) {
	fieldNumber := output.Tags[matchingTag].FieldsNumber
	outputFields := output.Tags[matchingTag].Fields

	tempLine := make([]string, fieldNumber) // temp slice was filled by default value:""
	tempLine[0] = matchingTag // first position is always placed by Tag 

	for fieldName, fieldInfo := range outputFields { // go through the fields of the output structure
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
// WARNING: Now we haven't relisation fill of output float field 
func setFieldInOutputRow(in, out *Modification, fN string, fI *Field, tL, rF []string, inT string) error {
	inputFields := in.Tags[inT].Fields

	// WE MUST SET ONE FIELD, BUT WE SET ALL, because we received a pointer to tempLine that is gradually being filled
	tempFloatPostiton := make(map[float64]string) // every match position will be collect to one string value and added to Line

	outputPosition := fI.Position - 1 
	if int(outputPosition) == 0 {
		return fmt.Errorf("incorrect position %f by fieldName %s", outputPosition, fN)
	}

	LinkedFields := fI.Linked // get list of avalible field links
	if len(LinkedFields) <= 1 { // If count linked fields is 1 
		if len(LinkedFields) == 0 { // Linked Fields not specified
			err := setDefaultValueToField(in, fN, fI, tL, tempFloatPostiton)
			if err != nil { return err }
		} else {
			rowInfo, ok := inputFields[LinkedFields[0]]
			if !ok { // if we have not found the field, we set the default value
				err := setDefaultValueToField(in, fN, fI, tL, tempFloatPostiton)
				if err != nil { return err }
			} else { // if we have found the field, we set it
				err := setValueToField(in, fI, &rowInfo, tL, rF, tempFloatPostiton)
				if err != nil { // if set value has been unsecseful try set default
					err := setDefaultValueToField(in, fN, fI, tL, tempFloatPostiton)
					if err != nil { return err }
				}
			}
		}

	} else { // If count linked fields is more than 1
		err := setValueToFieldWithMoreLinks(in, out, fI, inputFields, tL, rF, tempFloatPostiton)
		if err != nil { return err }
	}

	// Join tempFloatPostiton
	allKeys := make([]float64, 0, 10)
	for outputFloatPos, _ := range tempFloatPostiton { 
		allKeys = append(allKeys, outputFloatPos)
	}
	sort.Float64s(allKeys) // sorted float keys(position) [1.1, 1.2, 1.8]
	
	// line := tempFloatPostiton[allKeys[0]] // get first value of float positions
	// for index, numbFloatPos := range allKeys[1:] {
	// 	if int(numbFloatPos) == int(allKeys[index-1]) { // check that 1.1 and 1.8 belong to the same position
	// 		line += output.ComponentSeparator
	// 		line += tempFloatPostiton[numbFloatPos]
	// 	} else {
	// 		line = ""
	// 		line = ""
	// 	}
	// }

	return nil
}

func setDefaultValueToField(in *Modification, fN string, fI *Field, tL []string, tFP map[float64]string) error {
	outputPosition := fI.Position - 1 // cannot be equal to 0
	
	defaultValue := fI.DefaultValue
	if defaultValue != "" {
		if isInt(outputPosition) { // Check that postion int or float
			pos := int(outputPosition)
			tL[pos] = defaultValue
		} else {
			tFP[outputPosition] = defaultValue
		}
	} else {
		return fmt.Errorf("fieldName %s not found in input modification %v and default value is not specified", fN, in)
	}

	return nil
}

// setValueToField
func setValueToField(in *Modification, fI, rI *Field, tL, rF []string, tFP map[float64]string) error {
	outputPosition := fI.Position - 1 // cannot be equal to 0
	inputPosition := rI.Position - 1 

	// Check that postion int or float
	if isInt(inputPosition) && isInt(outputPosition) { // if inputPosition and outputPosition Int
		posOu := int(outputPosition)
		posInp := int(inputPosition)

		tL[posOu] = rF[posInp]	

	} else if isInt(inputPosition) { // if inputPosition Int
		posInp := int(inputPosition)

		tFP[outputPosition] = rF[posInp]	

	} else if isInt(outputPosition) { // if outputPosition Int
		posInp := int(inputPosition) // round to less number (3.8 -> 3)
		subposInp := getTenth(inputPosition) // position in field 
		posOu := int(outputPosition)

		val, err := getFieldComponent(in, rF, posInp, subposInp)
		if err != nil || val == "" {
			return err
		}

		rF[posOu] = val

	} else { // if inputPosition and outputPosition NOT Int
		posInp := int(inputPosition) // round to less number (11.2 -> 11)
		subposInp := getTenth(inputPosition) // position in field 

		val, err := getFieldComponent(in, rF, posInp, subposInp)
		if err != nil || val == "" {
			return err
		}

		tFP[outputPosition] = val
	}

	return nil
}

// setValueToFieldWithMoreLinks
func setValueToFieldWithMoreLinks(in, out *Modification, fI *Field, iF map[string]Field,  tL, rF []string, tFP map[float64]string) error {
	outputPosition := fI.Position - 1 // cannot be equal to 0
	linkedFields := fI.Linked
	lenLinked := len(linkedFields)

	if isInt(outputPosition) {
		line := ""
		for i, inputField := range linkedFields { // Len Linked fields more than 1
			inpPos := iF[inputField].Position

			posInp := int(iF[inputField].Position) - 1 // BE CAREFUL WITH Position
			if isInt(inpPos) {
				line += rF[posInp]
			} else {
				subposInp := getTenth(inpPos)
				val, err := getFieldComponent(in, rF, posInp, subposInp)
				if err != nil || val == "" {
					return err
				}

				line += val
			}
			

			if i != (lenLinked - 1) {
				line += out.ComponentSeparator
			}  
		}

		posOu := int(outputPosition)
		tL[posOu] = line

	} else {
		line := ""
		for i, inputField := range linkedFields[1:] { // Len Linked fields more than 1
			posInp := int(iF[inputField].Position) // Don't work with float linked 
			line += rF[posInp]
			if i != (lenLinked - 1) {
				line += out.ComponentSeparator
			}  
		}

		tFP[outputPosition] = line
	}

	return nil
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

