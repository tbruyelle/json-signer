package main

import (
	"encoding/json"
	"fmt"

	"github.com/tbruyelle/keyring-compat"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

const signModeAminoJSON = "SIGN_MODE_LEGACY_AMINO_JSON"

func signTx(tx Tx, kr keyring.Keyring, signer, chainID, account, sequence string) (Tx, []byte, error) {
	key, err := kr.Get(signer + ".info")
	if err != nil {
		return Tx{}, nil, err
	}
	// Get bytesToSign from tx
	bytesToSign, err := getBytesToSign(tx, chainID, account, sequence)
	if err != nil {
		return Tx{}, nil, err
	}

	// Sign those bytes
	signature, pubKey, err := key.Sign(bytesToSign)
	if err != nil {
		return Tx{}, nil, err
	}

	// Update tx with signature and signer infos
	tx.Signatures = [][]byte{signature}
	tx.AuthInfo.SignerInfos = []SignerInfo{{
		PublicKey: map[string]any{
			"@type": codectypes.MsgTypeURL(pubKey),
			"key":   pubKey.Bytes(),
		},
		ModeInfo: ModeInfo{
			Single: Single{
				Mode: signModeAminoJSON,
			},
		},
		Sequence: fmt.Sprint(sequence),
	}}
	return tx, bytesToSign, nil
}

// getBytesToSign creates the SignDoc from tx, and serializes it using the
// amino-json format.
func getBytesToSign(tx Tx, chainID, account, sequence string) ([]byte, error) {
	msgsBytes := make([]json.RawMessage, 0, len(tx.Body.Messages))
	for _, msg := range tx.Body.Messages {
		// This is the weak part of the program, where proto-json format from msg
		// is transformed into the amino-json format.
		x, err := protoToAminoJSON(msg)
		if err != nil {
			return nil, fmt.Errorf("protoToAminoJSON: %v", err)
		}
		bz, err := json.Marshal(x)
		if err != nil {
			return nil, fmt.Errorf("marshalling aminoMsg: %v", err)
		}
		msgsBytes = append(msgsBytes, mustSortJSON(bz))
	}
	feeBytes, err := json.Marshal(tx.AuthInfo.Fee.FeeToSign())
	if err != nil {
		return nil, err
	}
	signDoc := SignDoc{
		AccountNumber: account,
		ChainID:       chainID,
		Fee:           json.RawMessage(feeBytes),
		Memo:          tx.Body.Memo,
		Msgs:          msgsBytes,
		Sequence:      sequence,
	}
	if tx.Body.TimeoutHeight != "0" {
		// manual omit empty TimeoutHeight since it's represented by a string
		signDoc.TimeoutHeight = tx.Body.TimeoutHeight
	}
	bz, err := json.Marshal(signDoc)
	if err != nil {
		return nil, err
	}
	return mustSortJSON(bz), nil
}

// mustSortJSON ensures JSON canonicalization (at least for field ordering).
func mustSortJSON(bz []byte) []byte {
	var c any
	err := json.Unmarshal(bz, &c)
	if err != nil {
		panic(err)
	}
	js, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return js
}

type Tx struct {
	Body       Body     `json:"body"`
	AuthInfo   AuthInfo `json:"auth_info"`
	Signatures [][]byte `json:"signatures"`
}

type Body struct {
	// TODO use []any?
	Messages      []map[string]any `json:"messages"`
	Memo          string           `json:"memo"`
	TimeoutHeight string           `json:"timeout_height"`
}

type AuthInfo struct {
	SignerInfos []SignerInfo `json:"signer_infos"`
	Fee         Fee          `json:"fee"`
}

type Fee struct {
	Amount   []Coin `json:"amount,omitempty"`
	GasLimit string `json:"gas_limit,omitempty"`
	Payer    string `json:"payer,omitempty"`
	Granter  string `json:"granter,omitempty"`
}

// FeeToSign is the same as Fee except for the Gas field which must outputs as
// `gas` instead of `gas_limit` in the JSON format.
type FeeToSign struct {
	Amount  []Coin `json:"amount,omitempty"`
	Gas     string `json:"gas,omitempty"`
	Payer   string `json:"payer,omitempty"`
	Granter string `json:"granter,omitempty"`
}

func (f Fee) FeeToSign() FeeToSign {
	return FeeToSign{
		Amount:  f.Amount,
		Gas:     f.GasLimit,
		Payer:   f.Payer,
		Granter: f.Granter,
	}
}

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type SignerInfo struct {
	PublicKey any      `json:"public_key"`
	ModeInfo  ModeInfo `json:"mode_info"`
	Sequence  string   `json:"sequence"`
}

type ModeInfo struct {
	Single Single `json:"single"`
}

type Single struct {
	Mode string `json:"mode"`
}

type SignDoc struct {
	AccountNumber string            `json:"account_number"`
	Sequence      string            `json:"sequence"`
	TimeoutHeight string            `json:"timeout_height,omitempty"`
	ChainID       string            `json:"chain_id"`
	Memo          string            `json:"memo"`
	Fee           json.RawMessage   `json:"fee"`
	Msgs          []json.RawMessage `json:"msgs"`
}
