package hl7converter_test

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	hl7converter "github.com/singl3focus/hl7-converter/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const CR = "\r"

var (
	sampleInputWithComments = []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
		"P|1||||^||||||||||||||||||||||||||||\n" +
		"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
		"C|||||||||||||||\n" +
		"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
		"C|||||||||||||||\n" +
		"C|||||||||||||||\n" +
		"C|||||||||||||||\n" +
		"R|2|^^^Urina4^screening^^tempo-analisi-minuti|90|||||F|||||\n" +
		"L|1|N")

	sampleInputDrivenInput = []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
		"P|1||||^||||||||||||||||||||||||||||\n" +
		"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
		"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
		"R|2|^^^Urina4^screening^^tempo-analisi-minuti|90|||||F|||||")

	expectedConvertedMessage = []byte("MSH|^\\&|Manufacturer|Model|||20220327||ORU^R01||P|2.3.1||||||ASCII|" + CR +
		"PID||142212|||||||||||||||||||||||||||" + CR +
		"OBR||142212|||||||||||||URI|||||||||||||||||||||||||||" + CR +
		"OBX|||Urina4^screening^tempo-analisi-minuti|tempo-analisi-minuti|180||||||F|||||" + CR +
		"OBX|||Urina4^screening^tempo-analisi-minuti|tempo-analisi-minuti|90||||||F|||||")
)

func TestConverterParseInput(t *testing.T) {
	t.Parallel()

	convParams := mustParams(t, "astm_hbl", "mindray_hbl")

	c, err := hl7converter.NewConverter(convParams, hl7converter.WithUsingPositions())
	require.NoError(t, err)

	result, err := c.ParseInput(sampleInputWithComments)
	require.NoError(t, err)

	err = result.ApplyAliases(convParams.InputModification.Aliases)
	require.NoError(t, err)

	aliases, ok := result.Aliases()
	require.True(t, ok)
	assert.Equal(t, "sireAstmCom", aliases["Header"])
	assert.Equal(t, "142212", aliases["Number"])
	assert.Equal(t, "N", aliases["LowerFlag"])
}

func TestConvertRow(t *testing.T) {
	t.Parallel()

	c := mustConverter(t, false)

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
			output: "<nil>",
			err:    hl7converter.ErrUndefinedInputTag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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

func TestConvertInputDrivenMultiRows(t *testing.T) {
	t.Parallel()

	c := mustConverter(t, false)

	result, err := c.Convert(sampleInputDrivenInput)
	require.NoError(t, err)
	assert.Equal(t, expectedConvertedMessage, result.Bytes())
}

func TestConvertMultiRowsWithManipulations(t *testing.T) {
	t.Parallel()

	convParams := mustParams(t, "astm_hbl", "mindray_hbl")
	c, err := hl7converter.NewConverter(convParams, hl7converter.WithUsingPositions())
	require.NoError(t, err)

	t.Run("convert_and_identify", func(t *testing.T) {
		t.Parallel()

		result, err := c.Convert(sampleInputWithComments)
		require.NoError(t, err)
		assert.Equal(t, expectedConvertedMessage, result.Bytes())

		msgType, err := hl7converter.IndetifyMsg(convParams, sampleInputWithComments)
		require.NoError(t, err)
		assert.Equal(t, "Results", msgType)
	})

	t.Run("script_usage", func(t *testing.T) {
		t.Parallel()

		result, err := c.Convert(sampleInputWithComments)
		require.NoError(t, err)

		const (
			oldField = "MSH"
			newField = "NEW-MSH"
		)

		script := `msg.Rows[0].Fields[0].ChangeValue("%s");`

		assert.Equal(t, oldField, result.Rows[0].Fields[0].Value)
		require.NoError(t, result.UseScript(hl7converter.KeyScript, fmt.Sprintf(script, newField)))
		assert.Equal(t, newField, result.Rows[0].Fields[0].Value)
		require.NoError(t, result.UseScript(hl7converter.KeyScript, fmt.Sprintf(script, oldField)))
		assert.Equal(t, oldField, result.Rows[0].Fields[0].Value)
	})

	t.Run("aliases", func(t *testing.T) {
		t.Parallel()

		result, err := c.Convert(sampleInputWithComments)
		require.NoError(t, err)

		aliases := hl7converter.Aliases{
			"Header":    "MSH-9.2",
			"PatientID": "PID-3",
			"Key":       "OBR-16",
		}

		require.NoError(t, result.ApplyAliases(aliases))

		al, ok := result.Aliases()
		require.True(t, ok)
		assert.Equal(t, "R01", al["Header"])
		assert.Equal(t, "142212", al["PatientID"])
		assert.Equal(t, "URI", al["Key"])
	})
}

func TestConvertConcurrentSharedConverter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		withPositions  bool
		input          []byte
		expectedOutput []byte
	}{
		{
			name:           "input_driven",
			withPositions:  false,
			input:          sampleInputDrivenInput,
			expectedOutput: expectedConvertedMessage,
		},
		{
			name:           "position_driven",
			withPositions:  true,
			input:          sampleInputWithComments,
			expectedOutput: expectedConvertedMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := mustConverter(t, tt.withPositions)

			var wg sync.WaitGroup
			errCh := make(chan error, 32)

			for i := 0; i < 32; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					result, err := c.Convert(tt.input)
					if err != nil {
						errCh <- err
						return
					}

					if string(result.Bytes()) != string(tt.expectedOutput) {
						errCh <- fmt.Errorf("unexpected output: %q", result.String())
					}
				}()
			}

			wg.Wait()
			close(errCh)

			for err := range errCh {
				require.NoError(t, err)
			}
		})
	}
}

func TestConvertWithUsingAliasesCanReuseSameConverter(t *testing.T) {
	t.Parallel()

	params := mustParams(t, "astm_hbl", "mindray_hbl")
	c, err := hl7converter.NewConverter(params, hl7converter.WithUsingPositions(), hl7converter.WithUsingAliases())
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		result, err := c.Convert(sampleInputDrivenInput)
		require.NoError(t, err)

		aliases, ok := result.Aliases()
		require.True(t, ok)
		assert.Equal(t, "R01", aliases["Header"])
		assert.Equal(t, "142212", aliases["PatientID"])
		assert.Equal(t, "URI", aliases["Key"])
	}
}

func FuzzConvert(f *testing.F) {
	convParams, err := hl7converter.NewConverterParams(filepath.Join(workDir, "examples", testConfigJSON), "astm_hbl", "mindray_hbl")
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

func TestConverterParseMsg(t *testing.T) {
	t.Parallel()

	convParams := mustParams(t, "mindray_hbl", "astm_hbl")
	c, err := hl7converter.NewConverter(convParams, hl7converter.WithUsingPositions())
	require.NoError(t, err)

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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

func mustParams(t *testing.T, inputBlock, outputBlock string) *hl7converter.ConverterParams {
	t.Helper()

	configPath := filepath.Join(workDir, "examples", testConfigJSON)
	params, err := hl7converter.NewConverterParams(configPath, inputBlock, outputBlock)
	require.NoError(t, err)

	return params
}

func mustConverter(t *testing.T, withPositions bool) *hl7converter.Converter {
	t.Helper()

	params := mustParams(t, "astm_hbl", "mindray_hbl")

	opts := make([]hl7converter.OptionFunc, 0, 1)
	if withPositions {
		opts = append(opts, hl7converter.WithUsingPositions())
	}

	c, err := hl7converter.NewConverter(params, opts...)
	require.NoError(t, err)

	return c
}
