package hl7converter_test

import (
	"testing"
	"path/filepath"

    hl7converter "github.com/singl3focus/hl7-converter/v2"
)

func Convert(p *hl7converter.ConverterParams, msg []byte) (*hl7converter.Result, error) {
	c, err := hl7converter.NewConverter(p, hl7converter.WithUsingPositions())
	if err != nil {
		return nil, err
	}

	res, err := c.Convert(msg)
	if err != nil {
		return nil, err
	}
	
	return res, nil
}

func BenchmarkConvertWithPositions(b *testing.B) {
	var (
		inputMsgHBL = []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
			"P|1||||^||||||||||||||||||||||||||||\n" +
			"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
			"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
			"R|2|^^^Urina4^screening^^tempo-analisi-minuti|90|||||F|||||\n" +
			"L|1|N")
	
		configPath = filepath.Join(workDir, hl7converter.CfgJSON)
		
		configInputBlockType = "astm_hbl"
		configOutputBlockType = "mindray_hbl"
	)

	convParams, err := hl7converter.NewConverterParams(configPath, configInputBlockType, configOutputBlockType)
	if err != nil {
		b.Fatal(err)
	}

	// c, err := hl7converter.NewConverter(convParams, hl7converter.WithUsingPositions())
	// if err != nil {
	// 	b.Fatalf("------%s------", err.Error())
	// }

	for i := 0; i < b.N; i++ {
		_, err := Convert(convParams, inputMsgHBL)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConvertWithoutPositions(b *testing.B) {
	var (
		inputMsgHBL = []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
			"P|1||||^||||||||||||||||||||||||||||\n" +
			"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
			"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
			"R|2|^^^Urina4^screening^^tempo-analisi-minuti|90|||||F|||||")
	
		configPath = filepath.Join(workDir, hl7converter.CfgJSON)
		
		configInputBlockType = "astm_hbl"
		configOutputBlockType = "mindray_hbl"
	)

	convParams, err := hl7converter.NewConverterParams(configPath, configInputBlockType, configOutputBlockType)
	if err != nil {
		b.Fatal(err)
	}

	c, err := hl7converter.NewConverter(convParams, hl7converter.WithUsingPositions())
	if err != nil {
		b.Fatalf("------%s------", err.Error())
	}

	for i := 0; i < b.N; i++ {
		_, err := c.Convert(inputMsgHBL)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadJSONConfig(b *testing.B) {
	var (
		configPath = filepath.Join(workDir, hl7converter.CfgJSON)

		testModification = "astm_hbl"
	)

	for i := 0; i < b.N; i++ {
		inputModification1, err := hl7converter.ReadJSONConfigBlock(configPath, testModification)
		if err != nil || inputModification1 == nil {
			b.Fatal(err)
		}
	}
}