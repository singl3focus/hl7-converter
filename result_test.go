package hl7converter_test

import (
	"reflect"
	"strings"
	"testing"

	hl7converter "github.com/singl3focus/hl7-converter/v2"
	"github.com/stretchr/testify/assert"
)


func TestNewField(t *testing.T) {
    var (
        componentSep    = "^"
        componentArrSep = "/"
    )

    tests := []struct {
        name           string
        fieldValue     string
        wantComponents hl7converter.Components
        wantArray      []*hl7converter.Field
    }{
        {
            name:       "Ok",
            fieldValue: "sireAstmCom^1^P/LIS02^20241021",
            wantComponents: hl7converter.Components{"sireAstmCom", "1", "P", "LIS02", "20241021"},
            wantArray: []*hl7converter.Field{
                {
                    Value: "sireAstmCom^1^P",
                },
                {
                    Value: "LIS02^20241021",
                },
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            res := hl7converter.NewField(tt.fieldValue, componentSep, componentArrSep)

            // --- Components check ---
            components := res.Components()
            assert.Equal(t, tt.wantComponents, components)

			// --- Array check ---
            array := res.Array()
            if len(array) != len(tt.wantArray) {
                t.Fatalf("array length mismatch: got %d, want %d", len(array), len(tt.wantArray))
            }

			// --- Additional components check ---
			repeatedComponents := res.Components() 
			assert.Equal(t, components, repeatedComponents)

			res.ChangeValue("new^value")

			componentsAfterChange := res.Components()
			newComponents := hl7converter.Components{"new", "value"}
			assert.Equal(t, newComponents, componentsAfterChange)

			// --- Array fields check---
            for i, field := range array {
                if field.Value != tt.wantArray[i].Value {
                    t.Errorf("Array[%d].Value mismatch: got %q, want %q", i, field.Value, tt.wantArray[i].Value)
                }

                components := field.Components()
                expectedComponents := hl7converter.Components(strings.Split(tt.wantArray[i].Value, componentSep))
                if !reflect.DeepEqual(components, expectedComponents) {
                    t.Errorf("Array[%d].Components mismatch:\ngot: %v\nwant: %v", i, components, expectedComponents)
                }
            }
        })
    }
}