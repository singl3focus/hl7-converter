package hl7converter

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// JsonExtension is required extension for config files used by NewConverterParams.
const JsonExtension = ".json"

var _ error = Error{} // compile-time check of error implementation

// Error wraps root error with optional caller and additional info. Stable part of public API.
type Error struct {
	Err            error
	Caller         string
	AdditionalInfo string
}

// NewError constructs Error; if addCaller is true, fills Caller with caller function name.
func NewError(reason error, addCaller bool, info ...string) Error {
	Assert(reason != nil, "init Error with nil error as reason, info %v", info)

	var additional strings.Builder
	for i, v := range info {
		if i != 0 {
			additional.WriteString("; ")
		}
		additional.WriteString(v)
	}

	err := Error{
		Err:            reason,
		AdditionalInfo: additional.String(),
	}

	if addCaller {
		pc, _, _, ok := runtime.Caller(1)
		details := runtime.FuncForPC(pc)

		if ok && details != nil {
			err.Caller = details.Name()
		}
	}

	return err
}

func (e Error) Error() string {
	return fmt.Sprintf("%s, %s", e.Err.Error(), e.AdditionalInfo)
}

// Is allows errors.Is to match wrapped Err.
func (e Error) Is(target error) bool {
	return e.Err.Error() == target.Error()
}

func Assert(condition bool, format string, fields ...any) {
	if !condition {
		panic(fmt.Sprintf(format, fields...))
	}
}

// Manual errors
var (
	ErrUndefinedScannerFailure = errors.New("unknown scanner failure")
	ErrParseFailure            = errors.New("parse failure")
)
