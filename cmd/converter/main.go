package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/minerofish/go-hl7"
	"github.com/singl3focus/hl7-converter"
)

const (
	configFilename = "config.json"
	configInputBlockType = "hl7_astm_hbl"
	configOutputBlockType = "hl7_mindray_hbl"

	configInputBlockType2 = "cl1200_astm"
	configOutputBlockType2 = "access"
)

func main() {
	/*
	- ВЕРСИЮ В MSH добавлять с помошью hl7.IdentifyMessage и тип сообщения тоже
	
	*/

	log.Println("App started....")

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	configPath := filepath.Join(wd, configFilename)

	inputModification, err := hl7converter.ReadJSONConfigBlock(configPath, configInputBlockType2)
	if err != nil || inputModification == nil {
		log.Fatal(err)
	}

	outputModification, err := hl7converter.ReadJSONConfigBlock(configPath, configOutputBlockType)
	if err != nil || outputModification == nil {
		log.Fatal(err)
	}

	mock := []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20221011\n" +
	"P|1||||^||||||||||||||||||||||||||||\n" +
	"O|1|4^???. 1 - ???. 4||^^^Urine 10*4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
	"R|1|^^^Urine 10*4^screening^^tempo-analisi-minuti|195|||||F|||||\n" +
	"R|2|^^^Urine 10*4^screening^^carica-alfanumerica|-Negative|||||F|||||\n" +
	"R|3|^^^Urine 10*4^screening^^carica-numerica|0|||||F|||||\n" +
	"R|4|^^^Urine 10*4^screening^^torbido|F|||||F|||||\n" +
	"R|5|^^^Urine 10*4^screening^^anomalo|F|||||F|||||\n" +
	"R|6|^^^Urine 10*4^screening^^invalido|F|||||F|||||\n" +
	"R|7|^^^Urine 10*4^screening^^assente|F|||||F|||||\n" +
	"R|8|^^^Urine 10*4^screening^^fine-anticipata|F|||||F|||||\n" +
	"R|9|^^^Urine 10*4^screening^^percentuale-agitazione|95|||||F|||||\n" +
	"L|1|N")

	msg, err := hl7converter.ConvertMsg(inputModification, outputModification, mock)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(msg))
	// _______________________

	mock = []byte("MSH|^~\\|Manufacturer|Model|||20070719145353||ORU^R01|2|P|2.3.1||||0||ASCII|||")
	messageType, protocolVersion, err := hl7.IdentifyMessage(mock, hl7.EncodingUTF8)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Message type:", messageType)
	log.Println("Protocol version:", protocolVersion)
	
	// // _______________________
	// x := 10
	// log.Println(reflect.ValueOf(x))
	
	// // _______________________
	// mock = []byte("O|1|1^???. 1 - ???. 1||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||")
	
	// err = hl7converter.ParseData(header, string(mock), hl7converter.ConfigNameOrder, )
	// if err != nil {
	// 	log.Fatal(err)
	// }
}