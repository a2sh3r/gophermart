package apperrors

import "errors"

var (
	ErrInvalidRequest     = errors.New("invalid request")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInternalServer     = errors.New("internal server error")
	ErrInvalidAuthHeader  = errors.New("invalid or missing Authorization header")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrInvalidCredentials = errors.New("invalid login or password")
)
