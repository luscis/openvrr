package sub

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

func OutJson(data interface{}) error {
	if out, err := json.Marshal(data); err == nil {
		fmt.Println(string(out))
	} else {
		return err
	}
	return nil
}

func OutYaml(data interface{}) error {
	if out, err := yaml.Marshal(data); err == nil {
		fmt.Println(string(out))
	} else {
		return err
	}
	return nil
}

func Out(data interface{}, format string) error {
	switch format {
	case "json":
		return OutJson(data)
	default:
		return OutYaml(data)
	}
}
