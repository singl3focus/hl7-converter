# HL7-Converter 

Package for convert hl7 message to different modifications

<p> <center>
<img src="https://img.shields.io/badge/made_by-singl3focus-blue"> <img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat">
</center> </p>

## Usage

```go get github.com/singl3focus/hl7-converter```

```go
package main

import (
	"os"
	"log"
	"path/filepath"

	"github.com/singl3focus/hl7-converter"
)

const (
	configFilename = "config.json"

	configInputBlockType = "hl7_astm_hbl"
	configOutputBlockType = "hl7_mindray_hbl"
)

func main() {
	// Reference:
	// if you want convert msg withou same tag
	// use FullConvertMsg and send the full message to the function

	msg := []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
		"P|1||||^||||||||||||||||||||||||||||\n" +
		"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
		"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
		"L|1|N") 

	ready, err := hl7converter.FullConvertMsg(configFilename, configInputBlockType, configOutputBlockType, inputNewMsgHBL)
	if err != nil {
		log.Fatal(err)
	}

	// if you want to convert a message with the same tags that you needed,
	// split this message so that each sub-message contains all the service tags (non-repeating tags)
	// and one of the repeating tags in turn. Example:
	// [[]byte("H|1|2|\n" + "R||info|||\n" + "R||13144||||"), []byte("H|1|2|\n" + "R||info|||\n" + "R||13155||||")]
	//
	// After converting message with the same tags return converted msg which you could be assemble to your needed msg

	inMsgHL7CL1200Mult = [][]byte{[]byte("MSH|^~\\&|||||20120508150259||QRY^Q02|7|P|2.3.1||||||ASCII|||\n" + 
		"PID|1|1001|||Mike||19851001095133|M|||keshi|||||||||||||||beizhu|||||\n" +
		"OBR|1|12345678|10|^|Y|20120405193926|20120405193914|20120405193914|||||linchuangzhenduan|20120405193914|serum|lincyisheng|keshi||||||||3|||||||||||||||||||||||\n" +
		"OBX|1|NM|2|TBil|100| umol/L |-|N|||F||100|20120405194245||yishen|0|"),
	
		[]byte("MSH|^~\\&|||||20120508150259||QRY^Q02|7|P|2.3.1||||||ASCII|||\n" + 
		"PID|1|1001|||Mike||19851001095133|M|||keshi|||||||||||||||beizhu|||||\n" +
		"OBR|1|12345678|10|^|Y|20120405193926|20120405193914|20120405193914|||||linchuangzhenduan|20120405193914|serum|lincyisheng|keshi||||||||3|||||||||||||||||||||||\n" + 
		"OBX|2|NM|5|ALT|98.2| umol/L |-|N|||F||98.2|20120405194403||yishen|0|")}

	readyMsgs, err := hl7converter.FullConvertMsgWithSameTags(configFilename, configInputBlockType2, configOutputBlockType2, inMsgHL7CL1200Mult, "OBX")
	if err != nil {
		log.Fatal(err)
	}

	var finalLine string
	for i, msg := range readyMsgs {
		breakL := false
		if (i + 1) % 4 == 0 {
			finalLine += string(msg) // it's line with same tag and we get it and add to finalLine 
			breakL = true
		} else if i < 3 {
			finalLine += string(msg) // it's service tags (it's duplicate in every msg)
			breakL = true
		} else {
			breakL = false
		}

		if i != (len(readyMsgs) - 1) && breakL{
			finalLine += "\n"
		}
	}

	log.Println("Final result: ", finalLine)

	log.Printf("%s TestConvertMsg right", success)
}
```
**Remember, this is just an example of how you can use it, but it can differently used or something else.**


## Config file 
### Rules
- **See examples and fill it out like**

- ***Before convert your message Make sure that message not contain line separator (in case if line separator equal component separator). If you meet this you need change all component separator in message on other symbol (in config you need specified replaced symbol), and after converting replace component separator back***

- all keys in Json must be string
- if you specify multiple values in array, they must be separated by a comma ***(,)***
- in case if linked tags has more than one match will be choose first one 
- field "Delimeters" must be set manual with the help setup default_value, which must contains line with values of all separators 

\- Converter wiil be try set default value in this cases:
	- If set value from input field to output has been unssuccessful
	- If "linked_fields" is empty 


> if you want this not to happen, just delete the "default_value" field in the configuration


### Structure
- First comes the name of the config(*example: "device_protocol"*), then the delimiter fields and an array of tags

- If tag hasn't have Linked - write ["none"]

- Elements can contains default value
    - default value must be element not array 
    - link consists words which bonded by Component_separator specified in root hl7 modification

## Support
If you have any difficulties, problems or questions, you can just write to me by e-mail <tursunov.imran@mail.ru> or telegram <https://t.me/single_focus>.
