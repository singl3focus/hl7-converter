package hl7converter_test

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"testing"

	hl7converter "github.com/singl3focus/hl7-converter"
)

const (
	success = "\u2713"
	failed  = "\u2717"

	CR = "^"
)

var (
	configFilename        = "config.json"
	configInputBlockType  = "hl7_astm_hbl"
	configOutputBlockType = "hl7_mindray_hbl"
)

var inputMsgHBL = []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
	"P|1||||^||||||||||||||||||||||||||||\n" +
	"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
	"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
	"L|1|N")


// NOTE: Sequence number skipped
//
// ERROR: if specifie some other tag in input linked tags, but ouput tag specified to other tag an error occurs
// Because - HARDCODE and when we set result, WE depend on the input tag as well
var outputMsgHBL = []byte("MSH|^~\\&|Manufacturer|Model|||20220327||ORU^R01||P|2.3.1||||||ASCII|" + CR +
	"OBR||142212|||||||||||||URI|||||||||||||||||||||||||||" + CR + "OBX|||Urina4~screening~tempo-analisi-minuti|tempo-analisi-minuti|180||||||F|||||")


var (
	inWithFloatPos = ""
	outWithFloatPos = ""

	inWithSomeLinked = ""
	outWithSomeLinked = ""

	inWithSomeLinkedAndFloatPos = ""
	outWithSomeLinkedAndFloatPos = ""
)


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

	log.Println("Result:", string(out))

	if !bytes.Equal(out, outputMsgHBL) {
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

	if !bytes.Equal(out, outputMsgHBL) {
		b.Fatal(failed, "Ouput msg has been wrong converted")
	}

	b.Logf("%s BenchmarkConvertMsg right", success)
}

