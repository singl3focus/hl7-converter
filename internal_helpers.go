package hl7converter

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type optionHandler func(fields []string, tagInfo Tag) ([]string, error)

type optionSpec struct {
	description string
	apply       optionHandler
}

var optionRegistry = map[string]optionSpec{
	"autofill": {
		description: "automatically append empty fields until the parsed row matches fields_number",
		apply: func(fields []string, tagInfo Tag) ([]string, error) {
			newFields := append([]string(nil), fields...)
			diff := (tagInfo.FieldsNumber - 1) - len(newFields)
			for i := 0; i < diff; i++ {
				newFields = append(newFields, "")
			}

			return newFields, nil
		},
	},
}

type templateFieldParts struct {
	template     string
	defaultValue string
}

type linkRef struct {
	Raw       string
	Tag       string
	Position  string
	Field     int
	Component int
}

func supportedOptionsSummary() string {
	names := make([]string, 0, len(optionRegistry))
	for name := range optionRegistry {
		names = append(names, name)
	}
	sort.Strings(names)

	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, fmt.Sprintf("%s: %s", name, optionRegistry[name].description))
	}

	return strings.Join(parts, ", ")
}

func validateTagOptions(tagName string, tag Tag) error {
	for _, opt := range tag.Options {
		if _, ok := optionRegistry[opt]; !ok {
			return NewErrUndefinedOption(opt, tagName)
		}
	}

	return nil
}

func applyTagOptions(tagName string, fields []string, tagInfo Tag) ([]string, error) {
	newFields := append([]string(nil), fields...)

	for _, option := range tagInfo.Options {
		spec, ok := optionRegistry[option]
		if !ok {
			return nil, NewErrUndefinedOption(option, tagName)
		}

		var err error
		newFields, err = spec.apply(newFields, tagInfo)
		if err != nil {
			return nil, err
		}
	}

	return newFields, nil
}

func splitTemplateField(field string) (templateFieldParts, error) {
	parts := strings.Split(field, OR)

	switch len(parts) {
	case 1:
		return templateFieldParts{template: parts[0]}, nil
	case 2:
		if parts[0] != "" && parts[1] != "" {
			return templateFieldParts{}, fmt.Errorf("field %q cannot contain template and default simultaneously", field)
		}
		if parts[0] == "" && parts[1] == "" {
			return templateFieldParts{}, NewErrEmptyDefaultValue(field)
		}
		if parts[0] == "" {
			if parts[1] == "" {
				return templateFieldParts{}, NewErrEmptyDefaultValue(field)
			}

			return templateFieldParts{defaultValue: parts[1]}, nil
		}

		return templateFieldParts{template: parts[0]}, nil
	default:
		return templateFieldParts{}, NewErrWrongParamCount(field, OR)
	}
}

func extractTemplateLinks(template string) ([]linkRef, error) {
	if template == "" {
		return nil, nil
	}

	if _, err := TempalateParse(template); err != nil {
		return nil, err
	}

	refs := make([]linkRef, 0, strings.Count(template, linkElemSt))
	for i := 0; i < len(template); i++ {
		if string(template[i]) != linkElemSt {
			continue
		}

		end := strings.Index(template[i:], linkElemEnd)
		if end == -1 {
			return nil, NewErrInvalidLink(template[i:])
		}

		raw := template[i+1 : i+end]
		ref, err := parseLinkRef(raw)
		if err != nil {
			return nil, err
		}

		if ref.Field < 2 {
			return nil, NewErrInvalidLink(raw)
		}
		refs = append(refs, ref)
		i += end
	}

	return refs, nil
}

func parseLinkRef(link string) (linkRef, error) {
	tag, position, ok := strings.Cut(link, linkToField)
	if !ok || strings.Contains(position, linkToField) {
		return linkRef{}, NewErrInvalidLink(link)
	}
	if tag == "" || position == "" {
		return linkRef{}, NewErrInvalidLinkElems(link)
	}
	parts := strings.Split(position, ".")
	if len(parts) > 2 {
		return linkRef{}, NewErrInvalidLink(link)
	}
	field, err := positivePosition(parts[0])
	if err != nil {
		return linkRef{}, NewErrInvalidLink(link)
	}
	component := 0
	if len(parts) == 2 {
		component, err = positivePosition(parts[1])
		if err != nil {
			return linkRef{}, NewErrInvalidLink(link)
		}
	}
	return linkRef{Raw: link, Tag: tag, Position: position, Field: field, Component: component}, nil
}

