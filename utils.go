package starx

import (
	"bytes"
	"encoding/gob"

	"github.com/chrislonng/starx/log"
)

func serializeOrRaw(v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}

	data, err := serializer.Serialize(v)
	if err != nil {
		log.Errorf(err.Error())
		return nil, err
	}
	return data, nil
}

func gobEncode(args ...interface{}) ([]byte, error) {
	buf := bytes.NewBuffer([]byte(nil))
	if err := gob.NewEncoder(buf).Encode(args); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gobDecode(reply interface{}, data []byte) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(reply)
}
