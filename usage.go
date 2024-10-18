package hl7converter

import (
	"fmt"
	"strings"
)

// ConvertWithConverter function
//
// It's just an example of how convert some data, you can use it as the basics for your conversion
//
// Deprecated: logic of get message and it's type was splitted to two different func,
// use Convert and IndetifyMsg instead.
//
/*
func ConvertWithConverter(cfgPath, cfgInBlockName, cfgOutBlockName string, msg []byte, format string) ([][]string, string, error) {
	var inputModification *Modification
	var outputModification *Modification
	var err error


	switch format {
	case "json": 
		inputModification, err = ReadJSONConfigBlock(cfgPath, cfgInBlockName)
		if err != nil || inputModification == nil {
			log.Fatal(err)
		}

		outputModification, err = ReadJSONConfigBlock(cfgPath, cfgOutBlockName)
		if err != nil || outputModification == nil {
			log.Fatal(err)
		}
	case "yaml": 
		inputModification, err = ReadYAMLConfigBlock(cfgPath, cfgInBlockName)
		if err != nil || inputModification == nil {
			log.Fatal(err)
		}

		outputModification, err = ReadYAMLConfigBlock(cfgPath, cfgOutBlockName)
		if err != nil || outputModification == nil {
			log.Fatal(err)
		}
	default:
		return nil, "", fmt.Errorf("undefined format, receive %s", format)
	}

	c, err := NewConverter(inputModification, outputModification)
	if err != nil {
		return nil, "", err
	}

	msgAndFields, err := c.Convert(msg)
	if err != nil {
		return nil, "", err
	}

	
	msgType, ok := IndetifyMsg(c.InMsg, c.Input)
	if !ok {
		return nil, "", fmt.Errorf("undefined message type, msg: %v", c.InMsg)
	}
	
	return msgAndFields, msgType, nil
}
*/


// Details for FUTURE UPDATING
//
type ConverterParams struct {
	InMod, OutMod *Modification
	LineSplit func(data []byte, atEOF bool) (advance int, token []byte, err error)
}

func NewConverterParams(cfgPath, cfgInBlockName, cfgOutBlockName string) (*ConverterParams, error) {
	if !strings.Contains(cfgPath, ".json") {
		return nil, NewErrInvalidJsonExtension(cfgPath)
	}

	inputModification, err := ReadJSONConfigBlock(cfgPath, cfgInBlockName)
	if err != nil {
		return nil, err
	} else if inputModification == nil {
		return nil, NewErrNilModification(cfgInBlockName, cfgPath)
	}

	outputModification, err := ReadJSONConfigBlock(cfgPath, cfgOutBlockName)
	if err != nil {
		return nil, err
	} else if outputModification == nil {
		return nil, NewErrNilModification(cfgOutBlockName, cfgPath)
	}

	splitByLine := GetCustomSplit(inputModification.LineSeparator)
	
	return &ConverterParams{
		InMod: inputModification,
		OutMod: outputModification,
		LineSplit: splitByLine,
	}, nil
}



// Convert
//
// _______[INFO]_______
// - Can works only with JSON Config.
// - Return splitted message, Converter (for any using) and an error.
//
func Convert(p *ConverterParams, msg []byte) (*Result, *WrapperConverter, error) {
	c, err := NewConverter(p.InMod, p.OutMod)
	if err != nil {
		return nil, nil, err
	}

	msgAndFields, err := c.Convert(msg)
	if err != nil {
		return nil, nil, err
	}

	wrapper := NewWrapperConverter(c)
	
	return msgAndFields, wrapper, nil
}


func IndetifyMsg(p ConverterParams, msg []byte) (string, error) {
	MSG, err := ConvertToMSG(p, msg)
	if err != nil {
		return "", err
	}

	msgType, ok := indetifyMsg(MSG, p.InMod)
	if !ok {
		return "", fmt.Errorf("undefined type, msg: %v", MSG)
	}

	return msgType, nil
}