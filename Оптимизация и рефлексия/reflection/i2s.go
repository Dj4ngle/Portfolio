package main

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func interfaceToStruct(data interface{}, out interface{}) error {
	val := reflect.ValueOf(out)

	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New("output must be a non-nil pointer")
	}
	val = val.Elem()

	// Проверяем, что out действительно является структурой
	if val.Kind() != reflect.Struct {
		return errors.New("output must be a struct")
	}

	// Проверяем, что data действительно является map[string]interface{}
	mapData, ok := data.(map[string]interface{})
	if !ok {
		return errors.New("data must be a map[string]interface{}")
	}

	for key, value := range mapData {
		field := val.FieldByName(strings.Title(key))
		if !field.IsValid() || !field.CanSet() {
			continue
		}

		requiredType := field.Type()
		v := reflect.ValueOf(value)

		// Проверяем, нужно ли рекурсивно обрабатывать структуры
		if v.Kind() == reflect.Map && requiredType.Kind() == reflect.Struct {
			subStructPtr := reflect.New(requiredType)
			err := i2s(value, subStructPtr.Interface())
			if err != nil {
				return err
			}
			field.Set(subStructPtr.Elem())
			continue
		}

		// Обработка слайсов
		if v.Kind() == reflect.Slice && requiredType.Kind() == reflect.Slice {
			elementType := requiredType.Elem()
			slice := reflect.MakeSlice(requiredType, 0, v.Len())

			for i := 0; i < v.Len(); i++ {
				element := reflect.New(elementType).Elem()
				err := i2s(v.Index(i).Interface(), element.Addr().Interface())
				if err != nil {
					return err
				}
				slice = reflect.Append(slice, element)
			}

			field.Set(slice)
			continue
		}

		if requiredType != v.Type() {
			// Пытаемся преобразовать float к int
			switch requiredType.Kind() {
			case reflect.Int:
				if v.Type().Kind() == reflect.Float64 {
					intValue := int(v.Float())
					field.Set(reflect.ValueOf(intValue))
				} else {
					return fmt.Errorf("cannot convert %v to int", v.Type())
				}
			default:
				return fmt.Errorf("unsupported type conversion from %v to %v", v.Type(), requiredType)
			}
		} else {
			field.Set(v)
		}
	}

	return nil
}

func i2s(data interface{}, out interface{}) error {
	// Проверяем, что out является указателем на структуру или слайс
	val := reflect.ValueOf(out)

	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New("output must be a non-nil pointer")
	}
	val = val.Elem()

	switch val.Kind() {
	case reflect.Slice:
		// Если это слайс, обходим все элементы
		slice, ok := data.([]interface{})
		if !ok {
			return errors.New("input data should be a slice of interface{}")
		}

		if val.Len() != len(slice) {
			val.Set(reflect.MakeSlice(val.Type(), len(slice), len(slice)))
		}

		for i := 0; i < val.Len(); i++ {
			item := val.Index(i).Addr()
			err := interfaceToStruct(slice[i], item.Interface())
			if err != nil {
				return err
			}
		}
	case reflect.Struct:
		// Если это структура, вызываем interfaceToStruct напрямую
		err := interfaceToStruct(data, val.Addr().Interface())
		if err != nil {
			return err
		}
	default:
		return errors.New("out must be a pointer to struct or slice of structs")
	}

	return nil
}
