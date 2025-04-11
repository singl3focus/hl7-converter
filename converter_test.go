package hl7converter_test

import (
	"fmt"
	"path/filepath"
	"testing"

	hl7converter "github.com/singl3focus/hl7-converter/v2"
	"github.com/stretchr/testify/assert"
)

const CR = "\r"

func TestConverterParseInput(t *testing.T) {
	var (
		configPath            = filepath.Join(workDir, hl7converter.CfgJSON)
		configInputBlockType  = "astm_hbl"
		configOutputBlockType = "mindray_hbl"
	)

	convParams, err := hl7converter.NewConverterParams(configPath, configInputBlockType, configOutputBlockType)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	var msg = []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
		"P|1||||^||||||||||||||||||||||||||||\n" +
		"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
		"C|||||||||||||||\n" +
		"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
		"C|||||||||||||||\n" +
		"C|||||||||||||||\n" +
		"C|||||||||||||||\n" +
		"R|2|^^^Urina4^screening^^tempo-analisi-minuti|90|||||F|||||\n" +
		"L|1|N")

	c, err := hl7converter.NewConverter(convParams, hl7converter.WithUsingPositions())
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	result, err := c.ParseInput(msg)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	t.Run("aliases_usage", func(t *testing.T) {
		err = result.ApplyAliases(convParams.InMod.Aliases)
		if err != nil {
			t.Fatalf("%s", err.Error())
		}

		t.Log(result.Aliases())
	})
}

func TestConvertRow(t *testing.T) {
	// t.Parallel() // TODO: uncommit after pointerIndx will be internal field of Converter

	var (
		configPath            = filepath.Join(workDir, hl7converter.CfgJSON)
		configInputBlockType  = "astm_hbl"
		configOutputBlockType = "mindray_hbl"
	)

	convParams, err := hl7converter.NewConverterParams(configPath, configInputBlockType, configOutputBlockType)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	c, err := hl7converter.NewConverter(convParams)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	tests := []struct {
		name   string
		input  []byte
		output string
		err    error
	}{
		{
			name:   "ok_header",
			input:  []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327"),
			output: "MSH|^\\&|Manufacturer|Model|||20220327||ORU^R01||P|2.3.1||||||ASCII|",
		},
		{
			name:   "ok_order",
			input:  []byte("O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||"),
			output: "OBR||142212|||||||||||||URI|||||||||||||||||||||||||||",
		},
		{
			name:   "err_linked_tag_not_found",
			input:  []byte("P|1||||^||||||||||||||||||||||||||||"),
			output: "<nil>",                           // TODO: change <nil> to "" with ',ok' notation
			err:    hl7converter.ErrUndefinedInputTag, // TODO: edit error name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ready, err := c.Convert(tt.input)

			if tt.err != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.output, ready.String())
		})
	}
}

