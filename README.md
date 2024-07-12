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
	"strings"
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

	ready, msgType, err := hl7converter.ConvertWithConverter(configFilename, configInputBlockType, configOutputBlockType, inputMsgHBL)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	res := make([]byte, 0, 1024)
	for i, rowFields := range ready {
		readyRow := strings.Join(rowFields, "|")

		log.Logf("%d row: %v\n", i+1, readyRow)
		res = append(res, []byte(readyRow)...)

		if i < (len(ready) - 1) {
			res = append(res, []byte("\r")...)
		}
	}

	
	log.Println("Final result: ", string(res))

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

- Converter wiil be try set default value in this cases:
	- If set value from input field to output has been unsuccessful 
	- If "linked_fields" is empty 


> if you want this not to happen, just delete the "default_value" field in the configuration


### Structure
- First comes the name of the config(*example: "device_protocol"*), then the delimiter fields and an array of tags

- If tag hasn't have Linked - write [""]

- Elements can contains default value
    - default value must be element not array 
    - link consists words which bonded by Component_separator specified in root hl7 modification

## Support
If you have any difficulties, problems or questions, you can just write to me by telegram <https://t.me/single_focus>.
