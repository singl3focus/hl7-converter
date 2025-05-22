package examples

import (
	"os"
	"path/filepath"

	"github.com/singl3focus/hl7-converter/v2"
)

// I don't recommend to use init(), but it's using here in order to you can copy example and and run it immediately.
var workDir string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}

	workDir = wd
}

var (
	configPath            = filepath.Join(workDir, hl7converter.CfgJSON)
	configInputBlockType  = "astm_hbl"
	configOutputBlockType = "mindray_hbl"
)

func main() {
	// Some func which get message from anyone source
	getInputMessage := func () []byte  {
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

	convParams, err := hl7converter.NewConverterParams(configPath, configInputBlockType, configOutputBlockType)
	if err != nil {
		panic(err) // Change panic to error handler or return err (if it's not main)
	}

	c, err := hl7converter.NewConverter(convParams, hl7converter.WithUsingPositions(), hl7converter.WithUsingAliases())
	if err != nil {
		panic(err)
	}

	inputMsgHBL := getInputMessage()

	msgType, err := hl7converter.IndetifyMsg(convParams, inputMsgHBL)
	if err != nil {
		panic(err)
	}

	ready, err := c.Convert(inputMsgHBL)
	if err != nil {
		panic(err)
	}

	switch msgType { // just example of usage
	case "type_1":
		func(*hl7converter.Result) (proccessedMessage []byte, typeMessage string){
			proccessedMessage = append(ready.Bytes(), byte('\r'))
			typeMessage = msgType

			return proccessedMessage, typeMessage
		}(ready)
	case "type_2":
		func(*hl7converter.Result) {
			_, _ = ready.Aliases()
		}(ready)
	case "type_3":
		func(*hl7converter.Result) (proccessedMessage []byte, err error) {
			script := `msg.Rows[0].Fields[1].ChangeValue("TYPE");`
			err = ready.UseScript(hl7converter.KeyScript, script)
			if err != nil {
				return nil, err
			}

			proccessedMessage = ready.Bytes()
			
			return proccessedMessage, nil
		}(ready)
	}
}