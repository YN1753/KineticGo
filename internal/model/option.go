package model

import (
	"encoding/json"
)

type TempleConfig string

func (t TempleConfig) ToJson() json.RawMessage {
	return json.RawMessage(t)
}
