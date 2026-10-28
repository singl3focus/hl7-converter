package hl7converter

import (
	"errors"
	"fmt"
	"strings"
)

// Diagnostic is a stable machine-readable error summary. It omits message payloads.
type Diagnostic struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorCode returns a stable category for errors from the converter API.
func ErrorCode(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrAmbiguousMessageType):
		return "ambiguous_type"
	case errors.Is(err, ErrMessageTypeNotFound):
		return "type_not_found"
	case errors.Is(err, ErrInvalidConfig), errors.Is(err, ErrModificationNotFound), errors.Is(err, ErrNilModification):
		return "invalid_config"
	case errors.Is(err, ErrInvalidLink), errors.Is(err, ErrInvalidLinkElems):
		return "invalid_link"
	case errors.Is(err, ErrInvalidParseMsg), errors.Is(err, ErrParseFailure), errors.Is(err, ErrUndefinedScannerFailure):
		return "parse_error"
	case errors.Is(err, ErrUndefinedInputTag):
		return "undefined_input_tag"
	case errors.Is(err, ErrOutputTagNotFound):
		return "output_tag_not_found"
	case errors.Is(err, ErrWrongComponentLink), errors.Is(err, ErrWrongComponentsNumber):
		return "component_error"
	case errors.Is(err, ErrTooBigIndex):
		return "row_index_error"
	case errors.Is(err, ErrAliasInvalidLinkPosition), errors.Is(err, ErrAliasLinkTagNotExists):
		return "alias_error"
	case errors.Is(err, ErrWrongFieldsNumber):
		return "field_count_error"
	case errors.Is(err, FatalErrOfConverting):
		return "conversion_panic"
	default:
		return "error"
	}
}

// Diagnose returns a safe error category without copying the input message.
func Diagnose(err error) Diagnostic {
	code := ErrorCode(err)
	messages := map[string]string{
		"ambiguous_type":       "more than one configured message type matches",
		"type_not_found":       "no configured message type matches",
		"invalid_config":       "configuration validation failed",
		"invalid_link":         "a field link is invalid",
		"parse_error":          "message parsing failed",
		"undefined_input_tag":  "an input tag is not configured",
		"output_tag_not_found": "an input tag has no configured output mapping",
		"component_error":      "a referenced component is unavailable",
		"row_index_error":      "a referenced row is unavailable",
		"alias_error":          "an alias could not be resolved",
		"field_count_error":    "a row has an invalid field count",
		"conversion_panic":     "conversion encountered an internal error",
		"error":                "operation failed",
	}
	return Diagnostic{Code: code, Message: messages[code]}
}

// MappingExplanation describes static dependencies without reading a message.
type MappingExplanation struct {
	Mode    string           `json:"mode"`
	Tags    []TagExplanation `json:"tags"`
	Aliases Aliases          `json:"aliases,omitempty"`
}

// TagExplanation describes an output tag and its field expressions.
type TagExplanation struct {
	Tag    string             `json:"tag"`
	Linked string             `json:"linked,omitempty"`
	Fields []FieldExplanation `json:"fields"`
}

// FieldExplanation uses one-based row positions, including the tag at position 1.
type FieldExplanation struct {
	Position   int      `json:"position"`
	Expression string   `json:"expression"`
	Links      []string `json:"links,omitempty"`
	Default    *string  `json:"default,omitempty"`
}

// ExplainMapping validates a pair and describes its output mappings.
func ExplainMapping(p *ConverterParams, positions bool) (*MappingExplanation, error) {
	if p == nil {
		return nil, fmt.Errorf("%w: nil params", ErrInvalidConfig)
	}
	if err := validateConversionPair(p.InputModification, p.OutputModification); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidConfig, err)
	}
	names := sortedKeys(p.OutputModification.TagsInfo.Tags)
	mode := "input"
	if positions {
		var err error
		names, err = p.OutputModification.OrderedPositionTags()
		if err != nil {
			return nil, err
		}
		mode = "positions"
	} else {
		used := make(map[string]bool)
		for _, tag := range p.InputModification.TagsInfo.Tags {
			used[tag.Linked] = true
		}
		filtered := make([]string, 0, len(names))
		for _, name := range names {
			if used[name] {
				filtered = append(filtered, name)
			}
		}
		names = filtered
	}
	result := &MappingExplanation{Mode: mode, Tags: make([]TagExplanation, 0, len(names)), Aliases: make(Aliases, len(p.OutputModification.Aliases))}
	for name, link := range p.OutputModification.Aliases {
		result.Aliases[name] = link
	}
	for _, name := range names {
		tag := p.OutputModification.TagsInfo.Tags[name]
		explained := TagExplanation{Tag: name, Linked: tag.Linked}
		fields := strings.Split(tag.Tempalate, p.OutputModification.FieldSeparator)
		for i, value := range fields {
			field := FieldExplanation{Position: i + 1, Expression: value}
			if i > 0 {
				parts, err := splitTemplateField(value)
				if err != nil {
					return nil, err
				}
				if parts.defaultValue != "" {
					v := parts.defaultValue
					field.Default = &v
				}
				links, err := extractTemplateLinks(parts.template)
				if err != nil {
					return nil, err
				}
				for _, link := range links {
					field.Links = append(field.Links, link.Raw)
				}
			}
			explained.Fields = append(explained.Fields, field)
		}
		result.Tags = append(result.Tags, explained)
	}
	return result, nil
}
