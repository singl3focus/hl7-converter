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
	
		configInputBlockType = "hl7_astm_hbl"
		configOutputBlockType = "hl7_mindray_hbl"
	)

	configPath := filepath.Join(workDir, configFilenameYaml)
	schemaPath := filepath.Join(workDir, configFilenameJSONSchema)
	for i := 0; i < b.N; i++ {
		_, _, err := hl7converter.ConvertWithConverter(
			schemaPath, configPath, configInputBlockType, configOutputBlockType, inputMsgHBL, yaml)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.Log(success, "TestConvertMsg right")
}

func BenchmarkReadYamlConfig(b *testing.B) {
	configPath := filepath.Join(workDir, configFilenameYaml)

	for i := 0; i < b.N; i++ {
		inputModification1, err := hl7converter.ReadYAMLConfigBlock(configPath, "hl7_astm_hbl")
		if err != nil || inputModification1 == nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadJSONConfig(b *testing.B) {
	configPath := filepath.Join(workDir, configFilenameJSON)
	schemaPath := filepath.Join(workDir, configFilenameJSONSchema)

	testModification := "hl7_astm_hbl"

	for i := 0; i < b.N; i++ {
		inputModification1, err := hl7converter.ReadJSONConfigBlock(schemaPath, configPath, testModification)
		if err != nil || inputModification1 == nil {
			b.Fatal(err)
		}
	}
}