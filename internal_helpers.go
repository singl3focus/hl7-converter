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
	Raw           string
	Tag           string
	Position      string
	PositionValue float64
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

		refs = append(refs, ref)
		i += end
	}

	return refs, nil
}

func parseLinkRef(link string) (linkRef, error) {
	elems := strings.Split(link, linkToField)
	if len(elems) != 2 {
		return linkRef{}, NewErrInvalidLink(link)
	}

	tag, pos := elems[0], elems[1]
	if tag == "" || pos == "" {
		return linkRef{}, NewErrInvalidLinkElems(link)
	}

	position, err := strconv.ParseFloat(pos, 64)
	if err != nil {
		return linkRef{}, NewErrInvalidLink(link)
	}

	if err := validateLinkPosition(link, position); err != nil {
		return linkRef{}, err
	}

	return linkRef{
		Raw:           link,
		Tag:           tag,
		Position:      pos,
		PositionValue: position,
	}, nil
}

func validateLinkPosition(link string, position float64) error {
	if position < 2 {
		return NewErrInvalidLink(link)
	}

	if isInt(position) {
		return nil
	}

	if getTenth(position) < 1 {
		return NewErrInvalidLink(link)
	}

	return nil
}

func fieldIndexFromPosition(link string, position float64) (int, error) {
	if err := validateLinkPosition(link, position); err != nil {
		return 0, err
	}

	return int(position) - 2, nil
}

func componentIndexFromPosition(link string, position float64) (int, error) {
	if err := validateLinkPosition(link, position); err != nil {
		return 0, err
	}

	if isInt(position) {
		return -1, nil
	}

	return getTenth(position) - 1, nil
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
	for outputTagName, outputTag := range outputMod.TagsInfo.Tags {
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

			fieldIndex, err := fieldIndexFromPosition(ref.Raw, ref.PositionValue)
			if err != nil {
				return fmt.Errorf("output tag %q has invalid link %q: %w", outputTagName, ref.Raw, err)
			}

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
