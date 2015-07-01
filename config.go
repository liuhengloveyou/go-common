package common

import (
	"encoding/json"
	"os"
)

func LoadJsonConfig(fn string, config interface{}) error {
	r, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer r.Close()

	decoder := json.NewDecoder(r)
	if err := decoder.Decode(config); err != nil {
		return err
	}

	return nil
}
