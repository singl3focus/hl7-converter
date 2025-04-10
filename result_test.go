package hl7converter_test

import (
	"reflect"
	"strings"
	"testing"

	hl7converter "github.com/singl3focus/hl7-converter/v2"
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
            if !reflect.DeepEqual(components, tt.wantComponents) {
                t.Fatalf("Components mismatch:\ngot: %v\nwant: %v", components, tt.wantComponents)
            }

			// --- Array check ---
            array := res.Array()
            if len(array) != len(tt.wantArray) {
                t.Fatalf("Array length mismatch: got %d, want %d", len(array), len(tt.wantArray))
            }

			// --- Additional components check ---
			components2 := res.Components() // Убедимся, что повторный вызов Components() не изменяет данные
			if !reflect.DeepEqual(components, components2) {
				t.Error("Components changed after second call")
			}

			res.ChangeValue("new^value")
			componentsAfterChange := res.Components()
			expected := hl7converter.Components{"new", "value"}
			if !reflect.DeepEqual(componentsAfterChange, expected) {
				t.Error("Components cache not reset after ChangeValue")
			}

			// --- Проверка полей массива ---
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