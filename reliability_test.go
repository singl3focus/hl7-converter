package hl7converter_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/robertkrimen/otto"
	hl7 "github.com/singl3focus/hl7-converter/v2"
)

func auditMods(fieldSep string) (*hl7.Modification, *hl7.Modification) {
	input := &hl7.Modification{
		ComponentSeparator: "^", ComponentArrSeparator: "~",
		FieldSeparator: fieldSep, LineSeparator: "\n",
		TagsInfo: hl7.TagsInfo{
			Positions: map[string]string{"1": "A"},
			Tags:      map[string]hl7.Tag{"A": {Linked: "X", FieldsNumber: 2}},
		},
	}
	output := &hl7.Modification{
		ComponentSeparator: "^", ComponentArrSeparator: "~",
		FieldSeparator: fieldSep, LineSeparator: "\r",
		TagsInfo: hl7.TagsInfo{
			Positions: map[string]string{"1": "X"},
			Tags:      map[string]hl7.Tag{"X": {Linked: "A", FieldsNumber: 2, Tempalate: "X" + fieldSep + "<A-2>"}},
		},
	}
	return input, output
}

func auditConverter(t testing.TB, input, output *hl7.Modification, options ...hl7.OptionFunc) *hl7.Converter {
	t.Helper()
	if err := input.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := output.Validate(); err != nil {
		t.Fatal(err)
	}
	c, err := hl7.NewConverter(&hl7.ConverterParams{InputModification: input, OutputModification: output}, options...)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestRegressionParseInputDoesNotSilentlyTruncate(t *testing.T) {
	input, output := auditMods("|")
	c := auditConverter(t, input, output)
	result, err := c.ParseInput([]byte("A|first\nA|" + strings.Repeat("x", 70*1024)))
	if !errors.Is(err, hl7.ErrInvalidParseMsg) || result != nil {
		t.Fatalf("partial message accepted: %v %v", result, err)
	}
}

func TestRegressionValidatedOutputCountsAreExecutable(t *testing.T) {
	for _, count := range []int{-1, 0} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			input, output := auditMods("|")
			tag := output.TagsInfo.Tags["X"]
			tag.FieldsNumber = count
			output.TagsInfo.Tags["X"] = tag
			if _, err := hl7.NewConverter(&hl7.ConverterParams{InputModification: input, OutputModification: output}); err == nil {
				t.Fatalf("constructor accepted impossible output fields_number=%d", count)
			}
		})
	}
}

func TestRegressionValidatedPositionKeysDoNotLoseTags(t *testing.T) {
	input, output := auditMods("|")
	_ = input
	output.TagsInfo.Positions = map[string]string{"01": "X"}
	if err := output.Validate(); err == nil {
		t.Fatal("non-canonical position accepted")
	}
}

func TestRegressionMultiCharacterFieldSeparator(t *testing.T) {
	for _, positionDriven := range []bool{false, true} {
		t.Run(fmt.Sprint(positionDriven), func(t *testing.T) {
			input, output := auditMods("||")
			var options []hl7.OptionFunc
			if positionDriven {
				options = append(options, hl7.WithUsingPositions())
			}
			c := auditConverter(t, input, output, options...)
			result, err := c.Convert([]byte("A||one"))
			if err != nil {
				t.Fatalf("accepted separator || cannot be converted in this mode: %v", err)
			}
			if result.String() != "X||one" {
				t.Fatalf("unexpected output: %q", result.String())
			}
		})
	}
}

func TestRegressionIdentifyDoesNotMutateConfig(t *testing.T) {
	input, _ := auditMods("|")
	input.TagsInfo.Tags["B"] = hl7.Tag{FieldsNumber: 2}
	input.Types = map[string][][]string{"sample": {{"B", "A"}}}
	params := &hl7.ConverterParams{InputModification: input}
	before := append([]string(nil), input.Types["sample"][0]...)
	got, err := hl7.IndetifyMsg(params, []byte("A|one\nB|two"))
	if err != nil || got != "sample" {
		t.Fatalf("identify failed: %q %v", got, err)
	}
	if !reflect.DeepEqual(before, input.Types["sample"][0]) {
		t.Fatalf("identification modified shared config: before=%v after=%v", before, input.Types["sample"][0])
	}
}

func TestRegressionIdentifySharedParamsRace(t *testing.T) {
	input, _ := auditMods("|")
	tags := make([]string, 64)
	var message strings.Builder
	for i := range tags {
		tags[i] = fmt.Sprintf("TAG%04d", len(tags)-i)
		input.TagsInfo.Tags[tags[i]] = hl7.Tag{FieldsNumber: 2}
		fmt.Fprintf(&message, "%s|value\n", tags[i])
	}
	input.Types = map[string][][]string{"sample": {tags}}
	if err := input.Validate(); err != nil {
		t.Fatal(err)
	}
	params := &hl7.ConverterParams{InputModification: input}
	payload := []byte(message.String())
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			got, err := hl7.IndetifyMsg(params, payload)
			if err != nil || got != "sample" {
				t.Errorf("got %q %v", got, err)
			}
		}()
	}
	close(start)
	wg.Wait()
}

