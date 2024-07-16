package hl7converter

import (
	"fmt"
	"log"
)

// ConvertWithConverter function
//
// It's just an example of how convert some data, you can use it as the basics for your conversion
func ConvertWithConverter(schemaPath, cfgPath, cfgInBlockName, cfgOutBlockName string, msg []byte, format string) ([][]string, string, error) {
	var inputModification *Modification
	var outputModification *Modification
	var err error


	switch format {
	case "json": 
		inputModification, err = ReadJSONConfigBlock(schemaPath, cfgPath, cfgInBlockName)
		if err != nil || inputModification == nil {
			log.Fatal(err)
		}

		outputModification, err = ReadJSONConfigBlock(schemaPath, cfgPath, cfgOutBlockName)
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

