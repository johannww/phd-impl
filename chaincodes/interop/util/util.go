package util

import (
	"encoding/json"
	"fmt"
)

func MarshallInvokeArgs(funcName string, args ...interface{}) ([][]byte, error) {
	result := [][]byte{[]byte(funcName)}
	for _, arg := range args {
		var argBytes []byte
		var err error
		switch v := arg.(type) {
		case []string:
			argBytes, err = json.Marshal(v)
			if err != nil {
				return nil, err
			}
		case string:
			argBytes = []byte(v)
		default:
			argBytes = fmt.Appendf(nil, "%v", v)
		}
		result = append(result, argBytes)
	}
	return result, nil
}
