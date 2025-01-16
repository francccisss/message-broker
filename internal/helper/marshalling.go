package helper

import "encoding/json"

func ParseIncomingHeader(data []byte) (interface{}, error) {
	var v interface{}
	err := json.Unmarshal(data, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}
