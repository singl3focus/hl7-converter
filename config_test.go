package hl7converter_test

import (
	"os"
	"path/filepath"
	"testing"

	hl7converter "github.com/singl3focus/hl7-converter/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var workDir string

const testConfigJSON = "config.json"

func init() {
	wd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}

	workDir = wd
}

func TestReadJSONConfigBlock(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(workDir, "examples", testConfigJSON)

	m, err := hl7converter.ReadJSONConfigBlock(configPath, "astm_hbl")
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestModificationValidate(t *testing.T) {
	t.Parallel()

	base := &hl7converter.Modification{
		ComponentSeparator:    "^",
		ComponentArrSeparator: "~",
		FieldSeparator:        "|",
		LineSeparator:         "\n",
		TagsInfo: hl7converter.TagsInfo{
			Positions: map[string]string{"1": "H"},
			Tags: map[string]hl7converter.Tag{
				"H": {
					Linked:       "MSH",
					FieldsNumber: 3,
					Tempalate:    "H|<MSH-2>|??DEFAULT",
				},
			},
		},
	}

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		mod := cloneModification(base)
		assert.NoError(t, mod.Validate())
	})

	t.Run("position key must be integer", func(t *testing.T) {
		t.Parallel()

		mod := cloneModification(base)
		mod.TagsInfo.Positions["bad"] = "H"
		assert.Error(t, mod.Validate())
	})

	t.Run("separators must not conflict", func(t *testing.T) {
		t.Parallel()

		mod := cloneModification(base)
		mod.ComponentArrSeparator = mod.ComponentSeparator
		assert.Error(t, mod.Validate())
	})

	t.Run("unknown option rejected", func(t *testing.T) {
		t.Parallel()

		mod := cloneModification(base)
		tag := mod.TagsInfo.Tags["H"]
		tag.Options = []string{"unknown"}
		mod.TagsInfo.Tags["H"] = tag
		assert.ErrorIs(t, mod.Validate(), hl7converter.ErrUndefinedOption)
	})

	t.Run("invalid template default rejected", func(t *testing.T) {
		t.Parallel()

		mod := cloneModification(base)
		tag := mod.TagsInfo.Tags["H"]
		tag.Tempalate = "H|??"
		mod.TagsInfo.Tags["H"] = tag
		assert.Error(t, mod.Validate())
	})

	t.Run("invalid link syntax rejected", func(t *testing.T) {
		t.Parallel()

		mod := cloneModification(base)
		tag := mod.TagsInfo.Tags["H"]
		tag.Tempalate = "H|<MSH-2"
		mod.TagsInfo.Tags["H"] = tag
		assert.Error(t, mod.Validate())
	})
}

func TestConverterTempalateParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		output  []int
		wantErr bool
	}{
		{
			name:   "Ok - full line",
			input:  "1.1^<H-2>^MINDRAY",
			output: []int{1, 1, 1, 1, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1},
		},
		{
			name:   "Ok - just link",
			input:  "<H-2>",
			output: []int{0, 0, 0, 0, 0},
		},
		{
			name:    "Error - link without end char",
			input:   "astm^<H-2",
			wantErr: true,
		},
		{
			name:    "Error - link without start char",
			input:   "astm^H-2>",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mask, err := hl7converter.TempalateParse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.output, mask)
		})
	}
}

