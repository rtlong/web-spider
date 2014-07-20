package spider

import (
	"fmt"
	"runtime"
)

type Error struct {
	Error      error
	Message    string
	SourceFile string
	SourceLine int
}

func NewError(message string, err error) (newError Error) {
	newError = Error{
		Error:   err,
		Message: message,
	}
	_, file, line, ok := runtime.Caller(1)
	if ok {
		newError.SourceFile = file
		newError.SourceLine = line
	}
	return
}

func (e *Error) String() string {
	if e.SourceFile != "" {
		return fmt.Sprintf("%s - %s (%s:%d)", e.Message, e.Error, e.SourceFile, e.SourceLine)
	} else {
		return fmt.Sprintf("%s - %s (unknown source)", e.Message, e.Error)
	}
}
