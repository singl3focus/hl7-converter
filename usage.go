package hl7converter

import (
	"fmt"
	"strings"
)

// ConverterParams
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
func Convert(p *ConverterParams, msg []byte, cfgWithPositions bool) (*Result, *WrapperConverter, error) {
	var c *Converter
	var err error

	if cfgWithPositions {
		c, err = NewConverter(p.InMod, p.OutMod, WithUsingPositions())
	} else {
		c, err = NewConverter(p.InMod, p.OutMod)
	}
	if err != nil {
		return nil, nil, err
	}

	res, err := c.Convert(msg)
	if err != nil {
		return nil, nil, err
	}

	wrapper := NewWrapperConverter(c)
	
	return res, wrapper, nil
}

// IndetifyMsg
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