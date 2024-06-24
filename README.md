# HL7-Converter 

Package for convert hl7 message to different modifications

<p> <center>
<img src="https://img.shields.io/badge/made_by-singl3focus-blue"> <img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat">
</center> </p>

## Usage

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
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	configPath := filepath.Join(wd, configFilename) // Getted config path

    // Load needed modification from Config (input)
	inputModification, err := hl7converter.ReadJSONConfigBlock(configPath, configInputBlockType2)
	if err != nil || inputModification == nil {
		log.Fatal(err)
	} 

    // Load needed modification from Config (output)
	outputModification, err := hl7converter.ReadJSONConfigBlock(configPath, configOutputBlockType)
	if err != nil || outputModification == nil {
		log.Fatal(err)
	}

	var mock []byte = ...your getted msg 

	msg, err := hl7converter.ConvertMsg(inputModification, outputModification, mock)
	if err != nil {
		log.Fatal(err)
	}

	...your using msg
}
```
**Remember, this is just an example of how you can use it, but it can differently used or something else.**


## Config file 
### Rules
- **See examples and fill it out like**

- all keys in Json nust be string
- if you specify multiple values in array, they must be separated by a comma ***(,)***
- in case if linked tags has more than one match will be choose first one 
- field "Delimeters" must be set manual with the help setup default_value, which must contains line with values of ""

### Structure
- First comes the name of the config(*example: "device_protocol"*), then the delimiter fields and an array of tags

- Elements can contains default value
    - default value must be element not array 
    - link consists words which bonded by Component_separator specified in root hl7 modification

## Support
If you have any difficulties, problems or questions, you can just write to me by e-mail <tursunov.imran@mail.ru> or telegram <https://t.me/single_focus>.