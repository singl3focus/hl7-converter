package hl7converter_test

import (
	"bytes"
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

	configInputBlockType2 = "cl1200_hl7"
	configOutputBlockType2 = "access_cl1200"
)

var (
	inputMsgHBL = []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
		"P|1||||^||||||||||||||||||||||||||||\n" +
		"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
		"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
		"L|1|N")

	outputMsgHBL = []byte("MSH|^~\\&|Manufacturer|Model|||20220327||ORU~R01||P|2.3.1||||||ASCII|" + CR +
		"PID||142212|||||||||||||||||||||||||||" + CR + "OBR||142212|||||||||||||URI|||||||||||||||||||||||||||" +
		CR + "OBX|||Urina4~screening~tempo-analisi-minuti|tempo-analisi-minuti|180||||||F|||||")

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
	"Q|1|123~134144||||||||||") // some transformations (^ => ~)
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
)


func TestFullConvertMsg(t *testing.T) {
	ready, err := hl7converter.FullConvertMsg(configFilename, configInputBlockType, configOutputBlockType, inputMsgHBL)
	if err != nil {
		t.Fatal(err)
	}

	for _, msg := range ready {
		t.Log("Result (frame):", string(msg))
	}

	res := bytes.Join(ready, []byte(CR))

	if !bytes.Equal(res, outputMsgHBL) {
		t.Fatal(failed, "Ouput msg has been wrong converted")
	}

	t.Log("New checking")

	ready, err = hl7converter.FullConvertMsg(configFilename, configOutputBlockType, configInputBlockType, inputNewMsgHBL)
	if err != nil {
		t.Fatal(err)
	}

	for _, msg := range ready {
		t.Log("Result (frame):", string(msg))
	}

	res = bytes.Join(ready, []byte("\n"))

	if !bytes.Equal(res, outputNewMsgHBL) {
		t.Fatal(failed, "Ouput msg has been wrong converted")
	}

	t.Log("New checking")

	ready, err = hl7converter.FullConvertMsg(configFilename, configInputBlockType2, configOutputBlockType2, inMsgHL7CL1200)
	if err != nil {
		t.Fatal(err)
	}

	for _, msg := range ready {
		t.Log("Result (frame):", string(msg))
	}

	t.Log("New checking")

	ready, err = hl7converter.FullConvertMsg(configFilename, "astm_cl_8000", "access_cl_8000", NEWinpCL8000)
	if err != nil {
		t.Fatal(err)
	}

	for _, msg := range ready {
		t.Log("Result (frame):", string(msg))
	}
	

	t.Log(success, "TestConvertMsg right")
}


func TestFullConvertMsgWithSameTags(t *testing.T) {
	readyMsgs, err := hl7converter.FullConvertMsgWithSameTags(configFilename, configInputBlockType2, configOutputBlockType2, inMsgHL7CL1200Mult, "OBX")
	if err != nil {
		t.Fatal(err)
	}

	for _, msg := range readyMsgs {
		t.Log("Result:", string(msg))
	}


	var finalLine string

	for i, msg := range readyMsgs {
		breakL := false
		if (i + 1) % 4 == 0 {
			finalLine += string(msg) // it's line with same tag and we get it and add to finalLine 
			breakL = true
		} else if i < 3 {
			finalLine += string(msg) // it's service tags (it's duplicate in every msg)
			breakL = true
		} else {
			breakL = false
		}

		if i != (len(readyMsgs) - 1) && breakL{
			finalLine += "\n"
		}
	}

	t.Log("Final result: ", finalLine)

	t.Logf("%s TestConvertMsg right", success)
}


func BenchmarkConvertMsg(b *testing.B) {
	ready, err := hl7converter.FullConvertMsg(configFilename, configInputBlockType, configOutputBlockType, inputMsgHBL)
	if err != nil {
		b.Fatal(err)
	}

	for _, msg := range ready {
		b.Log("Result (frame):", string(msg))
	}

	res := bytes.Join(ready, []byte(CR))

	if !bytes.Equal(res, outputMsgHBL) {
		b.Fatal(failed, "Ouput msg has been wrong converted")
	}

	b.Logf("%s BenchmarkConvertMsg right", success)
}

