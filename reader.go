package hl7converter

import (
	"os"
	"fmt"
	"encoding/json"

	"gopkg.in/yaml.v2"
	"github.com/xeipuuv/gojsonschema"
)

// readJSONConfig
//
// param::p it's config path
// param::bТ it's block name (name needed json block)
func ReadJSONConfigBlock(schemaPath, configPath, bN string) (*Modification, error) {
	validateJSONConfig(schemaPath, configPath)

	dataFile, err := os.ReadFile(configPath) // Reading config file
	if err != nil {
		return nil, err
	}


	objMap := make(map[string]any)
	err = json.Unmarshal(dataFile, &objMap) // Unmrashal config data to map
	if err != nil {
		return nil, err
	}


	value, ok := objMap[bN] // Get needed blockName from map
	if !ok {
		return nil, fmt.Errorf("specified modification not found in config")
	}

	
	dataBlock, ok  := value.(map[string]any) // Check type blockName
	if !ok {
		return nil, fmt.Errorf("specified modification is not 'map[string]any'")
	}
	
		
	jsonData, err := json.Marshal(dataBlock) // Marshal block data in order to convert block to needed structure 
    if err != nil {
        return nil, nil
    }

	var obj Modification
	err = json.Unmarshal(jsonData, &obj) // Unmarshal block data to convert to needed structure
	if err != nil {
		return nil, err
	}

	return &obj, nil
}


func validateJSONConfig(schemaPath, configPath string) {

	schPath := "file:///" + schemaPath
	cfgPath := "file:///" + configPath

	schemaLoader := gojsonschema.NewReferenceLoader(schPath)
    documentLoader := gojsonschema.NewReferenceLoader(cfgPath)

    result, err := gojsonschema.Validate(schemaLoader, documentLoader)
    if err != nil {
        fmt.Println(err.Error())
		os.Exit(1)
	}

    if result.Valid() {
        fmt.Println("The document is valid")
    } else {
        fmt.Println("The document is not valid. see errors:")
        for _, desc := range result.Errors() {
            fmt.Printf("- %s\n", desc)
        }

		os.Exit(1)
    }
}






// ReadYAMLConfigBlock
//
// param::p it's config path
// param::bТ it's block name (name needed json block)
func ReadYAMLConfigBlock(p, bN string) (*Modification, error) {
	dataFile, err := os.ReadFile(p) // Reading config file
	if err != nil {
		return nil, err
	}


	objMap := make(map[string]any)
	err = yaml.Unmarshal(dataFile, &objMap) // Unmrashal config data to map
	if err != nil {
		return nil, err
	}


	value, ok := objMap[bN] // Get needed blockName from map
	if !ok {
		return nil, fmt.Errorf("specified modification not found in config")
	}

	dataBlock, ok  := value.(map[any]any) // Check type blockName
	if !ok {
		return nil, fmt.Errorf("specified modification is not 'map[string]any'")
	}
		
	yamlData, err := yaml.Marshal(dataBlock) // Marshal block data in order to convert block to needed structure 
    if err != nil {
        return nil, err
    }

	var obj Modification	
	err = yaml.Unmarshal(yamlData, &obj) // Unmarshal block data to convert to needed structure
	if err != nil {
		return nil, err
	}

	return &obj, nil
}