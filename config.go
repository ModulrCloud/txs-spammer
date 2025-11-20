package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	NodeURL               string         `json:"nodeUrl"`
	TickMS                int            `json:"tickMs"`
	TransactionsPerSender int            `json:"transactionsPerSender"`
	RequestTimeoutMS      int            `json:"requestTimeoutMs"`
	DefaultPayload        map[string]any `json:"defaultPayload"`
	Senders               []Sender       `json:"senders"`
}

type Sender struct {
	Name       string         `json:"name"`
	PublicKey  string         `json:"publicKey"`
	PrivateKey string         `json:"privateKey"`
	Recipients []string       `json:"recipients"`
	Amount     uint64         `json:"amount"`
	Fee        uint64         `json:"fee"`
	Version    uint           `json:"version"`
	Payload    map[string]any `json:"payload"`
}

func loadConfig(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.NodeURL == "" {
		return nil, fmt.Errorf("nodeUrl is required")
	}
	if cfg.TransactionsPerSender == 0 {
		cfg.TransactionsPerSender = 1
	}

	return &cfg, nil
}
