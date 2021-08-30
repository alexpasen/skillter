package model

import "fmt"


type AppError struct {
	What  string
	Where string
	Error error
}

func NewAppError(what, where string, err error) *AppError {
	return &AppError{}
}

func (a *AppError) String() string{
	return fmt.Sprintf("What: %s. Where: %s. Err=%s.", a.What, a.Where, a.Error.Error())
}
