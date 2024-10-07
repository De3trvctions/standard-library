// Package json Json解析封装
package json

import (
	"encoding/json"

	"github.com/tidwall/gjson"
)

type Typ interface {
	string | []byte | json.RawMessage
}

// Parse json string to json object
// 如果传入字符串转换时报错则v不会被赋值
func Parse[T Typ](data T, v any) {
	err := json.Unmarshal([]byte(data), v)
	if err != nil {
		return
	}
}

// ParseE json string to json object
// 如果传入字符串转换时报错则v不会被赋值，并且返回error
func ParseE[T Typ](data T, v any) error {
	err := json.Unmarshal([]byte(data), v)
	return err
}

// Stringify json object to json string []byte
// 如果传入对象转换时报错则返回nil
func Stringify(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

// String json object to json string
// 如果传入对象转换时报错则返回nil
func String(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// StringE json object to json string
// 如果传入对象转换时报错则返回nil,error
func StringE(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), err
}

// StringifyE json object to json string
// 如果传入对象转换时报错则返回nil，并且返回error
func StringifyE(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Get 获取json中某个path的值
func Get(v, path string) gjson.Result {
	return gjson.Get(v, path)
}

// Valid JSON是否有效
func Valid[T Typ](data T) bool {
	return json.Valid([]byte(data))
}
