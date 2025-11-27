package errors

import (
	"fmt"
)

// AppError 應用程式錯誤
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError 建立應用程式錯誤
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// 預定義錯誤
var (
	ErrUserNotFound      = NewAppError(CodeUserNotFound, "user not found", nil)
	ErrInvalidPassword   = NewAppError(CodeInvalidPassword, "invalid password", nil)
	ErrInvalidInvitation  = NewAppError(CodeInvalidInvitation, "invalid invitation code", nil)
	ErrUserExists        = NewAppError(CodeUserExists, "user already exists", nil)
	ErrChannelNotFound   = NewAppError(CodeChannelNotFound, "channel not found", nil)
	ErrProgramNotFound   = NewAppError(CodeProgramNotFound, "program not found", nil)
	ErrAccessDenied      = NewAppError(CodeAccessDenied, "access denied", nil)
	ErrInvalidAccessKey  = NewAppError(CodeInvalidAccessKey, "invalid access key", nil)
)

