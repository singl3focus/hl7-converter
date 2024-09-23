# HL7-Converter 

Package for convert hl7 message to different modifications.
At the same time, the converter depends only on the config file, so it has extensive conversion capabilities.

Some information about HL7: https://rhapsody.health/blog/complete-guide-to-hl7-standards/

***Now Package correct work only with ASCII symbols***

<p> <center>
<img src="https://img.shields.io/badge/made_by-singl3focus-blue"> <img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat">
</center> </p>

## Main Idea 
You send full message as []byte to "Converter" and convert it with "Converter.Convert".  
As a response, you will receive rows splited by field separator - this is done so that it is convenient to work on the line on top of the conversion. And after any your manipulation on any row, you can assemble it with "Converter.AssembleOutput".

## Usage
1. **Get package**
```go get github.com/singl3focus/hl7-converter@TAG```

- Tag representations: \
vX.X.X - for using on many platforms \
vX.X.X-go1.20 - for building on Windows 7 

2. **Example of converting**:
```
Input (you send full message as []byte):
	"H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
	"P|1||||^||||||||||||||||||||||||||||\n" +
	"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
	"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
	"R|2|^^^Urina4^screening^^tempo-analisi-minuti|90|||||F|||||\n" +
	"L|1|N"

Output (you receive rows splited by field separator):
 	MSH|^\&|Manufacturer|Model|||20220327||ORU^R01||P|2.3.1||||||ASCII|
	PID||142212|||||||||||||||||||||||||||
    OBR||142212|||||||||||||URI|||||||||||||||||||||||||||
    OBX|||Urina4^screening^tempo-analisi-minuti|tempo-analisi-minuti|180||||||F|||||
    OBX|||Urina4^screening^tempo-analisi-minuti|tempo-analisi-minuti|90||||||F|||||
    
	message type: Results
```

```go
package main

import (
	"os"
	"log"
	"strings"
)

const (
	configFilename = "config.json"

	configInputBlockType = "hl7_astm"
	configOutputBlockType = "hl7_mindray"
)

func main() {
	msg := []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
		"P|1||||^||||||||||||||||||||||||||||\n" +
		"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
		"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
		"L|1|N") 

	convParams, err := hl7converter.NewConverterParams(configFilename, configInputBlockType, configOutputBlockType)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	ready, _, err := hl7converter.Convert(convParams, inputMsgHBL)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	/* YOU CAN ASSEMBLY MESSAGE OR CHANGE ANY FIELD JUST WITH INDEX*/
	
	// example of assemle message
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

### Advices
1) **When you are filling any modification that you should be guided by the fact that this Transformation will occur based on the output modification, and not the input one.**
2) ! **Now converter works by principle of create output rows based on Output Modification, it's means that you always can get rows according to the templates specified in Output Modification** !

### Rules
- **See examples and fill it out like**

#### Template filling

#### JSON
- All keys in JSON must be string

#### Rows
- ***Rows of message cannot contain a line separator otherwise the conversion will be incorrect. \
If you meet that line separator equal component separator, you need change all component separator in message on other symbol (in config you need specifie replaced symbol), and after converting replace component separator back*** 

#### Tags
- If you set `fields_numbe more than 0`, converter will be `compare fields_number and template fields_number` . If you set `fields_number = -1`, it's not be checking
- **It was decided that the `default_value (which specified with help 'OR' symbol)` of the field should not be substituted for the `template`, this was done so that the developer aimed at obtaining values through the template could be guaranteed to make sure that an error would come.**

[Deprecated]
- if you specify multiple values in array, they must be separated by a comma ***(,)***
- in case if linked tags has more than one match will be choose first one 
- field "Delimeters" must be set manual with the help setup default_value, which must contains line with values of all separators, else you can get incorrect conversion

- If message contains more than one multitag Converter.Convert return err. (Current version)

- Converter wiil be try set default value in this cases:
	- If set value from input field to output has been unsuccessful 
	- If "linked_fields" is empty 

> if you want this not to happen, just delete the "default_value" field in the configuration


### Structure [Deprecated]
- First comes the name of the config(*example: "device_protocol"*), then the delimiter fields and an array of tags.

- If tag hasn't have Linked - write [""]. It's will be means that this tag couldn't have links with other tags, but other tags could be have links with this tag.

- Elements can contains default value:
    - default value must be element, not array 

- linked_fields array consists words which bonded by "-". Structure element of linked_fields: TAG-POSITION (this params of input message)

### QA
❓ Is it possible to specify just the types (name of type), leave the field value empty, i.e. leave the tags empty to do a type check?
- No. Because the comparison takes place piecemeal and the meaning of the tags is important, if they are not specified, then the message will not be able to be identified in any way

❓ Is the "linked" field for the tag and "linked_fields" required for the fields?
- "linked" is required for the tag, but it may be empty => this tag does not refer to anything, but it can be referenced; "linked_fields" is optional

❓ How is a field linked if multiple tags are specified in linked_fields ["H-1", "H-2"]
- If several links are specified in linked_fields, then the values taken from these links will be connected using component_separator and written to the position of this field

❓ Are types required?
- No. It is important to know that the type of msg is determined both by the input message + input modification and by the converted message + output modification, conclusion: the main thing is to correctly pass the parameters to the function

❓ The conversion takes place exactly according to the fields specified in the config, or simply the initial message is divided by "|" and the field number is taken
- At the very beginning, the incoming message is parsed in a certain way for easy search, then the specified links are searched for a parsed structure

## Support
If you have any difficulties, problems or questions, you can just write to me by telegram <https://t.me/single_focus>.
