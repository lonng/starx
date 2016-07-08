package json

import "encoding/json"

type JsonSerialezer struct{}

func NewJsonSerializer() *JsonSerialezer {
	return &JsonSerialezer{}
}

func (s *JsonSerialezer) Serialize(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (s *JsonSerialezer) Deserialize(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
