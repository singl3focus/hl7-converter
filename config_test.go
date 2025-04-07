package hl7converter_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	hl7converter "github.com/singl3focus/hl7-converter/v2"
	"github.com/stretchr/testify/assert"
)

var workDir string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}

	workDir = wd
}

func TestReadJSONConfigBlock(t *testing.T) {
	var (
		configPath = filepath.Join(workDir, hl7converter.CfgJSON)

		cfgInBlockName = "astm_hbl"
	)

	m, err := hl7converter.ReadJSONConfigBlock(configPath, cfgInBlockName)
	if err != nil || m == nil {
		t.Fatal(err)
	}

	t.Log("------Success TestReadJSONConfigBlock------")
}

func TestConverterTempalateParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  []int
		wantErr bool
	}{
		{
			name:    "Ok - full line",
			input:   "1.1^<H-2>^MINDRAY",
			output:  []int{1, 1, 1, 1, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1},
			wantErr: false,
		},
		{
			name:    "Ok - just link",
			input:   "<H-2>",
			output:  []int{0, 0, 0, 0, 0},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mask, err := hl7converter.TempalateParse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if slices.Compare(mask, tt.output) != 0 {
				t.Fatal("incorrect answer", "current output", mask, "wait output", tt.output)
			}

			t.Logf("------Success %s------", tt.name)
		})
	}
}