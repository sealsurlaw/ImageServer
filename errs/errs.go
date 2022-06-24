package errs

import (
	"fmt"
	"net/http"
)

var ErrCannotConnectDatabase = fmt.Errorf("cannot connect to the database")

var ErrBadEncryptionSecret = fmt.Errorf("bad encryption secret")

var ErrTokenAlreadyExists = fmt.Errorf("token already exists")

var ErrInvalidContentType = fmt.Errorf("invalid content type")

var ErrTooManyAttempts = fmt.Errorf("too many attempts tried")

var ErrTokenNotFound = fmt.Errorf("token not found")

var ErrNotAuthorized = fmt.Errorf("not authorized")

var ErrTokenExpired = fmt.Errorf("token expired")

var ErrBadRequest = fmt.Errorf("bad request")

var ErrGif = fmt.Errorf("image is a gif")

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
