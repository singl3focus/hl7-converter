package hl7converter_test

import (
	"testing"
	"path/filepath"

	hl7converter "github.com/singl3focus/hl7-converter"
)

func BenchmarkConvertWithConverter(b *testing.B) {
	var (
		inputMsgHBL = []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
			"P|1||||^||||||||||||||||||||||||||||\n" +
			"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
			"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
			"R|2|^^^Urina4^screening^^tempo-analisi-minuti|90|||||F|||||\n" +
			"L|1|N")
	
		configPath = filepath.Join(workDir, hl7converter.CfgJSON)
		
		configInputBlockType = "hl7_astm_hbl"
		configOutputBlockType = "hl7_mindray_hbl"
	)

	convParams, err := hl7converter.NewConverterParams(configPath, configInputBlockType, configOutputBlockType)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {

		_, _, err := hl7converter.Convert(convParams, inputMsgHBL)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.Log(success, "TestConvertMsg right")
}

func BenchmarkReadYamlConfig(b *testing.B) {
	var (
		configPath = filepath.Join(workDir, hl7converter.CfgYaml)

		testModification = "hl7_astm_hbl"
	)

	for i := 0; i < b.N; i++ {
		inputModification1, err := hl7converter.ReadYAMLConfigBlock(configPath, testModification)
		if err != nil || inputModification1 == nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadJSONConfig(b *testing.B) {
	var (
		configPath = filepath.Join(workDir, hl7converter.CfgJSON)

		testModification = "hl7_astm_hbl"
	)

	for i := 0; i < b.N; i++ {
		inputModification1, err := hl7converter.ReadJSONConfigBlock(configPath, testModification)
		if err != nil || inputModification1 == nil {
			b.Fatal(err)
		}
	}
}