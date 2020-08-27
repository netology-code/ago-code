package main

import (
	"errors"
	"fmt"
)

var ErrOriginal = errors.New("original error")

func main() {
	err := simpleWrap(9999)
	fmt.Println(err == ErrOriginal)          // false
	fmt.Println(errors.Is(err, ErrOriginal)) // true
}

func simpleWrap(arg interface{}) error {
	err := apiCall(arg)
	if err != nil {
		return fmt.Errorf(
			"api error: %w with %v",
			err,
			arg,
		)
	}
	return nil
}

func apiCall(interface{}) error {
	return ErrOriginal
}

