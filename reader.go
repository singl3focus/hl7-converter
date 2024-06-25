package hl7converter

import (
	"encoding/json"
	"fmt"
	"os"
)

// readJSONConfig
// param::p it's config path
// param::bТ it's block name (name needed json block)
func ReadJSONConfigBlock(p, bN string) (*Modification, error) {
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


