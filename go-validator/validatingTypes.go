package go_validator

import (
	"fmt"
	"strconv"
)

type SupportedTypes interface {
	int | string
}

type Validating[T SupportedTypes] interface {
	Len(int) error
	Min(int) error
	Max(int) error
	In(map[T]struct{}) error
	Parse(string) (T, error)
}

type IntValidating int

func (IntValidating) Len(int) error {
	return fmt.Errorf("%w: operation \"len\" on type int", ErrUnsupportedOperationForType)
}

func (i IntValidating) Min(bound int) error {
	if int(i) < bound {
		return fmt.Errorf("%w: %d < %d", ErrMinValidationFailed, i, bound)
	}
	return nil
}

func (i IntValidating) Max(bound int) error {
	if int(i) > bound {
		return fmt.Errorf("%w: %d > %d", ErrMaxValidationFailed, i, bound)
	}
	return nil
}

func (i IntValidating) In(v map[int]struct{}) error {
	if _, has := v[int(i)]; !has {
		return fmt.Errorf("%w: %d is not in %s", ErrInValidationFailed, i, printMap(v))
	}
	return nil
}

func (i IntValidating) Parse(s string) (int, error) {
	return strconv.Atoi(s)
}

type StringValidating string

func (s StringValidating) Len(bound int) error {
	if len(s) != bound {
		return fmt.Errorf("%w: len(%s) == %d != %d", ErrLenValidationFailed, s, len(s), bound)
	}
	return nil
}

func (s StringValidating) Min(bound int) error {
	if len(s) < bound {
		return fmt.Errorf("%w: len(%s) == %d < %d", ErrMinValidationFailed, s, len(s), bound)
	}
	return nil
}

func (s StringValidating) Max(bound int) error {
	if len(s) > bound {
		return fmt.Errorf("%w: len(%s) == %d > %d", ErrMaxValidationFailed, s, len(s), bound)
	}
	return nil
}

func (s StringValidating) In(v map[string]struct{}) error {
	if _, has := v[string(s)]; !has {
		return fmt.Errorf("%w: \"%s\" is not in %s", ErrInValidationFailed, s, printMap(v))
	}
	return nil
}

func (s StringValidating) Parse(x string) (string, error) {
	return x, nil
}

func printMap[T SupportedTypes](mp map[T]struct{}) string {
	keys := make([]T, 0, len(mp))
	for k := range mp {
		keys = append(keys, k)
	}
	return fmt.Sprintf("%v", keys)
}
