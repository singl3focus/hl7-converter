package hl7converter_test

import (
	"strings"
	"testing"

	hl7converter "github.com/singl3focus/hl7-converter"
)

const (
	success = "\u2713"
	failed  = "\u2717"

	CR = "\r"
)

var (
	configFilename        = "config.json"

	configInputBlockType  = "hl7_astm_hbl"
	configOutputBlockType = "hl7_mindray_hbl"

	configInputBlockType2 = "cl1200_hl7"
	configOutputBlockType2 = "access_cl1200"

	configInputBlockType3 = "astm_cl_8000"
	configOutputBlockType3 = "access_cl_8000"
)

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


func TestConvertWithConverter(t *testing.T) {
	ready, msgType, err := hl7converter.ConvertWithConverter(configFilename, configInputBlockType, configOutputBlockType, inputMsgHBL)
	if err != nil {
		t.Fatal(err)
	}

	res := make([]byte, 0, 1024)
	for i, rowFields := range ready {
		t.Logf("%d row: %v\n", i+1, rowFields)
		
		readyRow := strings.Join(rowFields, "|")
		t.Logf("%d row: %v\n", i+1, readyRow)

		res = append(res, []byte(readyRow)...)

		if i < (len(ready) - 1) {
			res = append(res, []byte(CR)...)
		}
	}

	t.Log("message type:", msgType)
	if msgType != outputMsgTypeHBL {
		t.Fatal("message type  is wrong")
	}
	
	if !(string(res) == string(outputMsgHBL)) {
		t.Fatal("converted msg is wrong")
	}

	t.Logf("[]byte results len %v and cap %v", len(res), cap(res))

	t.Log(success, "TestConvertMsg right")
}


func BenchmarkConvertMsg(b *testing.B) {
	b.Logf("%s BenchmarkConvertMsg right", success)
}

