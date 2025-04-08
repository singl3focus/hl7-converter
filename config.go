package hl7converter

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"strconv"

	"github.com/xeipuuv/gojsonschema"
)

/*
For parsing metadata of config
*/
type Modification struct {
	ComponentSeparator    string `json:"component_separator"`
	ComponentArrSeparator string `json:"component_array_separator"`
	FieldSeparator        string `json:"field_separator"`
	LineSeparator         string `json:"line_separator"`

	Types map[string][][]string `json:"types,omitempty"` // [OPTIONAL]

	Aliases Aliases `json:"aliases,omitempty"` // [OPTIONAL]

	TagsInfo TagsInfo `json:"tags_info"`
}

type Aliases map[string]string

type TagsInfo struct {
	Positions map[string]string `json:"positions"`
	Tags      map[string]Tag    `json:"tags"`
}

type Tag struct {
	Count        int    `json:"-"`
	Linked       string `json:"linked"`
	FieldsNumber int    `json:"fields_number"`
	Tempalate    string `json:"template"`

	Options []string `json:"options,omitempty"` // [OPTIONAL]
}

func (m *Modification) OrderedPositionTags() ([]string, error) {
	if len(m.TagsInfo.Positions) == 0 {
		return nil, errors.New("positions zero") // todo: error definition
	}

	pos := make([]int, 0, len(m.TagsInfo.Positions))
	for p := range m.TagsInfo.Positions {
		k, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		pos = append(pos, k)
	}
	sort.Ints(pos)

	orderedTags := make([]string, 0, len(pos))
	for _, p := range pos {
		orderedTags = append(orderedTags, m.TagsInfo.Positions[strconv.Itoa(p)])
	}

	return orderedTags, nil
}

func TempalateParse(str string) ([]int, error) {
	mask := make([]int, 0, len(str)) // example: [1,1,1,1,0,0,0,1,1,1], 1 - Symbol, 0 - Link
	stLinkIndx, endLinkIndx := 0, 0

	for i, v := range str {
		if string(v) == linkElemSt {
			stLinkIndx = i
		} else if string(v) == linkElemEnd {
			endLinkIndx = i
		}

		if endLinkIndx > stLinkIndx {
			for j := stLinkIndx; j < endLinkIndx; j++ {
				mask[j] = itLink
			}
			mask = append(mask, itLink)
			stLinkIndx, endLinkIndx = 0, 0
		} else {
			mask = append(mask, itSymbol)
		}
	}

	return mask, nil
}

/*_______________________________________[PARSE CONFIG FILE]_______________________________________*/

// readJSONConfig
//
// param::p  - it's config path
// param::bТ - it's block name (name needed json block)
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
		return nil, NewErrModificationNotFound(bN)
	}

	dataBlock, ok := value.(map[string]any) // Check type blockName
	if !ok {
		return nil, NewErrInvalidJSON(value)
	}

	jsonData, err := json.Marshal(dataBlock) // Marshal block data in order to convert block to needed structure
	if err != nil {
		return nil, err
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

	if len(result.Errors()) > 0 {
		var errorStr string

		for i, err := range result.Errors() {
			errorStr += err.Description()

			if i != len(result.Errors())-1 {
				errorStr += "\n"
			}
		}

		return false, errors.New(errorStr)
	}

	return false, ErrInvalidConfig
}
