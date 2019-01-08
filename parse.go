package chirp

import (
	"encoding"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

// Parse request body, chi path parts and query parameters into dst
func Parse(req *http.Request, dst interface{}) error {
	value := reflect.ValueOf(dst)
	if value.Kind() != reflect.Ptr {

		return errors.New("dst should be an pointer")
	}

	if err := parse(req, value); err != nil {

		return err
	}

	return nil
}

func parse(req *http.Request, value reflect.Value) *Error {
	if err := parseQuery(req, value); err != nil {
		err.part = "query"

		return err
	}
	if err := parseBody(req, value); err != nil {
		err.part = "body"

		return err
	}
	if err := parseURLParams(req, value); err != nil {
		err.part = "path"

		return err
	}

	return nil
}

func parseQuery(req *http.Request, value reflect.Value) *Error {

	return withStructTag(value, "query", false, func(tag string, field reflect.Value) *Error {

		return unmarshalValue(req.URL.Query().Get(tag), field)
	})
}

func parseBody(req *http.Request, value reflect.Value) *Error {
	defer req.Body.Close()

	switch req.Header.Get("Content-Type") {
	case "application/x-www-form-urlencoded", "multipart/form-data":

		return parseFormData(req, value)

	default:
		body, _ := ioutil.ReadAll(req.Body)
		if err := json.Unmarshal(body, value.Interface()); err != nil {

			return &Error{
				source: string(body),
				cause:  errors.WithStack(err),
			}
		}

		return nil
	}
}

func parseFormData(req *http.Request, value reflect.Value) *Error {

	return withStructTag(value, "json", true, func(tag string, field reflect.Value) *Error {
		if tag == "-" {

			return nil
		}

		return unmarshalValue(req.FormValue(tag), field)
	})
}

func parseURLParams(req *http.Request, value reflect.Value) *Error {

	return withStructTag(value, "path", false, func(tag string, field reflect.Value) *Error {

		return unmarshalValue(chi.URLParam(req, tag), field)
	})
}

func withStructTag(
	value reflect.Value, tagName string, fieldNameOnEmpty bool,
	cb func(tag string, field reflect.Value) *Error,
) *Error {

	elem := value.Elem()
	reflectType := elem.Type()
	if reflectType.Kind() != reflect.Struct {

		return nil
	}
	for i := 0; i < reflectType.NumField(); i++ {
		fieldType := reflectType.Field(i)
		tag, ok := tagOrFieldName(fieldType, tagName)
		if !fieldNameOnEmpty && !ok {
			continue
		}
		if err := cb(tag, elem.Field(i)); err != nil {
			err.field = fieldType

			return err
		}
	}

	return nil
}

func unmarshalValue(source string, value reflect.Value) *Error {
	if len(source) == 0 {

		return nil
	}
	dst := reflect.New(value.Type()).Interface()
	if unmarshaler, ok := dst.(encoding.TextUnmarshaler); ok {
		if err := unmarshaler.UnmarshalText([]byte(source)); err != nil {

			return &Error{
				source: source,
				cause:  errors.WithStack(err),
			}
		}
	} else {
		if _, err := fmt.Sscan(source, dst); err != nil {

			return &Error{
				source: source,
				cause:  errors.WithStack(err),
			}
		}
	}

	value.Set(reflect.ValueOf(dst).Elem())

	return nil
}

func tagOrFieldName(field reflect.StructField, tagName string) (string, bool) {
	tag := strings.TrimSpace(strings.SplitN(field.Tag.Get(tagName), ",", 1)[0])
	if len(tag) > 0 {

		return tag, true
	}

	return field.Name, false
}