func TestNewConverterParamsCrossValidation(t *testing.T) {
	t.Parallel()

	t.Run("allows input tags without matching output tags when unused", func(t *testing.T) {
		t.Parallel()

		configPath := writeConfigFixture(t, `{
  "input": {
    "component_separator": "^",
    "component_array_separator": "~",
    "field_separator": "|",
    "line_separator": "\n",
    "tags_info": {
      "positions": {"1": "A", "2": "B"},
      "tags": {
        "A": {"linked": "X", "fields_number": 2, "template": ""},
        "B": {"linked": "Y", "fields_number": 2, "template": ""}
      }
    }
  },
  "output": {
    "component_separator": "^",
    "component_array_separator": "~",
    "field_separator": "|",
    "line_separator": "\r",
    "tags_info": {
      "positions": {"1": "X"},
      "tags": {
        "X": {"linked": "A", "fields_number": 2, "template": "X|<A-2>"}
      }
    }
  }
}`)

		params, err := hl7converter.NewConverterParams(configPath, "input", "output")
		require.NoError(t, err)
		require.NotNil(t, params)
	})

	t.Run("rejects unknown output linked input tag", func(t *testing.T) {
		t.Parallel()

		configPath := writeConfigFixture(t, `{
  "input": {
    "component_separator": "^",
    "component_array_separator": "~",
    "field_separator": "|",
    "line_separator": "\n",
    "tags_info": {
      "positions": {"1": "H"},
      "tags": {
        "H": {"linked": "OUT", "fields_number": 3, "template": "H|A|B"}
      }
    }
  },
  "output": {
    "component_separator": "^",
    "component_array_separator": "~",
    "field_separator": "|",
    "line_separator": "\r",
    "tags_info": {
      "positions": {"1": "OUT"},
      "tags": {
        "OUT": {"linked": "ZZ", "fields_number": 3, "template": "OUT|<H-2>|CONST"}
      }
    }
  }
}`)

		_, err := hl7converter.NewConverterParams(configPath, "input", "output")
		require.Error(t, err)
		assert.Contains(t, err.Error(), `unknown input tag "ZZ"`)
	})

	t.Run("rejects unknown template tag reference", func(t *testing.T) {
		t.Parallel()

		configPath := writeConfigFixture(t, `{
  "input": {
    "component_separator": "^",
    "component_array_separator": "~",
    "field_separator": "|",
    "line_separator": "\n",
    "tags_info": {
      "positions": {"1": "H"},
      "tags": {
        "H": {"linked": "OUT", "fields_number": 3, "template": "H|A|B"}
      }
    }
  },
  "output": {
    "component_separator": "^",
    "component_array_separator": "~",
    "field_separator": "|",
    "line_separator": "\r",
    "tags_info": {
      "positions": {"1": "OUT"},
      "tags": {
        "OUT": {"linked": "H", "fields_number": 3, "template": "OUT|<ZZ-2>|CONST"}
      }
    }
  }
}`)

		_, err := hl7converter.NewConverterParams(configPath, "input", "output")
		require.Error(t, err)
		assert.Contains(t, err.Error(), `references unknown input tag "ZZ"`)
	})

	t.Run("rejects field position outside input fields_number", func(t *testing.T) {
		t.Parallel()

		configPath := writeConfigFixture(t, `{
  "input": {
    "component_separator": "^",
    "component_array_separator": "~",
    "field_separator": "|",
    "line_separator": "\n",
    "tags_info": {
      "positions": {"1": "H"},
      "tags": {
        "H": {"linked": "OUT", "fields_number": 3, "template": "H|A|B"}
      }
    }
  },
  "output": {
    "component_separator": "^",
    "component_array_separator": "~",
    "field_separator": "|",
    "line_separator": "\r",
    "tags_info": {
      "positions": {"1": "OUT"},
      "tags": {
        "OUT": {"linked": "H", "fields_number": 3, "template": "OUT|<H-4>|CONST"}
      }
    }
  }
}`)

		_, err := hl7converter.NewConverterParams(configPath, "input", "output")
		require.Error(t, err)
		assert.Contains(t, err.Error(), `outside input tag "H" fields_number 3`)
	})
}

func cloneModification(in *hl7converter.Modification) *hl7converter.Modification {
	out := *in
	out.TagsInfo.Positions = make(map[string]string, len(in.TagsInfo.Positions))
	for k, v := range in.TagsInfo.Positions {
		out.TagsInfo.Positions[k] = v
	}

	out.TagsInfo.Tags = make(map[string]hl7converter.Tag, len(in.TagsInfo.Tags))
	for k, v := range in.TagsInfo.Tags {
		tagCopy := v
		tagCopy.Options = append([]string(nil), v.Options...)
		out.TagsInfo.Tags[k] = tagCopy
	}

	out.Types = make(map[string][][]string, len(in.Types))
	for k, v := range in.Types {
		rows := make([][]string, len(v))
		for i := range v {
			rows[i] = append([]string(nil), v[i]...)
		}
		out.Types[k] = rows
	}

	out.Aliases = make(hl7converter.Aliases, len(in.Aliases))
	for k, v := range in.Aliases {
		out.Aliases[k] = v
	}

	return &out
}

func writeConfigFixture(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}
