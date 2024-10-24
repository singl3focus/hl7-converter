package hl7converter_test

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	hl7converter "github.com/singl3focus/hl7-converter/v2"
	"github.com/stretchr/testify/assert"
)

const (
	success = "\u2713"
	failed  = "\u2717"

	CR = "\r"
)

var workDir string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}

	workDir = wd
}

func TestNewField(t *testing.T) {
	var (
		componentSep    = "^"
		componentArrSep = "/"
	)

	tests := []struct {
		name       string
		fieldValue string
		result     *hl7converter.Field
	}{
		{
			name:       "Ok",
			fieldValue: "sireAstmCom^1^P/LIS02^20241021",
			result: &hl7converter.Field{
				Value:      "sireAstmCom^1^P/LIS02^20241021",
				Components: []string{"sireAstmCom","1","P","LIS02","20241021"},
				Array:      []*hl7converter.Field{
					{
						Value: "sireAstmCom^1^P",
						Components: []string{"sireAstmCom","1","P"},
					},
					{
						Value: "LIS02^20241021",
						Components: []string{"LIS02","20241021"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := hl7converter.NewField(tt.fieldValue, componentSep, componentArrSep)

			if !reflect.DeepEqual(res, tt.result) {
				t.Fatal("incorrect answer", "current output", res, "wait output", tt.result)
			}

			t.Logf("%s ------Success %s------", success, tt.name)
		})
	}
}


func TestConvertWithConverterRow(t *testing.T) {
	var (
		inputMsgHBL  = []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327")
		outputMsgHBL = []byte("MSH|^\\&|Manufacturer|Model|||20220327||ORU^R01||P|2.3.1||||||ASCII|")

		configPath            = filepath.Join(workDir, hl7converter.CfgJSON)
		configInputBlockType  = "astm_hbl_single"
		configOutputBlockType = "mindray_hbl_single"
	)

	convParams, err := hl7converter.NewConverterParams(configPath, configInputBlockType, configOutputBlockType)
	if err != nil {
		t.Fatalf("------%s------", err.Error())
	}

	ready, _, err := hl7converter.Convert(convParams, inputMsgHBL, false)
	if err != nil {
		t.Fatalf("------%s------", err.Error())
	}

	res := ready.AssembleMessage()

	if !(res == string(outputMsgHBL)) {
		t.Fatal("------converted msg is wrong------", "wait", string(outputMsgHBL), "current", string(res))
	}

	t.Log(success, "TestConvertMsg right")
}

func TestConvertWithConverterMultiRows(t *testing.T) {
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

	msgType, err := hl7converter.IndetifyMsg(*convParams, inputMsgHBL)
	if err != nil {
		t.Fatalf("------%s------", err.Error())
	}

	ready, _, err := hl7converter.Convert(convParams, inputMsgHBL, true)
	if err != nil {
		t.Fatalf("------%s------", err.Error())
	}

	res := ready.AssembleMessage()

	t.Log("message type:", msgType)
	if msgType != outputMsgTypeHBL {
		t.Fatal("------message type is wrong------")
	}

	if !(res == string(outputMsgHBL)) {
		t.Fatal("------converted msg is wrong------ \n", res)
	}

	t.Log(success, "TestConvertMsg right")
}

// [ADDED TEST FOR EVERY FUNCTION OF CONVERTING]
/*
func TestNotLinkedTag(t *testing.T) {}

func TestTagOptions(t *testing.T) {}
*/

// [ADDED TEST FOR EVERY FUNCTION OF CONVERTING]
/*
func TestNotLinkedTag(t *testing.T) {}

func TestTagOptions(t *testing.T) {}
*/

func TestReadJSONConfigBlock(t *testing.T) {
	var (
		configPath = filepath.Join(workDir, hl7converter.CfgJSON)

		cfgInBlockName = "astm_hbl"
	)

	Modification, err := hl7converter.ReadJSONConfigBlock(configPath, cfgInBlockName)
	if err != nil || Modification == nil {
		t.Fatal(err)
	}

	t.Log(success, "------Success reading modification by JSON------")
}

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
			name:    "ok - just link",
			input:   "<H-2>",
			output:  []int{0, 0, 0, 0, 0},
			wantErr: false,
		},
	}

	conv := hl7converter.Converter{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mask, err := conv.TempalateParse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if slices.Compare(mask, tt.output) != 0 {
				t.Fatal("incorrect answer", "current output", mask, "wait output", tt.output)
			}

			t.Logf("%s ------Success %s------", success, tt.name)
		})
	}
}

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
					[]string{"^\\&", "Manufacturer", "Model", "", "", "20220327", "", "ORU^R01", "", "P", "2.3.1", "", "", "", "", "", "ASCII", ""}},
			},
			wantErr: false,
		},
	}

	convParams, err := hl7converter.NewConverterParams(configPath, configInputBlockType, configOutputBlockType)
	if err != nil {
		t.Fatalf("------%s------", err.Error())
	}
	converter, err := hl7converter.NewConverter(convParams.InMod, convParams.OutMod)
	if err != nil {
		t.Fatalf("------%s------", err.Error())
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := converter.ParseMsg(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if !reflect.DeepEqual(res, tt.output) {
				t.Fatal("incorrect answer", "current output", res, "wait output", tt.output)
			}

			t.Logf("%s ------Success %s------", success, tt.name)
		})
	}
}
