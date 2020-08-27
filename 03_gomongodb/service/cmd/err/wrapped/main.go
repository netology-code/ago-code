package main

import (
	"errors"
	"fmt"
)

// тип для ошибок вызова API
type APIError struct {
	ID   int64
	Code int64
}

func (a APIError) Error() string {
	return fmt.Sprintf("api error: id - %d, code - %d", a.ID, a.Code)
}

var ErrOriginal = &APIError{
	ID:   10342,
	Code: 208,
}

// тип для ошибок уровня сервиса
type ServiceError struct {
	Err    error
	Params interface{}
}

func (s ServiceError) Error() string {
	return fmt.Sprintf("call with %v params got %v", s.Params, s.Err)
}

func (s ServiceError) Unwrap() error {
	return s.Err
}

func main() {
	err := simpleWrap(9999)
	fmt.Println(err == ErrOriginal)          // false
	fmt.Println(errors.Is(err, ErrOriginal)) // true
	var typedError *APIError
	ok := errors.As(err, &typedError) // важно: указатель на указатель!
	fmt.Println(ok)            // true
	fmt.Println(typedError.ID) // 10342
}

func simpleWrap(arg interface{}) error {
	err := apiCall(arg)
	if err != nil {
		return &ServiceError{
			Err:    err,
			Params: arg,
		}
	}
	return nil
}

func apiCall(interface{}) error {
	return ErrOriginal
}

