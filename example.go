package hl7converter

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// FullConvertMsg read config, load input and output Modification
// and also converting msg
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
