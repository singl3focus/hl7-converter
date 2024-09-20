package hl7converter

import (
	"os"
	"encoding/json"

	"github.com/xeipuuv/gojsonschema"
)

// readJSONConfig
//
// param::p it's config path
// param::bТ it's block name (name needed json block)
func ReadJSONConfigBlock(p, bN string) (*Modification, error) {
	ok, err := validateJSONConfig(p)
	if !ok || err != nil {
		return nil, err 
	}

	dataFile, err := os.ReadFile(p) // Reading config file
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
		return nil, ErrModificationNotFound
	}

	
	dataBlock, ok  := value.(map[string]any) // Check type blockName
	if !ok {
		return nil, ErrInvalidJSON
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


func validateJSONConfig(p string) (bool, error) {
	cfgPath := "file:///" + p

	schemaLoader := gojsonschema.NewStringLoader(jsonSchema)
    documentLoader := gojsonschema.NewReferenceLoader(cfgPath)

    result, err := gojsonschema.Validate(schemaLoader, documentLoader)
    if err != nil {
        return false, err
	}

    if result.Valid() {
        return true, nil
    }

	return false, ErrInvalidConfig
}