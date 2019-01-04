// Copyright 2018 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package translate

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrSameType      = errors.New("Translating to and from same type")
	ErrNoTranslation = errors.New("No translation defined and implicit translation impossible")
)

func getFieldNameWithJsonTag(v reflect.Value, tag string) string {
	nFields := v.NumField()
	for i := 0; i < nFields; i++ {
		vtag, _ := v.Type().Field(i).Tag.Lookup("json")
		if vtag == tag {
			return v.Type().Field(i).Name
		}
	}
	return ""
}

// Check if "to" has all the fields that "from" has with the same json tags
func toIsSuperSetOfFrom(to, from reflect.Value) (bool, error) {
	// sanity check
	if to.Kind() != reflect.Struct || from.Kind() != reflect.Struct {
		return false, fmt.Errorf("Need structs....")
	}

	nFields := from.NumField()
	for i := 0; i < nFields; i++ {
		field := from.Type().Field(i)
		tag, ok := field.Tag.Lookup("json")
		if !ok || tag == "" { // technically redundant, better to be explicit
			continue // skip, structs are allow to have fields we don't care about
		}
		if getFieldNameWithJsonTag(to, tag) == "" {
			return false, nil
		}
	}
	return true, nil
}

type translator func(reflect.Value, reflect.Value) error

func getTranslator(to, from reflect.Value) translator {
	return nil
}

func translateStruct(to, from reflect.Value) error {
	if f := getTranslator(to, from); f != nil {
		return f(to, from)
	}
	if isSuper, err := toIsSuperSetOfFrom(to, from); err != nil {
		return err
	} else if isSuper {
		return translateStructImplicitly(to, from)
	}
	return ErrNoTranslation
}

func translateStructImplicitly(to, from reflect.Value) error {
	// We only care about fields with json tags
	if to.Kind() != reflect.Struct || from.Kind() != reflect.Struct {
		return fmt.Errorf("Need structs....")
	}

	nFields := from.NumField()
	for i := 0; i < nFields; i++ {
		field := from.Type().Field(i)
		tag, ok := field.Tag.Lookup("json")
		if !ok || tag == "" { // technically redundant, better to be explicit
			continue // skip, structs are allow to have fields we don't care about
		}
		toFieldName := getFieldNameWithJsonTag(to, tag)
		if toFieldName == "" {
			return fmt.Errorf("internal error")
		}

		toField := to.FieldByName(toFieldName)
		fmt.Println(toField.CanAddr())
		if err := Translate(toField, from.Field(i)); err != nil {
			return err
		}
	}
	return nil
}

func translateElem(to, from reflect.Value) error {
	if to.Kind() == reflect.Ptr && !to.IsNil() {
		return Translate(reflect.Indirect(to), reflect.Indirect(from))
	}
	if to.Type() != from.Type() {
		if to.Type().Name() == from.Type().Name() && from.Type().ConvertibleTo(to.Type()) {
			to.Set(from.Convert(to.Type()))
			return nil
		}
		return ErrNoTranslation
	}
	to.Set(from)
	return nil
}

func translateSlice(to, from reflect.Value) error {
	nElems := from.Len()
	fmt.Println(to)
	if from.IsNil() {
		return nil
	}
	to.Set(reflect.MakeSlice(reflect.SliceOf(from.Type().Elem()), nElems, nElems))
	for i := 0; i < nElems; i++ {
		Translate(to.Index(i), from.Index(i))
	}
	return nil
}

func Translate(to, from reflect.Value) error {
	switch from.Kind() {
	case reflect.Struct:
		return translateStruct(to, from)
	case reflect.Slice:
		return translateSlice(to, from)
	default:
		return translateElem(to, from)
	}
}
