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

	inputModification, err := hl7converter.ReadJSONConfigBlock(configPath, configInputBlockType)
	if err != nil || inputModification == nil {
		log.Fatal(err)
	}

	outputModification, err := hl7converter.ReadJSONConfigBlock(configPath, configOutputBlockType)
	if err != nil || outputModification == nil {
		log.Fatal(err)
	}

	mock := []byte("MSH|~\\&|Manufacturer|Model|||20220327||ORU~R01||P|2.3.1||||||ASCII|" + "^" +
	"OBR||142212|||||||||||||URI|||||||||||||||||||||||||||" + "^" + "OBX|||Urina4~screening~tempo-analisi-minuti|tempo-analisi-minuti|180||||||F|||||")

 	msg, err := hl7converter.ConvertMsg(outputModification, inputModification, mock)
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
}