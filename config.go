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
	// Separators are validated for non-emptiness and mutual non-overlap in Modification.Validate.
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

func (m *Modification) Validate() error {
	if m == nil {
		return ErrNilModification
	}

	if err := validateSeparators(m); err != nil {
		return err
	}

	if len(m.TagsInfo.Tags) == 0 {
		return ErrEmptyTags
	}

	if len(m.TagsInfo.Positions) == 0 {
		return ErrEmptyPositions
	}

	for _, posKey := range sortedKeys(m.TagsInfo.Positions) {
		tagName := m.TagsInfo.Positions[posKey]
		if err := validatePositionKey(posKey); err != nil {
			return err
		}

		if tagName == "" {
			return fmt.Errorf("positions value is empty for key %q", posKey)
		}

		if _, ok := m.TagsInfo.Tags[tagName]; !ok {
			return fmt.Errorf("positions references unknown tag %q", tagName)
		}
	}

	for _, tagName := range sortedKeys(m.TagsInfo.Tags) {
		tag := m.TagsInfo.Tags[tagName]
		if tag.FieldsNumber != ignoredFieldsNumber && tag.FieldsNumber < 1 {
			return fmt.Errorf("tag %s has invalid fields_number %d", tagName, tag.FieldsNumber)
		}

		if err := validateTagOptions(tagName, tag); err != nil {
			return err
		}

		if err := validateTemplateSyntax(tag.Tempalate, m.FieldSeparator, tag.FieldsNumber); err != nil {
			return fmt.Errorf("tag %s template invalid: %w", tagName, err)
		}
	}

	if err := validateTypes(m); err != nil {
		return err
	}
	for _, name := range sortedKeys(m.Aliases) {
		ref, err := parseLinkRef(m.Aliases[name])
		if err != nil {
			return fmt.Errorf("alias %q: %w", name, err)
		}
		tag, ok := m.TagsInfo.Tags[ref.Tag]
		if !ok {
			return fmt.Errorf("alias %q references unknown tag %q", name, ref.Tag)
		}
		if tag.FieldsNumber > 0 && ref.Field > tag.FieldsNumber {
			return fmt.Errorf("alias %q: field outside tag %q", name, ref.Tag)
		}
	}
	return nil
}

