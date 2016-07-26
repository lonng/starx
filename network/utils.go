package network

import "github.com/chrislonng/starx/log"

func serializeOrRaw(v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}

	data, err := serializer.Serialize(v)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return data, nil
}
