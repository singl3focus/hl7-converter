package hl7converter

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

var (
	ErrInvalidConfig = errors.New("validate json has been unsuccessful")
	ErrInvalidJSON   = errors.New("specified modification is not 'map[string]any'")

	ErrModificationNotFound = errors.New("specified modification is not found in config")

	ErrEmptyPositions = errors.New("positions is empty")

	ErrNilModification = errors.New("modification is nil")
	ErrEmptyTags       = errors.New("tags is empty")

	ErrLinkWithoutStartSymbol = errors.New("link without start symbol")
	ErrLinkWithoutEndSymbol   = errors.New("link without end symbol")
)

// Modification struct dor parsing metadata of config
type Modification struct {
	ComponentSeparator    string `json:"component_separator"` // TODO: validate `tags:""`
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

func (m *Modification) Validate() error {
	if m == nil {
		return ErrNilModification
	}

	if len(m.ComponentSeparator) == 0 || len(m.FieldSeparator) == 0 || len(m.LineSeparator) == 0 || len(m.ComponentArrSeparator) == 0 {
		return fmt.Errorf("invalid separators: component=%q component_array=%q field=%q line=%q", m.ComponentSeparator, m.ComponentArrSeparator, m.FieldSeparator, m.LineSeparator)
	}

	if len(m.TagsInfo.Tags) == 0 {
		return ErrEmptyTags
	}

	if len(m.TagsInfo.Positions) == 0 {
		return ErrEmptyPositions
	}

	for posKey, tagName := range m.TagsInfo.Positions {
		if _, err := strconv.Atoi(posKey); err != nil {
			return fmt.Errorf("positions key %q is not integer: %w", posKey, err)
		}

		if tagName == "" {
			return fmt.Errorf("positions value is empty for key %q", posKey)
		}

		if _, ok := m.TagsInfo.Tags[tagName]; !ok {
			return fmt.Errorf("positions references unknown tag %q", tagName)
		}
	}

	for tagName, tag := range m.TagsInfo.Tags {
		if tag.FieldsNumber != ignoredFieldsNumber && tag.FieldsNumber < 0 {
			return fmt.Errorf("tag %s has invalid fields_number %d", tagName, tag.FieldsNumber)
		}

		for _, opt := range tag.Options {
			if _, ok := mapOfOptions[opt]; !ok {
				return NewErrUndefinedOption(opt, tagName)
			}
		}

		if tag.Tempalate != "" && tag.FieldsNumber > 0 {
			fields := strings.Split(tag.Tempalate, m.FieldSeparator)
			if len(fields) != tag.FieldsNumber {
				return NewErrWrongFieldsNumber(tagName, &tag, len(fields))
			}
		}
	}

	return nil
}

func (m *Modification) OrderedPositionTags() ([]string, error) {
	if len(m.TagsInfo.Positions) == 0 {
		return nil, ErrEmptyPositions
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

	stLinkIndx, endLinkIndx := -1, -1

	for i, v := range str {
		if string(v) == linkElemSt {
			stLinkIndx = i
		} else if string(v) == linkElemEnd {
			endLinkIndx = i
		}

		if endLinkIndx > stLinkIndx {
			if stLinkIndx == -1 {
				return nil, NewError(ErrLinkWithoutStartSymbol, true,
					fmt.Sprintf("field with link %s, link place %s", str, str[:endLinkIndx+1]))
			}

			for j := stLinkIndx; j < endLinkIndx; j++ { // marking previous fields
				mask[j] = itLink
			}

			mask = append(mask, itLink) // marking that field with index endLinkIndx+1

			stLinkIndx, endLinkIndx = -1, -1
		} else {
			mask = append(mask, itSymbol)
		}
	}

	if stLinkIndx > endLinkIndx {
		return nil, NewError(ErrLinkWithoutEndSymbol, true,
			fmt.Sprintf("field with link %s, link place %s", str, str[stLinkIndx:]))
	}

	return mask, nil
}

//TODO: ADD ALIASES TO MINDTAY_HBL

// ReadJSONConfigBlock checking valid config, then read config, find specified block and umarshal it to Modification.
// Function accepts arguments: config path, name needed json block.
func ReadJSONConfigBlock(p, bN string) (*Modification, error) {
	ok, err := ValidateJSONConfig(p)
	if !ok || err != nil {
		return nil, err
	}

	dataFile, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}

	objMap := make(map[string]any)
	err = json.Unmarshal(dataFile, &objMap)
	if err != nil {
		return nil, err
	}

	v, ok := objMap[bN] // Get needed blockName from map
	if !ok {
		return nil, NewError(ErrModificationNotFound, true, fmt.Sprintf("modification %s", bN))
	}

	dataBlock, ok := v.(map[string]any) // Check type blockName
	if !ok {
		return nil, NewError(ErrInvalidJSON, true, fmt.Sprintf("value %v", v))
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

	if err := obj.Validate(); err != nil {
		return nil, err
	}

	return &obj, nil
}

func ValidateJSONConfig(p string) (bool, error) {
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
