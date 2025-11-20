# Modulr load generator

A small helper utility for the Modulr test network that periodically sends a batch of pre-signed transactions to a node, simulating user activity.

## Setup

1. Copy `config.example.json` to `config.json` and adjust parameters if needed: node URL, send interval, fee/amount, list of senders and recipients.
2. In `recipients`, specify addresses from your genesis. The example uses keys and balances from `templates/testnet_1/genesis.json` and matching private keys from `templates/testnet_*/configs_for_nodes`.

## Run

```bash
cd tx-loadgen

# Download dependencies and build the utility
go mod tidy

go run . -config ./config.json
```

Every `tickMs` milliseconds the utility fetches the latest nonce for each sender via the HTTP endpoint `/account/{address}`, forms `transactionsPerSender` transactions, signs them with the Ed25519 key (PKCS#8, base64), and sends them to `POST /transaction` on the configured node.

## Configuration

* `nodeUrl` — HTTP address of the node.
* `tickMs` — send interval in milliseconds.
* `transactionsPerSender` — how many transactions are sent from each sender per tick.
* `requestTimeoutMs` — HTTP client timeout.
* `defaultPayload` — optional payload used if the sender has no custom payload.
* `senders` — array of senders with fields `publicKey`, `privateKey`, `recipients`, `amount`, `fee`, `version`, and optional `payload`.