func TestRegressionErrorsIsDistinguishesSentinels(t *testing.T) {
	err := hl7.NewFatalErrOfConverting("sample")
	if errors.Is(err, hl7.ErrInvalidParseMsg) {
		t.Fatal("fatal conversion error also matches distinct ErrInvalidParseMsg sentinel because messages happen to be equal")
	}
	if !errors.Is(err, hl7.FatalErrOfConverting) {
		t.Fatal("root sentinel must match")
	}
}

var auditResultSink *hl7.Result
var auditVMSink *otto.Otto

func BenchmarkNewResult(b *testing.B) {
	rows := []*hl7.Row{hl7.NewRowWithFieldsValues("|", "^", "~", "X", "one")}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		auditResultSink = hl7.NewResult("\r", rows)
	}
}

func BenchmarkOttoInit(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		auditVMSink = otto.New()
	}
}

func BenchmarkConvertReused(b *testing.B) {
	input, output := auditMods("|")
	c := auditConverter(b, input, output)
	message := []byte("A|one\nA|two")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := c.Convert(message)
		if err != nil {
			b.Fatal(err)
		}
		auditResultSink = result
	}
}

func TestRegressionAllParsingEntrypointsPropagateSplitErrors(t *testing.T) {
	input, output := auditMods("|")
	failure := errors.New("injected split failure")
	for _, method := range []string{"ParseMsg", "ParseInput"} {
		t.Run(method, func(t *testing.T) {
			c := auditConverter(t, input, output)
			c.LineSplit = func(data []byte, atEOF bool) (int, []byte, error) {
				return 0, nil, failure
			}
			var err error
			if method == "ParseMsg" {
				_, err = c.ParseMsg([]byte("A|one"))
			} else {
				_, err = c.ParseInput([]byte("A|one"))
			}
			if err == nil {
				t.Fatal("parser swallowed an injected Scanner split error")
			}
		})
	}
}

func TestRegressionMultiDigitComponentLinks(t *testing.T) {
	input, output := auditMods("|")
	tag := output.TagsInfo.Tags["X"]
	tag.Tempalate = "X|<A-2.10>"
	output.TagsInfo.Tags["X"] = tag
	if err := output.Validate(); err != nil {
		t.Fatal(err)
	}
	c := auditConverter(t, input, output)
	result, err := c.Convert([]byte("A|c1^c2^c3^c4^c5^c6^c7^c8^c9^c10"))
	if err != nil {
		t.Fatal(err)
	}
	if result.String() != "X|c10" {
		t.Fatalf("accepted link to component 10 silently resolved to a different component: %q", result.String())
	}
}

func TestRegressionDuplicateTypeDefinitionsAreRejected(t *testing.T) {
	input, _ := auditMods("|")
	input.Types = map[string][][]string{"first": {{"A"}}, "second": {{"A"}}}
	if err := input.Validate(); !errors.Is(err, hl7.ErrAmbiguousMessageType) {
		t.Fatal(err)
	}
}

func TestRegressionConfigFilePreservesWrongComponentLink(t *testing.T) {
	config := `{
      "input":{"component_separator":"^","component_array_separator":"~","field_separator":"|","line_separator":"\n",
        "tags_info":{"positions":{"1":"A"},"tags":{"A":{"linked":"X","fields_number":2,"template":""}}}},
      "output":{"component_separator":"^","component_array_separator":"~","field_separator":"|","line_separator":"\r",
        "tags_info":{"positions":{"1":"X"},"tags":{"X":{"linked":"A","fields_number":2,"template":"X|<A-2.10>"}}}}
    }`
	filename := writeConfigFixture(t, config)
	params, err := hl7.NewConverterParams(filename, "input", "output")
	if err != nil {
		t.Fatal(err)
	}
	c, err := hl7.NewConverter(params)
	if err != nil {
		t.Fatal(err)
	}
	result, err := c.Convert([]byte("A|c1^c2^c3^c4^c5^c6^c7^c8^c9^c10"))
	if err != nil {
		t.Fatal(err)
	}
	if result.String() != "X|c10" {
		t.Fatalf("full config path accepted component 10 and returned %q", result.String())
	}
}

