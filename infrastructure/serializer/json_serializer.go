package serializer

import "encoding/json"

type Serializer interface {
	Serialize(data any) ([]byte, error)
	Deserialize(buffer []byte, data any) error
}

type JsonSerializer struct {
}

func (s *JsonSerializer) Serialize(data any) ([]byte, error) {
	return json.Marshal(data)
}

func (s *JsonSerializer) Deserialize(buffer []byte, data any) error {
	return json.Unmarshal(buffer, data)
}
