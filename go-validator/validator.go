package go_validator

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

const (
	validateTag  = "validate"
	inOperation  = "in"
	lenOperation = "len"
	minOperation = "min"
	maxOperation = "max"
)

var (
	ErrNotStruct                   = errors.New("wrong argument given, should be a struct")
	ErrInvalidValidatorSyntax      = errors.New("invalid validator syntax")
	ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")
	ErrLenValidationFailed         = errors.New("len validation failed")
	ErrInValidationFailed          = errors.New("in validation failed")
	ErrMaxValidationFailed         = errors.New("max validation failed")
	ErrMinValidationFailed         = errors.New("min validation failed")
	ErrUnsupportedType             = errors.New("unsupported type to validate")
	ErrUnsupportedOperation        = errors.New("unsupported validation operation")
	ErrUnsupportedOperationForType = errors.New("this operation is not supported for this type")
)

var SupportedOperations = []string{inOperation, lenOperation, minOperation, maxOperation}

type ValidationError struct {
	field string
	err   error
}

func NewValidationError(err error, field string) error {
	return &ValidationError{
		field: field,
		err:   err,
	}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.field, e.err)
}

func (e *ValidationError) Unwrap() error {
	return e.err
}

func parseQuery(query string) (opName string, args []string, err error) {
	ss := strings.Split(query, ":")
	if len(ss) != 2 {
		return "", nil, fmt.Errorf("%w: too many colons (%d)", ErrInvalidValidatorSyntax, len(ss)-1)
	}
	if !slices.Contains(SupportedOperations, ss[0]) {
		return "", nil, fmt.Errorf("%w: unsupported operation (%s)", ErrUnsupportedOperation, ss[0])
	}
	if len(ss[1]) == 0 {
		return "", nil, fmt.Errorf("%w: zero arguments for operation is provided", ErrInvalidValidatorSyntax)
	}
	args = strings.Split(ss[1], ",")
	if ss[0] != inOperation && len(args) != 1 {
		return "", nil, fmt.Errorf("%w: too many arguments (%d) for operation \"%s\"", ErrInvalidValidatorSyntax, len(args), ss[0])
	}

	return ss[0], args, nil
}

func executeValidating[T SupportedTypes](v Validating[T], name string, args []string) error {
	if name == inOperation {
		mp := map[T]struct{}{}
		for _, s := range args {
			x, e := v.Parse(s)
			if e != nil {
				return fmt.Errorf("%w: can't parse argument \"%s\" %w", ErrInvalidValidatorSyntax, s, e)
			}
			mp[x] = struct{}{}
		}
		return v.In(mp)
	}

	i, e := strconv.Atoi(args[0])
	if e != nil {
		return fmt.Errorf("%w: can't parse int argument \"%s\" %w", ErrInvalidValidatorSyntax, args[0], e)
	}
	switch name {
	case lenOperation:
		if i < 0 {
			return fmt.Errorf("%w: negative value for len operation (%d)", ErrInvalidValidatorSyntax, i)
		}
		return v.Len(i)
	case minOperation:
		return v.Min(i)
	case maxOperation:
		return v.Max(i)
	default:
		return fmt.Errorf("%w: (%s)", ErrUnsupportedOperation, name)
	}
}

func validateKind(v reflect.Value, opName string, args []string) error {
	switch v.Kind() {
	case reflect.Int:
		return executeValidating[int](IntValidating(v.Int()), opName, args)
	case reflect.String:
		return executeValidating[string](StringValidating(v.String()), opName, args)
	case reflect.Slice:
		for i := range v.Len() {
			e := validateKind(v.Index(i), opName, args)
			if e != nil {
				return e
			}
		}
		return nil
	default:
		return fmt.Errorf("%w: (%s)", ErrUnsupportedType, v.Kind().String())
	}
}

func validateQuery(v reflect.Value, query string) error {
	name, args, e := parseQuery(query)
	if e != nil {
		return e
	}
	return validateKind(v, name, args)
}

func Validate(s any) error {
	if reflect.ValueOf(s).Kind() != reflect.Struct {
		return ErrNotStruct
	}
	st := reflect.TypeOf(s)
	sv := reflect.ValueOf(s)

	var es []error
	for i := range st.NumField() {
		field := st.Field(i)
		if query, has := field.Tag.Lookup(validateTag); has {
			if !field.IsExported() {
				es = append(es, NewValidationError(ErrValidateForUnexportedFields, field.Name))
			}
			e := validateQuery(sv.Field(i), query)
			if e != nil {
				es = append(es, NewValidationError(e, field.Name))
			}
		}
	}

	if len(es) == 0 {
		return nil
	}
	return errors.Join(es...)
}
