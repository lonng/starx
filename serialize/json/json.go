package json

import "encoding/json"

type Serialezer struct{}

func NewSerializer() *Serialezer {
	return &Serialezer{}
}

func (s *Serialezer) Serialize(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (s *Serialezer) Deserialize(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
