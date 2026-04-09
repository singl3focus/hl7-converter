package hl7converter_test

import (
	"os"
	"path/filepath"
	"testing"

	hl7converter "github.com/singl3focus/hl7-converter/v2"
	"github.com/stretchr/testify/assert"
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
	var (
		configPath = filepath.Join(workDir, "examples", testConfigJSON)

		cfgInBlockName = "astm_hbl"
	)

	m, err := hl7converter.ReadJSONConfigBlock(configPath, cfgInBlockName)
	if err != nil || m == nil {
		t.Fatal(err)
	}
}

func TestModificationValidate(t *testing.T) {
	mod := &hl7converter.Modification{
		ComponentSeparator:    "^",
		ComponentArrSeparator: " ",
		FieldSeparator:        "|",
		LineSeparator:         "\n",
		TagsInfo: hl7converter.TagsInfo{
			Positions: map[string]string{"1": "H"},
			Tags: map[string]hl7converter.Tag{
				"H": {
					Linked:       "MSH",
					FieldsNumber: 1,
					Tempalate:    "H",
				},
			},
		},
	}

	assert.NoError(t, mod.Validate())

	mod.TagsInfo.Positions["bad"] = "MSH"
	assert.Error(t, mod.Validate())
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
