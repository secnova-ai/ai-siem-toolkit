package main

import (
	"encoding/json"
	sdk "example.local/tool/sdk"
	"fmt"
	"os"
)

func run() error {
	var in struct {
		Params map[string]any `json:"params"`
		Creds  struct {
			BaseURL string `json:"base_url"`
			APIKey  string `json:"api_key"`
		} `json:"creds"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		return err
	}
	if in.Creds.BaseURL == "" || in.Creds.APIKey == "" {
		return fmt.Errorf("base_url and api_key are required")
	}
	r, err := sdk.Get(in.Creds.BaseURL+"/health", map[string]string{"X-API-Key": in.Creds.APIKey})
	if err != nil {
		return err
	}
	if !r.OK {
		return fmt.Errorf("health API returned HTTP %d", r.Status)
	}
	var result struct {
		Status string `json:"status"`
	}
	if err = json.Unmarshal(r.Body, &result); err != nil {
		return fmt.Errorf("health API returned invalid JSON")
	}
	if result.Status == "" {
		return fmt.Errorf("health API omitted status")
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
