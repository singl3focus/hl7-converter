package hl7converter

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// ConvertWithConverter function
func ConvertWithConverter(cfgName, cfgInBlockName, cfgOutBlockName string, msg []byte) ([][]string, string, error) {
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

	c, err := NewConverter(inputModification, outputModification)
	if err != nil {
		return nil, "", err
	}

	msgAndFields, err := c.Convert(msg)
	if err != nil {
		return nil, "", err
	}

	msgType := IndetifyMsg(c.InMsg, c.Input)
	if msgType == "" {
		return nil, "", fmt.Errorf("undefined message type, msg: %v", c.InMsg)
	}
	
	return msgAndFields, msgType, nil
}

