package caddy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	APIEndpoint = "http://localhost:2019"
)

type API struct{}

func NewAPI() *API {
	return &API{}
}

func (a *API) Load(config *Config) error {
	configAsJson, err := json.Marshal(config)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		APIEndpoint+"/load",
		bytes.NewBuffer(configAsJson),
	)

	if err != nil {
		return err
	}

	req.Header.Set(
		"Content-Type",
		"application/json; charset=UTF-8",
	)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return err
		}

		return fmt.Errorf("caddy API error: %s", body)
	}

	return nil
}
