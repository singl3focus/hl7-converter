package hl7converter_test

import (
	"testing"
	"path/filepath"

    hl7converter "github.com/singl3focus/hl7-converter/v2"
)

func Convert(withPositions bool) error {
	var (
		inputMsgHBL = []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
			"P|1||||^||||||||||||||||||||||||||||\n" +
			"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
			"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
			"R|2|^^^Urina4^screening^^tempo-analisi-minuti|90|||||F|||||\n")
	
		configPath = filepath.Join(workDir, hl7converter.CfgJSON)
		
		configInputBlockType = "astm_hbl"
		configOutputBlockType = "mindray_hbl"
	)

	convParams, err := hl7converter.NewConverterParams(configPath, configInputBlockType, configOutputBlockType)
	if err != nil {
		return err
	}

	var c *hl7converter.Converter
	if withPositions {
		c, err = hl7converter.NewConverter(convParams, hl7converter.WithUsingPositions())
	} else {
		c, err = hl7converter.NewConverter(convParams)
	}
	if err != nil {
		return err
	}

	if _, err = c.Convert(inputMsgHBL); err != nil {
		return err
	}
	
	return nil
}

func BenchmarkConvertWithPositions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		err := Convert(true)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConvertWithoutPositions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		err := Convert(false)
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
		_, err := hl7converter.ReadJSONConfigBlock(configPath, testModification)
		if err != nil {
			b.Fatal(err)
		}
	}
}