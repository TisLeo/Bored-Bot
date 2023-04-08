package web

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

var httpClient = http.Client{Timeout: time.Second * 10}

// Performs a http GET request, and returns the api response in the form of a struct T.
func GetToStruct[T any](url string) (*T, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if unmarshalled, err := unmarshal[T](string(data)); err != nil {
		return nil, err
	} else {
		return unmarshalled, nil
	}
}

// Unmarshalls the given response string to struct T
func unmarshal[T any](resp string) (*T, error) {
	var toReturn T

	if err := json.Unmarshal([]byte(resp), &toReturn); err != nil {
		return nil, err
	} else {
		return &toReturn, nil
	}
}