func (m *Modification) OrderedPositionTags() ([]string, error) {
	if m == nil {
		return nil, ErrNilModification
	}
	if len(m.TagsInfo.Positions) == 0 {
		return nil, ErrEmptyPositions
	}

	pos := make([]int, 0, len(m.TagsInfo.Positions))
	for p := range m.TagsInfo.Positions {
		if err := validatePositionKey(p); err != nil {
			return nil, err
		}
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

// TemplateParse marks bytes belonging to links and rejects unmatched or nested brackets.
func TemplateParse(str string) ([]int, error) {
	mask := make([]int, len(str))
	start := -1
	for i := 0; i < len(str); i++ {
		mask[i] = itSymbol
		switch str[i] {
		case '<':
			if start >= 0 {
				return nil, NewErrInvalidLink(str)
			}
			start = i
		case '>':
			if start < 0 {
				return nil, NewError(ErrLinkWithoutStartSymbol, true)
			}
			for j := start; j <= i; j++ {
				mask[j] = itLink
			}
			start = -1
		}
	}
	if start >= 0 {
		return nil, NewError(ErrLinkWithoutEndSymbol, true)
	}
	return mask, nil
}

// TempalateParse is the legacy spelling of TemplateParse.
// Deprecated: use TemplateParse.
func TempalateParse(str string) ([]int, error) { return TemplateParse(str) }

func validatePositionKey(key string) error {
	n, err := positivePosition(key)
	if err != nil || strconv.Itoa(n) != key {
		return fmt.Errorf("positions key %q must be a canonical positive integer", key)
	}
	return nil
}

func sortedKeys[V any](values map[string]V) []string {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func validateTypes(m *Modification) error {
	owners := make(map[string]string)
	for _, name := range sortedKeys(m.Types) {
		if name == "" {
			return fmt.Errorf("empty type name")
		}
		for _, variant := range m.Types[name] {
			tags := append([]string(nil), variant...)
			sort.Strings(tags)
			for i, tag := range tags {
				if _, ok := m.TagsInfo.Tags[tag]; !ok {
					return fmt.Errorf("type %q references unknown tag %q", name, tag)
				}
				if i > 0 && tags[i-1] == tag {
					return fmt.Errorf("type %q repeats tag %q", name, tag)
				}
			}
			key, _ := json.Marshal(tags)
			if owner, ok := owners[string(key)]; ok && owner != name {
				return fmt.Errorf("%w: types %q and %q", ErrAmbiguousMessageType, owner, name)
			}
			owners[string(key)] = name
		}
	}
	return nil
}

func validateSeparators(m *Modification) error {
	if len(m.ComponentSeparator) == 0 || len(m.FieldSeparator) == 0 || len(m.LineSeparator) == 0 || len(m.ComponentArrSeparator) == 0 {
		return fmt.Errorf("invalid separators: component=%q component_array=%q field=%q line=%q", m.ComponentSeparator, m.ComponentArrSeparator, m.FieldSeparator, m.LineSeparator)
	}

	separators := map[string]string{
		"component_separator":       m.ComponentSeparator,
		"component_array_separator": m.ComponentArrSeparator,
		"field_separator":           m.FieldSeparator,
		"line_separator":            m.LineSeparator,
	}
	for leftName, leftValue := range separators {
		for rightName, rightValue := range separators {
			if leftName >= rightName {
				continue
			}
			if leftValue == rightValue {
				return fmt.Errorf("%s conflicts with %s: %q", leftName, rightName, leftValue)
			}
		}
	}

	return nil
}

// ReadJSONConfig reads the schema and validates every modification block.
func ReadJSONConfig(filename string) (map[string]*Modification, error) {
	blocks, err := readConfigBlocks(filename)
	if err != nil {
		return nil, err
	}
	for _, name := range sortedKeys(blocks) {
		if err := blocks[name].Validate(); err != nil {
			return nil, fmt.Errorf("%w: block %q: %w", ErrInvalidConfig, name, err)
		}
	}
	return blocks, nil
}

func readConfigBlocks(filename string) (map[string]*Modification, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	ok, err := validateConfigBytes(data)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidConfig
	}
	var blocks map[string]*Modification
	if err := json.Unmarshal(data, &blocks); err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, fmt.Errorf("%w: no modification blocks", ErrInvalidConfig)
	}
	return blocks, nil
}

// ReadJSONConfigBlock validates the schema and the selected modification.
func ReadJSONConfigBlock(filename, name string) (*Modification, error) {
	blocks, err := readConfigBlocks(filename)
	if err != nil {
		return nil, err
	}
	block, ok := blocks[name]
	if !ok {
		return nil, NewError(ErrModificationNotFound, true, fmt.Sprintf("modification %s", name))
	}
	if err := block.Validate(); err != nil {
		return nil, fmt.Errorf("%w: block %q: %w", ErrInvalidConfig, name, err)
	}
	return block, nil
}

// ValidateJSONConfig checks the JSON schema; ReadJSONConfig also checks semantics.
func ValidateJSONConfig(filename string) (bool, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return false, err
	}
	return validateConfigBytes(data)
}

func validateConfigBytes(data []byte) (bool, error) {
	result, err := gojsonschema.Validate(gojsonschema.NewStringLoader(jsonSchema), gojsonschema.NewBytesLoader(data))
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrInvalidConfig, err)
	}
	if result.Valid() {
		return true, nil
	}
	descriptions := make([]string, 0, len(result.Errors()))
	for _, issue := range result.Errors() {
		descriptions = append(descriptions, issue.Field()+": "+issue.Description())
	}
	sort.Strings(descriptions)
	return false, fmt.Errorf("%w: %s", ErrInvalidConfig, strings.Join(descriptions, "; "))
}
