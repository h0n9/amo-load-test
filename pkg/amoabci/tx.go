package amoabci

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex" // when encoding rpc request
	"encoding/json"
	"math/big"

	"github.com/amolabs/amo-client-go/lib/keys"
)

var (
	curve = elliptic.P256() // move to crypto sub-package
)

type TxToSign struct {
	Type       string          `json:"type"`
	Sender     string          `json:"sender"`
	Fee        string          `json:"fee"`
	LastHeight string          `json:"last_height"`
	Payload    json.RawMessage `json:"payload"`
}

type TxSig struct {
	Pubkey   string `json:"pubkey"`
	SigBytes string `json:"sig_bytes"`
}

type TxToSend struct {
	Type       string          `json:"type"`
	Sender     string          `json:"sender"`
	Fee        string          `json:"fee"`
	LastHeight string          `json:"last_height"`
	Payload    json.RawMessage `json:"payload"`
	Signature  TxSig           `json:"signature"`
}

func SignTx(txType string, payload interface{}, key keys.KeyEntry, fee, lastHeight string) ([]byte, error) {
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	txToSign := TxToSign{
		Type:       txType,
		Sender:     key.Address,
		Fee:        fee,
		LastHeight: lastHeight,
		Payload:    payloadJson,
	}
	msg, err := json.Marshal(txToSign)
	if err != nil {
		return nil, err
	}
	// do sign
	h := sha256.Sum256(msg)
	X, Y := curve.ScalarBaseMult(key.PrivKey[:])
	ecdsaPrivKey := ecdsa.PrivateKey{
		D: new(big.Int).SetBytes(key.PrivKey[:]),
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     X,
			Y:     Y,
		},
	}
	r, s, err := ecdsa.Sign(rand.Reader, &ecdsaPrivKey, h[:])
	if err != nil {
		return nil, err
	}
	rb := r.Bytes()
	sb := s.Bytes()
	sigBytes := make([]byte, 64)
	copy(sigBytes[32-len(rb):], rb)
	copy(sigBytes[64-len(sb):], sb)
	// done sign
	txSig := TxSig{
		Pubkey:   hex.EncodeToString(key.PubKey),
		SigBytes: hex.EncodeToString(sigBytes),
	}
	tx := TxToSend{
		Type:       txToSign.Type,       // forward
		Sender:     txToSign.Sender,     // forward
		Fee:        txToSign.Fee,        // forward
		LastHeight: txToSign.LastHeight, // forward
		Payload:    txToSign.Payload,    // forward
		Signature:  txSig,               // signature appendix
	}
	b, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	return b, nil
}
