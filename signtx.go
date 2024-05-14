package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/tbruyelle/legacykey/codec"
	"github.com/tbruyelle/legacykey/keyring"
	"golang.org/x/exp/maps"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

func signTx(tx Tx, kr keyring.Keyring, signer, chainID string, account, sequence uint64) (Tx, error) {
	key, err := kr.Get(signer + ".info")
	if err != nil {
		return Tx{}, err
	}
	// Get bytesToSign from tx
	bytesToSign, err := getBytesToSign(tx, chainID, account, sequence)
	if err != nil {
		return Tx{}, err
	}
	// fmt.Println("BYTESTOSIGN", string(bytesToSign))

	// Sign those bytes
	signature, pubKey, err := key.Sign(bytesToSign)
	if err != nil {
		return Tx{}, err
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
				Mode: "SIGN_MODE_LEGACY_AMINO_JSON",
			},
		},
		Sequence: fmt.Sprint(sequence),
	}}
	return tx, nil
}

var protoToAminoTypeMap = map[string]string{
	"/cosmos.bank.v1beta1.MsgSend":          "cosmos-sdk/MsgSend",
	"/govgen.gov.v1beta1.MsgSubmitProposal": "cosmos-sdk/MsgSubmitProposal",
	"/govgen.gov.v1beta1.TextProposal":      "cosmos-sdk/TextProposal",
}

func getBytesToSign(tx Tx, chainID string, account, sequence uint64) ([]byte, error) {
	fee := tx.AuthInfo.Fee
	gas, err := strconv.ParseUint(fee.GasLimit, 10, 64)
	if err != nil {
		return nil, err
	}
	stdFee := legacytx.StdFee{
		Gas:     gas,
		Payer:   fee.Payer,
		Granter: fee.Granter,
	}
	for _, a := range fee.Amount {
		amount, ok := math.NewIntFromString(a.Amount)
		if !ok {
			return nil, fmt.Errorf("Cannot get math.Int from %q", a.Amount)
		}
		stdFee.Amount = append(stdFee.Amount, sdk.NewCoin(a.Denom, amount))
	}
	msgsBytes := make([]json.RawMessage, 0, len(tx.Body.Messages))
	for _, msg := range tx.Body.Messages {
		bz, err := json.Marshal(protoToAminoJSON(msg))
		if err != nil {
			return nil, fmt.Errorf("marshalling aminoMsg: %v", err)
		}
		msgsBytes = append(msgsBytes, bz)
	}
	feeBytes, err := codec.Amino.MarshalJSON(stdFee)
	if err != nil {
		return nil, err
	}
	timeoutHeight, err := strconv.ParseUint(tx.Body.TimeoutHeight, 10, 64)
	if err != nil {
		return nil, err
	}

	bz, err := codec.Amino.MarshalJSON(legacytx.StdSignDoc{
		AccountNumber: account,
		ChainID:       chainID,
		Fee:           json.RawMessage(feeBytes),
		Memo:          tx.Body.Memo,
		Msgs:          msgsBytes,
		Sequence:      sequence,
		TimeoutHeight: timeoutHeight,
	})
	if err != nil {
		return nil, err
	}
	// TODO ensure this is really required, maybe we can just use the stdlib
	// json encoder
	return mustSortJSON(bz), nil
}

// WTH is that
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

// protoToAminoJSON turns proto json to amino json.
func protoToAminoJSON(m map[string]any) map[string]any {
	if protoType, ok := m["@type"]; ok {
		aminoType, ok := protoToAminoTypeMap[protoType.(string)]
		if !ok {
			panic(fmt.Sprintf("can't find amino mapping for proto @type=%q", protoType))
		}
		m := maps.Clone(m)
		delete(m, "@type")
		return map[string]any{
			"type":  aminoType,
			"value": protoToAminoJSON(m),
		}
	}
	for k, v := range m {
		if mm, ok := v.(map[string]any); ok {
			m[k] = protoToAminoJSON(mm)
		}
	}
	return m
}

type Tx struct {
	Body       Body     `json:"body"`
	AuthInfo   AuthInfo `json:"auth_info"`
	Signatures [][]byte `json:"signatures"`
}

type Body struct {
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

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}
