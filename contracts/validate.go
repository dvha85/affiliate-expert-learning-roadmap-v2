// Package contracts validates against the repository's canonical embedded schemas.
package contracts

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed *.schema.json
var schemas embed.FS

const base = "https://schemas.local.invalid/"

var compiled = map[string]*jsonschema.Schema{}
var once sync.Once
var compileErr error

type offlineLoader struct{}

func (offlineLoader) Load(url string) (any, error) {
	return nil, fmt.Errorf("schema resource not embedded: %s", url)
}

func compileSchemas() {
	c := jsonschema.NewCompiler()
	c.AssertFormat()
	c.UseLoader(offlineLoader{})
	entries, err := schemas.ReadDir(".")
	if err != nil {
		compileErr = err
		return
	}
	for _, entry := range entries {
		raw, err := schemas.ReadFile(entry.Name())
		if err != nil {
			compileErr = err
			return
		}
		value, err := Decode(raw)
		if err != nil {
			compileErr = err
			return
		}
		if err := c.AddResource(base+entry.Name(), value); err != nil {
			compileErr = err
			return
		}
	}
	for _, entry := range entries {
		schema, err := c.Compile(base + entry.Name())
		if err != nil {
			compileErr = err
			return
		}
		compiled[entry.Name()] = schema
	}
	for _, name := range []string{"history-record.schema.json#/properties/observations", "history-record.schema.json#/properties/recorded_result"} {
		schema, err := c.Compile(base + name)
		if err != nil {
			compileErr = err
			return
		}
		compiled[name] = schema
	}
}

// Decode rejects duplicate JSON object keys at every depth and trailing values.
// UseNumber avoids rounding IDs/numeric values before validation.
func Decode(raw []byte) (any, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	value, err := decodeValue(d, 0)
	if err != nil {
		return nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, fmt.Errorf("trailing or invalid JSON")
	}
	return value, nil
}

func decodeValue(d *json.Decoder, depth int) (any, error) {
	if depth > 128 {
		return nil, fmt.Errorf("JSON nesting exceeds 128")
	}
	token, err := d.Token()
	if err != nil {
		return nil, err
	}
	switch token {
	case json.Delim('{'):
		object := map[string]any{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return nil, err
			}
			name, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("invalid object key")
			}
			if _, exists := object[name]; exists {
				return nil, fmt.Errorf("duplicate JSON key %q", name)
			}
			value, err := decodeValue(d, depth+1)
			if err != nil {
				return nil, err
			}
			object[name] = value
		}
		_, err := d.Token()
		return object, err
	case json.Delim('['):
		array := []any{}
		for d.More() {
			value, err := decodeValue(d, depth+1)
			if err != nil {
				return nil, err
			}
			array = append(array, value)
		}
		_, err := d.Token()
		return array, err
	default:
		return token, nil
	}
}

// ValidateRaw validates the original JSON, before typed decoding loses fields.
func ValidateRaw(name string, raw []byte) error {
	value, err := Decode(raw)
	if err != nil {
		return err
	}
	return Validate(name, value)
}

func Validate(name string, value any) error {
	once.Do(compileSchemas)
	if compileErr != nil {
		return compileErr
	}
	schema, ok := compiled[name]
	if !ok {
		return fmt.Errorf("unknown embedded schema %q", name)
	}
	return schema.Validate(value)
}

// DecodeStrict preserves exact JSON key spelling (encoding/json alone accepts
// case-insensitive matches). Unknown typed fields are rejected, not discarded.
func DecodeStrict(raw []byte, target any) error {
	value, err := Decode(raw)
	if err != nil {
		return err
	}
	typ := reflect.TypeOf(target)
	if typ == nil || typ.Kind() != reflect.Pointer {
		return fmt.Errorf("target must be a pointer")
	}
	if err := checkKeys(value, typ.Elem()); err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}

func checkKeys(value any, typ reflect.Type) error {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if value == nil {
		return nil
	}
	switch typ.Kind() {
	case reflect.Struct:
		object, ok := value.(map[string]any)
		if !ok {
			return nil
		}
		fields := map[string]reflect.Type{}
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if field.PkgPath == "" && name != "-" {
				if name == "" {
					name = field.Name
				}
				fields[name] = field.Type
			}
		}
		for name, item := range object {
			field, ok := fields[name]
			if !ok {
				return fmt.Errorf("unknown or unsupported field %q", name)
			}
			if err := checkKeys(item, field); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		if array, ok := value.([]any); ok {
			for _, item := range array {
				if err := checkKeys(item, typ.Elem()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