func positivePosition(value string) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("empty position")
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("position must contain decimal digits")
		}
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("position must be a positive integer")
	}
	return n, nil
}

func validateTemplateSyntax(template, fieldSeparator string, fieldsNumber int) error {
	if template == "" {
		return nil
	}

	fields := strings.Split(template, fieldSeparator)
	if fieldsNumber != ignoredFieldsNumber && fieldsNumber > 0 && len(fields) != fieldsNumber {
		return fmt.Errorf("template %q: %w", template, NewErrWrongFieldsNumber(fields[0], &Tag{FieldsNumber: fieldsNumber, Tempalate: template}, len(fields)))
	}

	for _, field := range fields {
		parts, err := splitTemplateField(field)
		if err != nil {
			return err
		}

		if parts.template == "" {
			continue
		}

		if _, err := extractTemplateLinks(parts.template); err != nil {
			return err
		}
	}

	return nil
}

func referencedTemplateLinks(template, fieldSeparator string) ([]linkRef, error) {
	if template == "" {
		return nil, nil
	}

	fields := strings.Split(template, fieldSeparator)
	refs := make([]linkRef, 0)
	for _, field := range fields {
		parts, err := splitTemplateField(field)
		if err != nil {
			return nil, err
		}

		fieldRefs, err := extractTemplateLinks(parts.template)
		if err != nil {
			return nil, err
		}

		refs = append(refs, fieldRefs...)
	}

	return refs, nil
}

func validateConversionPair(inputMod, outputMod *Modification) error {
	if err := inputMod.Validate(); err != nil {
		return fmt.Errorf("input: %w", err)
	}
	if err := outputMod.Validate(); err != nil {
		return fmt.Errorf("output: %w", err)
	}
	emitted := make(map[string]bool)
	for _, name := range outputMod.TagsInfo.Positions {
		emitted[name] = true
	}
	for _, tag := range inputMod.TagsInfo.Tags {
		emitted[tag.Linked] = true
	}
	for _, outputTagName := range sortedKeys(outputMod.TagsInfo.Tags) {
		outputTag := outputMod.TagsInfo.Tags[outputTagName]
		if emitted[outputTagName] && outputTag.FieldsNumber < 1 {
			return fmt.Errorf("output tag %q: fields_number must be positive", outputTagName)
		}
		if emitted[outputTagName] && outputTag.Tempalate != "" {
			fields := strings.Split(outputTag.Tempalate, outputMod.FieldSeparator)
			if len(fields) != outputTag.FieldsNumber || fields[0] != outputTagName {
				return fmt.Errorf("output tag %q: template must start with the tag and match fields_number", outputTagName)
			}
		}
		if outputTag.Linked != "" {
			if _, ok := inputMod.TagsInfo.Tags[outputTag.Linked]; !ok {
				return fmt.Errorf("output tag %q links to unknown input tag %q", outputTagName, outputTag.Linked)
			}
		}

		refs, err := referencedTemplateLinks(outputTag.Tempalate, outputMod.FieldSeparator)
		if err != nil {
			return fmt.Errorf("output tag %q template validation failed: %w", outputTagName, err)
		}

		for _, ref := range refs {
			inputTag, ok := inputMod.TagsInfo.Tags[ref.Tag]
			if !ok {
				return fmt.Errorf("output tag %q references unknown input tag %q", outputTagName, ref.Tag)
			}

			if inputTag.FieldsNumber == ignoredFieldsNumber {
				continue
			}

			fieldIndex := ref.Field - 2

			if fieldIndex < 0 || fieldIndex >= inputTag.FieldsNumber-1 {
				return fmt.Errorf(
					"output tag %q references field position %q outside input tag %q fields_number %d",
					outputTagName,
					ref.Position,
					ref.Tag,
					inputTag.FieldsNumber,
				)
			}
		}
	}

	return nil
}
