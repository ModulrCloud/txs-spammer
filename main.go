package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to load generator config")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	client := &http.Client{Timeout: cfg.requestTimeout()}
	nonceTracker := newNonceTracker(cfg.Senders)

	if err := nonceTracker.Prime(context.Background(), client, cfg.NodeURL); err != nil {
		log.Fatalf("failed to initialize nonces: %v", err)
	}

	ticker := time.NewTicker(cfg.tick())
	defer ticker.Stop()

	log.Printf("starting tx load generator: %d txs per sender every %s against %s", cfg.TransactionsPerSender, cfg.tick(), cfg.NodeURL)

	for range ticker.C {
		for _, sender := range cfg.Senders {
			for i := 0; i < cfg.TransactionsPerSender; i++ {
				tx, err := cfg.newTransaction(sender, nonceTracker)
				if err != nil {
					log.Printf("skip tx for %s: %v", sender.PublicKey, err)
					continue
				}

				if err := submitTransaction(context.Background(), client, cfg.NodeURL, tx); err != nil {
					log.Printf("tx submit failed (%s -> %s nonce %d): %v", tx.From, tx.To, tx.Nonce, err)
					continue
				}

				nonceTracker.Increment(sender.PublicKey)
				log.Printf("submitted tx: from=%s to=%s amount=%d fee=%d nonce=%d", tx.From, tx.To, tx.Amount, tx.Fee, tx.Nonce)
			}
		}
	}
}

func submitTransaction(ctx context.Context, client *http.Client, nodeURL string, tx Transaction) error {
	txJSON, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("serialize tx: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(nodeURL, "/")+"/transaction", strings.NewReader(string(txJSON)))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (cfg *Config) newTransaction(sender Sender, nonces *nonceTracker) (Transaction, error) {
	nextNonce, ok := nonces.Next(sender.PublicKey)
	if !ok {
		return Transaction{}, fmt.Errorf("nonce for %s not initialized", sender.PublicKey)
	}

	payload := cfg.DefaultPayload
	if len(sender.Payload) > 0 {
		payload = sender.Payload
	}

	recipient, err := sender.pickRecipient()
	if err != nil {
		return Transaction{}, err
	}

	tx := Transaction{
		V:       sender.version(),
		From:    sender.PublicKey,
		To:      recipient,
		Amount:  sender.Amount,
		Fee:     sender.Fee,
		Nonce:   nextNonce,
		Payload: payload,
	}
	tx.Sig, err = sign(tx, sender.PrivateKey)
	if err != nil {
		return Transaction{}, err
	}
	return tx, nil
}

func (cfg *Config) tick() time.Duration {
	if cfg.TickMS == 0 {
		return time.Second
	}
	return time.Duration(cfg.TickMS) * time.Millisecond
}

func (cfg *Config) requestTimeout() time.Duration {
	if cfg.RequestTimeoutMS == 0 {
		return 10 * time.Second
	}
	return time.Duration(cfg.RequestTimeoutMS) * time.Millisecond
}

func (s *Sender) version() uint {
	if s.Version == 0 {
		return 1
	}
	return s.Version
}

func (s *Sender) pickRecipient() (string, error) {
	if len(s.Recipients) == 0 {
		return "", errors.New("no recipients configured")
	}
	if len(s.Recipients) == 1 {
		return s.Recipients[0], nil
	}
	return s.Recipients[rand.Intn(len(s.Recipients))], nil
}
