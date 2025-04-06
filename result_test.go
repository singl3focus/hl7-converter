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
        wantComponents []string
        wantArray      []*hl7converter.Field
    }{
        {
            name:       "Ok",
            fieldValue: "sireAstmCom^1^P/LIS02^20241021",
            wantComponents: []string{"sireAstmCom", "1", "P", "LIS02", "20241021"},
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
            res := hl7converter.NewField(tt.fieldValue, componentSep, componentArrSep)

            // --- Проверка компонентов ---
            components := res.Components()
            if !reflect.DeepEqual(components, tt.wantComponents) {
                t.Fatalf("Components mismatch:\ngot: %v\nwant: %v", components, tt.wantComponents)
            }

			// --- Проверка массива ---
            array := res.Array()
            if len(array) != len(tt.wantArray) {
                t.Fatalf("Array length mismatch: got %d, want %d", len(array), len(tt.wantArray))
            }

			// --- Проверка компонентов (дополнительно) ---
			components2 := res.Components() // Убедимся, что повторный вызов Components() не изменяет данные
			if !reflect.DeepEqual(components, components2) {
				t.Error("Components changed after second call")
			}

			res.ChangeValue("new^value")
			componentsAfterChange := res.Components()
			expected := []string{"new", "value"}
			if !reflect.DeepEqual(componentsAfterChange, expected) {
				t.Error("Components cache not reset after ChangeValue")
			}

			// --- Проверка полей массива ---
            for i, field := range array {
                if field.Value != tt.wantArray[i].Value {
                    t.Errorf("Array[%d].Value mismatch: got %q, want %q", i, field.Value, tt.wantArray[i].Value)
                }

                components := field.Components()
                expectedComponents := strings.Split(tt.wantArray[i].Value, componentSep)
                if !reflect.DeepEqual(components, expectedComponents) {
                    t.Errorf("Array[%d].Components mismatch:\ngot: %v\nwant: %v", i, components, expectedComponents)
                }
            }

            t.Logf("------ Success %s ------", tt.name)
        })
    }
}