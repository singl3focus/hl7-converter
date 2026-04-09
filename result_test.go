package hl7converter_test

import (
	"testing"

	hl7converter "github.com/singl3focus/hl7-converter/v2"
	"github.com/stretchr/testify/assert"
)

func TestFieldComponentsAndArray(t *testing.T) {
	t.Parallel()

	t.Run("components flatten array separators and reset after change", func(t *testing.T) {
		t.Parallel()

		field := hl7converter.NewField("sireAstmCom^1^P/LIS02^20241021", "^", "/")

		assert.Equal(
			t,
			hl7converter.Components{"sireAstmCom", "1", "P", "LIS02", "20241021"},
			field.Components(),
		)
		assert.Equal(t, field.Components(), field.Components())

		field.ChangeValue("new^value")
		assert.Equal(t, hl7converter.Components{"new", "value"}, field.Components())
	})

	t.Run("array returns split fields and resets after change", func(t *testing.T) {
		t.Parallel()

		field := hl7converter.NewField("left^1/right^2", "^", "/")

		array := field.Array()
		if assert.Len(t, array, 2) {
			assert.Equal(t, "left^1", array[0].Value)
			assert.Equal(t, hl7converter.Components{"left", "1"}, array[0].Components())
			assert.Equal(t, "right^2", array[1].Value)
			assert.Equal(t, hl7converter.Components{"right", "2"}, array[1].Components())
		}

		assert.Equal(t, array, field.Array())

		field.ChangeValue("single")
		assert.Empty(t, field.Array())
	})

	t.Run("array without separator stays empty but components still work", func(t *testing.T) {
		t.Parallel()

		field := hl7converter.NewField("A^B^C", "^", "/")

		assert.Empty(t, field.Array())
		assert.Equal(t, hl7converter.Components{"A", "B", "C"}, field.Components())
	})
}