func TestConvertMultiRowsWithManipulations(t *testing.T) {
	// t.Parallel() // TODO: uncommit after pointerIndx will be internal field of Converter

	var (
		configPath            = filepath.Join(workDir, hl7converter.CfgJSON)
		configInputBlockType  = "astm_hbl"
		configOutputBlockType = "mindray_hbl"

		inputMsgHBL = []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
			"P|1||||^||||||||||||||||||||||||||||\n" +
			"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
			"C|||||||||||||||\n" +
			"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
			"C|||||||||||||||\n" +
			"C|||||||||||||||\n" +
			"C|||||||||||||||\n" +
			"R|2|^^^Urina4^screening^^tempo-analisi-minuti|90|||||F|||||\n" +
			"L|1|N")

		outputMsgHBL = []byte("MSH|^\\&|Manufacturer|Model|||20220327||ORU^R01||P|2.3.1||||||ASCII|" + CR +
			"PID||142212|||||||||||||||||||||||||||" + CR +
			"OBR||142212|||||||||||||URI|||||||||||||||||||||||||||" + CR +
			"OBX|||Urina4^screening^tempo-analisi-minuti|tempo-analisi-minuti|180||||||F|||||" + CR +
			"OBX|||Urina4^screening^tempo-analisi-minuti|tempo-analisi-minuti|90||||||F|||||")

		outputMsgTypeHBL = "Results"
	)

	convParams, err := hl7converter.NewConverterParams(configPath, configInputBlockType, configOutputBlockType)
	if err != nil {
		t.Fatal(err.Error())
	}

	c, err := hl7converter.NewConverter(convParams, hl7converter.WithUsingPositions())
	if err != nil {
		t.Fatal(err.Error())
	}

	result, err := c.Convert(inputMsgHBL)
	if err != nil {
		t.Fatal(err.Error())
	}

	assert.Equal(t, outputMsgHBL, result.Bytes())

	t.Run("indentify_msg", func(t *testing.T) {
		msgType, err := hl7converter.IndetifyMsg(convParams, inputMsgHBL)
		if err != nil {
			t.Fatal(err.Error())
		}

		assert.Equal(t, outputMsgTypeHBL, msgType)
	})

	t.Run("script_usage", func(t *testing.T) {
		var (
			oldField = "MSH"
			newField = "NEW-MSH"

			script = `msg.Rows[0].Fields[0].ChangeValue("%s");`
		)

		assert.Equal(t, result.Rows[0].Fields[0].Value, oldField)

		err = result.UseScript(hl7converter.KeyScript, fmt.Sprintf(script, newField))
		if err != nil {
			t.Fatal(err.Error())
		}

		assert.Equal(t, result.Rows[0].Fields[0].Value, newField)

		err = result.UseScript(hl7converter.KeyScript, fmt.Sprintf(script, oldField))
		if err != nil {
			t.Fatal(err.Error())
		}

		assert.Equal(t, result.Rows[0].Fields[0].Value, oldField)
	})

	t.Run("aliases", func(t *testing.T) {
		var (
			aliases = hl7converter.Aliases{
				"Header":    "MSH-9.2",
				"PatientID": "PID-3",
				"Key":       "OBR-16",
			}
		)

		err := result.ApplyAliases(aliases)
		if err != nil {
			t.Fatal(err.Error())
		}

		al, ok := result.Aliases()
		if !ok {
			t.Fatal("aliases is empty")
		}

		assert.Equal(t, al["Header"], "R01")
		assert.Equal(t, al["PatientID"], "142212")
		assert.Equal(t, al["Key"], "URI")
	})
}

func FuzzConvert(f *testing.F) {
	var (
		configPath            = filepath.Join(workDir, hl7converter.CfgJSON)
		configInputBlockType  = "astm_hbl"
		configOutputBlockType = "mindray_hbl"
	)

	convParams, err := hl7converter.NewConverterParams(configPath, configInputBlockType, configOutputBlockType)
	if err != nil {
		f.Fatal(err.Error())
	}

	c, err := hl7converter.NewConverter(convParams)
	if err != nil {
		f.Fatal(err.Error())
	}

	f.Add([]byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327"))

	f.Fuzz(func(t *testing.T, input []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %+v", r)
			}
		}()

		_, _ = c.Convert(input)
	})
}

// TODO: RACE CONDITION TEST FOR CONVERTER (pointerIndx)

// TODO: [ADD TEST FOR EVERY FUNCTION OF CONVERTING]
/*
func TestNotLinkedTag(t *testing.T) {}
func TestTagOptions(t *testing.T) {}
*/

func TestConverterParseMsg(t *testing.T) {
	var (
		configPath            = filepath.Join(workDir, hl7converter.CfgJSON)
		configInputBlockType  = "mindray_hbl"
		configOutputBlockType = "astm_hbl"
	)

	tests := []struct {
		name    string
		input   []byte
		output  map[hl7converter.TagName]hl7converter.SliceFields
		wantErr bool
	}{
		{
			name:  "Ok - single row",
			input: []byte("MSH|^\\&|Manufacturer|Model|||20220327||ORU^R01||P|2.3.1||||||ASCII|"),
			output: map[hl7converter.TagName]hl7converter.SliceFields{
				"MSH": {
					[]string{"^\\&", "Manufacturer", "Model", "", "", "20220327", "", "ORU^R01", "", "P", "2.3.1", "", "", "", "", "", "ASCII", ""},
				},
			},
		},
	}

	convParams, err := hl7converter.NewConverterParams(configPath, configInputBlockType, configOutputBlockType)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	c, err := hl7converter.NewConverter(convParams, hl7converter.WithUsingPositions())
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := c.ParseMsg(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.output, res)
		})
	}
}
