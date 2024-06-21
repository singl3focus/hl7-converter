package hl7converter_test

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	hl7converter "github.com/singl3focus/hl7-converter"
)

const (
	success = "\u2713"
	failed  = "\u2717"
)

var (
	CR = "\x0D"

	configFilename        = "config.json"
	configInputBlockType  = "hl7_astm_hbl"
	configOutputBlockType = "hl7_mindray_hbl"

	configInputBlockType2 = "cl1200_astm"
	configOutputBlockType2 = "access"
)

var inputMsgHBL = []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
	"P|1||||^||||||||||||||||||||||||||||\n" +
	"O|1|1^???. 1 - ???. 1||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
	"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
	"L|1|N")

// Sequence number skipped
var outputMsgHBL = []byte("MSH|^~\\&|Manufacturer|Model|||20220327||ORU^R01||P|2.3.1||||||ASCII|" + CR +
	"PID||1|||" + CR + "OBR||1|||||||||||||URI" + CR + "OBX|1||Urina4~screening~tempo-analisi-minuti|tempo-analisi-minuti|180||||||F")

var inputMsgCL1200 = []byte("H|\\^&|||LIS|||||Access||P|1\n" +
	"P|1|1345861956\n" +
	"O|1|1345861956||^^^TS\\^^^ATTG|R||||||A||||Serum\n" +
	"L|1|N")

// Sequence number skipped
var outputMsgCL1200 = []byte("H|\\^&||||||||||SA|1394-97|20240621130329\n" + 
	"P|1||||||||||||||||||||||||||||||||||\n" +
	"O|1|1^^|1345861956|1^TS^^\\2^ATTG^^|R|20240621130329|20240621130329||||||||Serum||||||||||Q|||||\n" +
	"L|1|N")


func TestConvertMsg(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(failed, err)
	}
	configPath := filepath.Join(wd, configFilename)

	inputModification, err := hl7converter.ReadJSONConfigBlock(configPath, configInputBlockType)
	if err != nil || inputModification == nil {
		t.Fatal(failed, err)
	}

	outputModification, err := hl7converter.ReadJSONConfigBlock(configPath, configOutputBlockType)
	if err != nil || outputModification == nil {
		t.Fatal(failed, err)
	}

	out, err := hl7converter.ConvertMsg(inputModification, outputModification, inputMsgHBL)
	if err != nil {
		t.Fatal(failed, err)
	}

	if reflect.DeepEqual(out, outputMsgHBL) {
		t.Fatal(failed, "Ouput msg has been wrong converted")
	}

	// _______________________________________________

	inputModification2, err := hl7converter.ReadJSONConfigBlock(configPath, configInputBlockType2)
	if err != nil || inputModification2 == nil {
		t.Fatal(failed, err)
	}
	log.Println(inputModification2)

	outputModification2, err := hl7converter.ReadJSONConfigBlock(configPath, configOutputBlockType2)
	if err != nil || outputModification2 == nil {
		t.Fatal(failed, err)
	}
	log.Println(outputModification2)

	out, err = hl7converter.ConvertMsg(inputModification2, outputModification2, inputMsgCL1200)
	if err != nil {
		t.Fatal(failed, err)
	}

	log.Println(string(out))

	if reflect.DeepEqual(out, outputMsgCL1200) {
		t.Fatal(failed, "Ouput msg has been wrong converted")
	}

	t.Logf("%s TestConvertMsg right", success)
}

func BenchmarkConvertMsg(b *testing.B) {
	wd, err := os.Getwd()
	if err != nil {
		b.Fatal(failed, err)
	}
	configPath := filepath.Join(wd, configFilename)

	inputModification, err := hl7converter.ReadJSONConfigBlock(configPath, configInputBlockType)
	if err != nil || inputModification == nil {
		b.Fatal(failed, err)
	}

	outputModification, err := hl7converter.ReadJSONConfigBlock(configPath, configOutputBlockType)
	if err != nil || outputModification == nil {
		b.Fatal(failed, err)
	}

	out, err := hl7converter.ConvertMsg(inputModification, outputModification, inputMsgHBL)
	if err != nil {
		b.Fatal(failed, err)
	}

	if reflect.DeepEqual(out, outputMsgHBL) {
		b.Fatal(failed, "Ouput msg has been wrong converted")
	}

	b.Logf("%s BenchmarkConvertMsg right", success)
}

