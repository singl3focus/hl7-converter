package hl7converter

import (
	"os"
	"fmt"
	"log"
	"bufio"
	"bytes"
	"strings"
	"path/filepath"
)

// FullConvertMsg read config, load input and output Modification
// and also converting msg
//
// Reference:
// if you want convert msg withou same tag
// use FullConvertMsg and send the full message to the function
//
// if you want to convert a message with the same tags that you needed,
// split this message so that each sub-message contains all the service tags (non-repeating tags)
// and one of the repeating tags in turn. Example:
// [[]byte("H|1|2|\n" + "R||info|||\n" + "R||13144||||"), []byte("H|1|2|\n" + "R||info|||\n" + "R||13155||||")]
//
// After converting message with the same tags return converted msg which you could be assemble to your needed msg
//
// NOTE: Be careful. In this case, the config must be located in the same directory
// as the file in which you are calling this function
func FullConvertMsg(cfgName, cfgInBlockName, cfgOutBlockName string, msg []byte) ([]byte, error){
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	configPath := filepath.Join(wd, cfgName)

	inputModification, err := ReadJSONConfigBlock(configPath, cfgInBlockName)
	if err != nil || inputModification == nil {
		log.Fatal(err)
	}

	outputModification, err := ReadJSONConfigBlock(configPath, cfgOutBlockName)
	if err != nil || outputModification == nil {
		log.Fatal(err)
	}

	// _______________________________________
	
	LineSplit := GetCustomSplit(inputModification.LineSeparator)
	inputMsg, err := ConvertToMSG(inputModification, msg, LineSplit)
	if err != nil {
		return nil, err
	}	
	
	// _______________________________________
	var out []string

	scanner := bufio.NewScanner(bytes.NewReader(msg))
	scanner.Split(LineSplit)
	
	for scanner.Scan() {
		token := scanner.Text()

		res, err := ConvertRow(inputModification, outputModification, token, inputMsg)
		if err != nil || res == "" {
			return nil, err
		} else if res == "skip" {
			continue
		}

		out = append(out, res)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	convertedMsg := strings.Join(out, outputModification.LineSeparator)

	return []byte(convertedMsg), nil
}


// FullConvertMsgWithSameTags read config, load input and output Modification
// and also converting msg (but in params FullConvertMsgWithSameTags geting msgArr which must contain msgs without same Tags)
//
// Reference:
// if you want convert msg withou same tag
// use FullConvertMsg and send the full message to the function
//
// if you want to convert a message with the same tags that you needed,
// split this message so that each sub-message contains all the service tags (non-repeating tags)
// and one of the repeating tags in turn. Example:
// [[]byte("H|1|2|\n" + "R||info|||\n" + "R||13144||||"), []byte("H|1|2|\n" + "R||info|||\n" + "R||13155||||")]
//
// After converting message with the same tags return converted msg which you could be assemble to your needed msg
//
// NOTE: Be careful. In this case, the config must be located in the same directory
// as the file in which you are calling this function
func FullConvertMsgWithSameTags(cfgName, cfgInBlockName, cfgOutBlockName string, msgArr [][]byte, sameTag string) ([][]byte, error){
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	configPath := filepath.Join(wd, cfgName)

	inputModification, err := ReadJSONConfigBlock(configPath, cfgInBlockName)
	if err != nil || inputModification == nil {
		log.Fatal(err)
	}

	outputModification, err := ReadJSONConfigBlock(configPath, cfgOutBlockName)
	if err != nil || outputModification == nil {
		log.Fatal(err)
	}

	results := make([][]byte, 0, len(msgArr))
	// _______________________________________
	for _, msg := range msgArr {
		LineSplit := GetCustomSplit(inputModification.LineSeparator)
		inputMsg, err := ConvertToMSG(inputModification, msg, LineSplit)
		if err != nil {
			return nil, err
		}	
		
		if _, ok := inputMsg.Tags[sameTag]; !ok {
			return nil, fmt.Errorf("sameTag: '%s' isn't found in input msg: '%s'", sameTag, string(msg))
		}

		// _______________________________________
		
		scanner := bufio.NewScanner(bytes.NewReader(msg))
		scanner.Split(LineSplit)

		for scanner.Scan() {
			token := scanner.Text()

			res, err := ConvertRow(inputModification, outputModification, token, inputMsg)
			if err != nil || res == "" {
				return nil, err
			} else if res == "skip" { 
				continue
			}

			results = append(results, []byte(res))
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}


	return results, nil
}


// GetCustomSplit
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
