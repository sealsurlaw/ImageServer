package errs

import (
	"fmt"
	"net/http"
)

var ErrCannotConnectDatabase = fmt.Errorf("Cannot connect to the database.")

var ErrTokenAlreadyExists = fmt.Errorf("Token already exists.")

var ErrInvalidContentType = fmt.Errorf("Invalid content type.")

var ErrTooManyAttempts = fmt.Errorf("Too many attempts tried")

var ErrTokenNotFound = fmt.Errorf("Token not found.")

var ErrNotAuthorized = fmt.Errorf("Not authorized.")

var ErrTokenExpired = fmt.Errorf("Token expired.")

var ErrBadRequest = fmt.Errorf("Bad request.")

type ErrorResponse struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Error  string `json:"error,omitempty"`
}

func NewErrorResponse(code int, msg string, errs ...error) *ErrorResponse {
	response := &ErrorResponse{
		Code:   code,
		Status: http.StatusText(code),
		Msg:    msg,
	}

	if len(errs) > 0 {
		response.Error = errs[0].Error()
	}

	return response
}
