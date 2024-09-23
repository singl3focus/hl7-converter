package hl7converter

import (
	"bytes"
	"strings"
)

type WrapperConverter struct {
	// Data parsed from config.
	// goal: find metadata(position, default_value, components_number, linked and data about Tags, Separators)
	// about field so that then get value some field.
	inM, outM *Modification

	// For effective split by rows of input message with help "bufio.Scanner"
	lineSplit func(data []byte, atEOF bool) (advance int, token []byte, err error)

	// Convertred structure of input message for fast find a needed field by tag
	inMsg *Msg
}

func NewWrapperConverter(c *Converter) (*WrapperConverter) {
	return &WrapperConverter{
		inM: c.Input,
		outM: c.Output,
		lineSplit: c.LineSplit,
		inMsg: c.MsgSource,
	}
}


// AssembleOutput
//
// _______[INFO]_______
// - Func assemble message without adding 'Output.LineSeparator' after end row.
//   If you need it just add 'LineSeparator' in returned value.
//
func (w *WrapperConverter) AssembleOutput(rows [][]string) []byte {
	var buf bytes.Buffer
	for i, row := range rows {
		readyRow := strings.Join(row, w.outM.FieldSeparator)

		buf.WriteString(readyRow)

		if i != (len(rows) - 1) {
			buf.WriteString(w.outM.LineSeparator)
		}
	}

	return buf.Bytes()
}