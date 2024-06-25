package main

import (
	"log"
	"github.com/minerofish/go-hl7"
	"github.com/singl3focus/hl7-converter"
)

const (
	configFilename = "config.json"
	configInputBlockType = "hl7_astm_hbl"
	configOutputBlockType = "hl7_mindray_hbl"
)

var inputMsgHBL = []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
	"P|1||||^||||||||||||||||||||||||||||\n" +
	"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
	"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
	"L|1|N")

func main() {
	/*
	- ВЕРСИЮ В MSH добавлять с помошью hl7.IdentifyMessage и тип сообщения тоже
	
	*/

	log.Println("App started....")

	
	ready_str, err := hl7converter.FullConvertMsg(configFilename, configInputBlockType, configOutputBlockType, inputMsgHBL)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(ready_str))
	
	
	// _______________________

	mock := []byte("MSH|^~\\|Manufacturer|Model|||20070719145353||ORU^R01|2|P|2.3.1||||0||ASCII|||")
	messageType, protocolVersion, err := hl7.IdentifyMessage(mock, hl7.EncodingUTF8)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Message type:", messageType)
	log.Println("Protocol version:", protocolVersion)
}