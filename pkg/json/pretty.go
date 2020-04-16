package json

import (
	"bytes"
	"encoding/json"
)

func PrettyJson(i interface{}) ([]byte, error) {
	var bf bytes.Buffer
	e := json.NewEncoder(&bf)
	e.SetIndent("", " ")
	err := e.Encode(i)
	if err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}