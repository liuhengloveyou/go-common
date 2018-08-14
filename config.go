package common

import (
	"encoding/json"
	"os"

	"github.com/BurntSushi/toml"
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

func LoadTomlConfig(fn string, config interface{}) error {
	if _, err := toml.DecodeFile(fn, config); err != nil {
		return err
	}

	return nil
}
