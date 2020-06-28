package client

import "fmt"

type Error struct {
	msg string
	err error
}

func (e *Error) Error() (message string) {
	return fmt.Sprintf("%s / %s", e.msg, e.err.Error())
}

func (e *Error) Unwrap() (included error) {
	return e.err
}

func errorwf(err error, msg string, a ...interface{}) (wrapped error) {
	return &Error{msg: fmt.Sprintf(msg, a...), err: err}
}
