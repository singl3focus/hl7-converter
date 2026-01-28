package examples

import (
	"log"
	"os"
	"path/filepath"

	hl7converter "github.com/singl3focus/hl7-converter/v2"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	configPath := filepath.Join(workDir, hl7converter.CfgJSON)

	convParams, err := hl7converter.NewConverterParams(configPath, "astm_hbl", "mindray_hbl")
	if err != nil {
		return err
	}

	conv, err := hl7converter.NewConverter(convParams, hl7converter.WithUsingPositions(), hl7converter.WithUsingAliases())
	if err != nil {
		return err
	}

	inputMsgHBL := sampleInput()

	msgType, err := hl7converter.IndetifyMsg(convParams, inputMsgHBL)
	if err != nil {
		return err
	}

	ready, err := conv.Convert(inputMsgHBL)
	if err != nil {
		return err
	}

	// Simple routing example based on type
	switch msgType {
	case "type_1":
		_ = append(ready.Bytes(), byte('\r'))
	case "type_2":
		_, _ = ready.Aliases()
	case "type_3":
		script := `msg.Rows[0].Fields[1].ChangeValue("TYPE");`
		if err := ready.UseScript(hl7converter.KeyScript, script); err != nil {
			return err
		}
	}

	log.Println(ready.String())

	return nil
}

func sampleInput() []byte {
	return []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
		"P|1||||^||||||||||||||||||||||||||||\n" +
		"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
		"C|||||||||||||||\n" +
		"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
		"C|||||||||||||||\n" +
		"C|||||||||||||||\n" +
		"C|||||||||||||||\n" +
		"R|2|^^^Urina4^screening^^tempo-analisi-minuti|90|||||F|||||\n" +
		"L|1|N")
}
