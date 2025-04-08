package hl7converter_test

import (
	"path/filepath"
	"reflect"
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
		t.Fatalf("------%s------", err.Error())
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

	t.Run("converter parse input", func(t *testing.T) {
		t.Parallel()

		c, err := hl7converter.NewConverter(convParams, hl7converter.WithUsingPositions())
		if err != nil {
			t.Fatalf("------%s------", err.Error())
		}

		result, err := c.ParseInput(msg)
		if err != nil {
			t.Fatalf("------%s------", err.Error())
		}

		err = result.UseScript(hl7converter.KeyScript, `msg.Rows[0].Fields[1].ChangeValue("MSGTEST");`)
		if err != nil {
			t.Fatalf("------%s------", err.Error())
		}

		err = result.ApplyAliases(convParams.InMod.Aliases)
		if err != nil {
			t.Fatalf("------%s------", err.Error())
		}

		t.Log(result.String())

		t.Log(result.Aliases())
	})
}

func TestConvertRow(t *testing.T) {
	var (
		configPath            = filepath.Join(workDir, hl7converter.CfgJSON)
		configInputBlockType  = "astm_hbl_single"
		configOutputBlockType = "mindray_hbl_single"
	)

	convParams, err := hl7converter.NewConverterParams(configPath, configInputBlockType, configOutputBlockType)
	if err != nil {
		t.Fatalf("------%s------", err.Error())
	}

	c, err := hl7converter.NewConverter(convParams)
	if err != nil {
		t.Fatalf("------%s------", err.Error())
	}

	tests := []struct {
		name    string
		input   []byte
		output  []byte
		wantErr bool
	}{
		{
			name:   "Ok",
			input:  []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327"),
			output: []byte("MSH|^\\&|Manufacturer|Model|||20220327||ORU^R01||P|2.3.1||||||ASCII|"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ready, err := c.Convert(tt.input)
			if err != nil {
				t.Fatalf("------%s------", err.Error())
			}

			res := ready.String()

			if !(res == string(tt.output)) {
				t.Fatal("------converted msg is wrong------", "expected", string(tt.output), "received", string(res))
			}

			t.Logf("------Success %s------", tt.name)
		})
	}
}

func TestConvertMultiRows(t *testing.T) {
	var (
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

		configPath            = filepath.Join(workDir, hl7converter.CfgJSON)
		configInputBlockType  = "astm_hbl"
		configOutputBlockType = "mindray_hbl"
	)

	convParams, err := hl7converter.NewConverterParams(configPath, configInputBlockType, configOutputBlockType)
	if err != nil {
		t.Fatalf("------%s------", err.Error())
	}

	c, err := hl7converter.NewConverter(convParams, hl7converter.WithUsingPositions())
	if err != nil {
		t.Fatalf("------%s------", err.Error())
	}

	t.Run("convert multi rows", func(t *testing.T) {
		t.Parallel()

		msgType, err := hl7converter.IndetifyMsg(convParams, inputMsgHBL)
		if err != nil {
			t.Fatalf("------%s------", err.Error())
		}

		ready, err := c.Convert(inputMsgHBL)
		if err != nil {
			t.Fatalf("------%s------", err.Error())
		}

		res := ready.String()

		if msgType != outputMsgTypeHBL {
			t.Fatal("------message type is wrong------")
		}

		if !(res == string(outputMsgHBL)) {
			t.Fatal("------converted msg is wrong------ \n", res)
		}
	})
}

// todo: [ADD TEST FOR EVERY FUNCTION OF CONVERTING]
/*
func TestNotLinkedTag(t *testing.T) {}
func TestTagOptions(t *testing.T) {}
*/

