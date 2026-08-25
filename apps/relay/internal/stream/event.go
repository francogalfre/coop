package stream

import "encoding/json"

type Event struct {
	Seq  int             `json:"seq"`
	Data json.RawMessage `json:"data"`
}