func TestLinkGrammarAndAliasComponents(t *testing.T) {
	for _, position := range []int{1, 9, 10, 12} {
		values := []string{"c1", "c2", "c3", "c4", "c5", "c6", "c7", "c8", "c9", "c10", "c11", "c12"}
		input, output := auditMods("|")
		tag := output.TagsInfo.Tags["X"]
		tag.Tempalate = fmt.Sprintf("X|<A-2.%d>", position)
		output.TagsInfo.Tags["X"] = tag
		c := auditConverter(t, input, output)
		result, err := c.Convert([]byte("A|" + strings.Join(values, "^")))
		if err != nil || result.String() != "X|"+values[position-1] {
			t.Fatalf("%v %v", result, err)
		}
		result = hl7.NewResult("\r", []*hl7.Row{hl7.NewRowWithFieldsValues("|", "^", "~", "X", strings.Join(values, "^"))})
		aliases := hl7.Aliases{"component": fmt.Sprintf("X-2.%d", position)}
		if err := result.ApplyAliases(aliases); err != nil {
			t.Fatal(err)
		}
		got, _ := result.Aliases()
		if got["component"] != values[position-1] {
			t.Fatalf("wrong alias %v", got)
		}
	}
	for _, position := range []string{"2e0", "2.0", "2.-1", "NaN", "Inf", "+2", "2.1.2", "2.", "1", "0", "999999999999999999999999"} {
		_, output := auditMods("|")
		tag := output.TagsInfo.Tags["X"]
		tag.Tempalate = "X|<A-" + position + ">"
		output.TagsInfo.Tags["X"] = tag
		if err := output.Validate(); err == nil {
			t.Fatalf("accepted %q", position)
		}
	}
}

func TestConstructorAndErrorChain(t *testing.T) {
	for _, params := range []*hl7.ConverterParams{nil, {}, {InputModification: &hl7.Modification{}}} {
		if _, err := hl7.NewConverter(params); err == nil {
			t.Fatal("invalid params accepted")
		}
	}
	input, output := auditMods("|")
	if _, err := hl7.NewConverter(&hl7.ConverterParams{InputModification: input, OutputModification: output}, nil); err == nil {
		t.Fatal("nil option accepted")
	}
	cause := &testCause{}
	wrapped := hl7.NewError(cause, false)
	var got *testCause
	if !errors.As(wrapped, &got) || got != cause {
		t.Fatal("root type lost")
	}
	for _, key := range []string{"01", "0", "-1", "+1", "bad"} {
		output.TagsInfo.Positions = map[string]string{key: "X"}
		if err := output.Validate(); err == nil {
			t.Fatalf("accepted position %q", key)
		}
		if _, err := output.OrderedPositionTags(); err == nil {
			t.Fatalf("ordered accepted position %q", key)
		}
	}
}

type testCause struct{}

func (*testCause) Error() string { return "cause" }

func FuzzTemplateSyntax(f *testing.F) {
	for _, seed := range []string{"<A-2.10>", "é<A-2>", "<<A-2>>", "A-2>", "<A-2"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, template string) { _, _ = hl7.TemplateParse(template) })
}

func FuzzIdentityMapping(f *testing.F) {
	input, output := auditMods("|")
	c := auditConverter(f, input, output)
	for _, seed := range []string{"", "one", "кириллица", "a^b", "trailing  "} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 16*1024 || strings.ContainsAny(value, "|\r\n") {
			t.Skip()
		}
		result, err := c.Convert([]byte("A|" + value))
		if err != nil {
			t.Fatal(err)
		}
		if result.String() != "X|"+value {
			t.Fatalf("mapping changed the data: got %q", result.String())
		}
	})
}

func TestStrictConstructorContracts(t *testing.T) {
	input, output := auditMods("|")
	params := &hl7.ConverterParams{InputModification: input, OutputModification: output}
	if _, err := hl7.NewConverter(nil); !errors.Is(err, hl7.ErrInvalidConfig) {
		t.Fatal(err)
	}
	for _, option := range []hl7.OptionFunc{
		func(c *hl7.Converter) { c.Input = nil },
		func(c *hl7.Converter) { c.LineSplit = nil },
	} {
		if _, err := hl7.NewConverter(params, option); !errors.Is(err, hl7.ErrInvalidConfig) {
			t.Fatal(err)
		}
	}
	for _, template := range []string{"BAD|<A-2>", "X|<A-2>|extra"} {
		tag := output.TagsInfo.Tags["X"]
		tag.Tempalate = template
		output.TagsInfo.Tags["X"] = tag
		if _, err := hl7.NewConverter(params); err == nil {
			t.Fatalf("accepted template %q", template)
		}
	}
}

func TestDeterministicTypeErrors(t *testing.T) {
	input, _ := auditMods("|")
	input.Types = map[string][][]string{"first": {{"A"}}, "second": {{"A"}}}
	if err := input.Validate(); !errors.Is(err, hl7.ErrAmbiguousMessageType) {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		name, err := hl7.IdentifyMsg(&hl7.ConverterParams{InputModification: input}, []byte("A|one"))
		if name != "" || !errors.Is(err, hl7.ErrAmbiguousMessageType) {
			t.Fatalf("%q %v", name, err)
		}
	}
}

func TestUnicodeTemplateLiteral(t *testing.T) {
	input, output := auditMods("|")
	tag := output.TagsInfo.Tags["X"]
	tag.Tempalate = "X|префикс:<A-2>:конец"
	output.TagsInfo.Tags["X"] = tag
	result, err := auditConverter(t, input, output).Convert([]byte("A|значение"))
	if err != nil || result.String() != "X|префикс:значение:конец" {
		t.Fatalf("%v %v", result, err)
	}
}
