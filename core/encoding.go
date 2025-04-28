package core

import "encoding/json"

// DecodeJSON unmarshals JSON data into the provided interface
func DecodeJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// EncodeJSON marshals the provided interface into JSON bytes
func EncodeJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return data
}
