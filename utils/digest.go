package utils

import "encoding/json"

func Digest(object interface{}) (string, error) {
	msg, err := json.Marshal(object)

	if err != nil {
		return "", nil
	}
	return Hash(msg), nil
}
