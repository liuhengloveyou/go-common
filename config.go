package common

import (
	"encoding/json"
	"os"

	yaml "gopkg.in/yaml.v3"
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

func LoadYamlConfig(fn string, config interface{}) (err error) {
	r, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer r.Close()

	decoder := yaml.NewDecoder(r)
	err = decoder.Decode(config)

	return
}