/*
var (
	inputMsgHBL = []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
		"P|1||||^||||||||||||||||||||||||||||\n" +
		"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
		"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
		"R|2|^^^Urina4^screening^^tempo-analisi-minuti|90|||||F|||||\n" +
		"L|1|N")

	outputMsgHBL = []byte("MSH|^\\&|Manufacturer|Model|||20220327||ORU^R01||P|2.3.1||||||ASCII|" + CR +
		"PID||142212|||||||||||||||||||||||||||" + CR +
		"OBR||142212|||||||||||||URI|||||||||||||||||||||||||||" + CR +
		"OBX|||Urina4^screening^tempo-analisi-minuti|tempo-analisi-minuti|180||||||F|||||" + CR +
		"OBX|||Urina4^screening^tempo-analisi-minuti|tempo-analisi-minuti|90||||||F|||||")

	outputMsgTypeHBL = "Results"
)

var (
	inputNewMsgHBL = []byte("MSH|~\\&|Manufacturer|Model|||20220327||ORU~R01||P|2.3.1||||||ASCII|" + CR +
		"PID||142212|||||||||||||||||||||||||||" + CR + "OBR||142212|||||||||||||URI|||||||||||||||||||||||||||" +
		CR + "OBX|||Urina4~screening~tempo-analisi-minuti|tempo-analisi-minuti|180||||||F|||||")

	outputNewMsgHBL = []byte("H|\\^&||||||||||P||20220327\n" +
		"P||||||||||||||||||||||||||||||\n" +
		"O||142212|||||||||||||URI^^|||||||||||||||\n" +
		"R||^^^Urina4~screening~tempo-analisi-minuti^Urina4~screening~tempo-analisi-minuti^^Urina4~screening~tempo-analisi-minuti|180|||||F|||||")

	inMsgHL7CL1200 = []byte("MSH|^~\\&|||||20120508150259||QRY^Q02|7|P|2.3.1||||||ASCII|||\n" +
		"PID|1|1001|||Mike||19851001095133|M|||keshi|||||||||||||||beizhu|||||\n" +
		"OBR|1|12345678|10|^|Y|20120405193926|20120405193914|20120405193914|||||linchuangzhenduan|20120405193914|serum|lincyisheng|keshi||||||||3|||||||||||||||||||||||\n" +
		"OBX|1|NM|2|TBil|100| umol/L |-|N|||F||100|20120405194245||yishen|0|")

	NEWinpCL8000 = []byte("H|\\~&|||eCL8000~00.00.03~I05A16100023|||||||RQ|1394-97|20191105190721" + CR +
		"Q|1|123~134144||||||||||")
)

var (
	inMsgHL7CL1200Mult = [][]byte{[]byte("MSH|^~\\&|||||20120508150259||QRY^Q02|7|P|2.3.1||||||ASCII|||\n" +
		"PID|1|1001|||Mike||19851001095133|M|||keshi|||||||||||||||beizhu|||||\n" +
		"OBR|1|12345678|10|^|Y|20120405193926|20120405193914|20120405193914|||||linchuangzhenduan|20120405193914|serum|lincyisheng|keshi||||||||3|||||||||||||||||||||||\n" +
		"OBX|1|NM|2|TBil|100| umol/L |-|N|||F||100|20120405194245||yishen|0|"),

		[]byte("MSH|^~\\&|||||20120508150259||QRY^Q02|7|P|2.3.1||||||ASCII|||\n" +
			"PID|1|1001|||Mike||19851001095133|M|||keshi|||||||||||||||beizhu|||||\n" +
			"OBR|1|12345678|10|^|Y|20120405193926|20120405193914|20120405193914|||||linchuangzhenduan|20120405193914|serum|lincyisheng|keshi||||||||3|||||||||||||||||||||||\n" +
			"OBX|2|NM|5|ALT|98.2| umol/L |-|N|||F||98.2|20120405194403||yishen|0|"),

		[]byte("MSH|^~\\&|||||20120508150259||QRY^Q02|7|P|2.3.1||||||ASCII|||\n" +
			"PID|1|1001|||Mike||19851001095133|M|||keshi|||||||||||||||beizhu|||||\n" +
			"OBR|1|12345678|10|^|Y|20120405193926|20120405193914|20120405193914|||||linchuangzhenduan|20120405193914|serum|lincyisheng|keshi||||||||3|||||||||||||||||||||||\n" +
			"OBX|3|NM|6|AST|26.4| umol/L |-|N|||F||26.4|||yishen||")}

	inMsgECL8000 = [][]byte{[]byte(
		"H|\\~&|||eCL8000~01.00.02.251693~IA5A00001230|||||||PR|1394-97|20240701043119^P|N0002||||||~0~|||||||||||||||||||||||||||^O||160013~||~~0|R|||||||||20000101000000|0||||||||||F|||||^R|22|0~TSH~F|12.300~~↑|mIU/L|10.04~20.45|||N|||20000101000000|20000101000000|eCL8000~IA5A00001230^L|1|N")}
)
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
			wantErr: false,
		},
	}

	convParams, err := hl7converter.NewConverterParams(configPath, configInputBlockType, configOutputBlockType)
	if err != nil {
		t.Fatalf("------%s------", err.Error())
	}

	c, err := hl7converter.NewConverter(convParams, hl7converter.WithUsingPositions())
	if err != nil {
		t.Fatalf("------%s------", err.Error())
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := c.ParseMsg(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if !reflect.DeepEqual(res, tt.output) {
				t.Fatal("incorrect answer", "current output", res, "wait output", tt.output)
			}

			t.Logf("------Success %s------", tt.name)
		})
	}
}
