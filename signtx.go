package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/tbruyelle/legacykey/codec"
	"github.com/tbruyelle/legacykey/keyring"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

func signTx(txFile, keyringDir, signer, chainID string, account, sequence uint64) error {
	tx, err := readTxFile(txFile)
	if err != nil {
		return err
	}
	kr, err := keyring.New(keyringDir, "")
	if err != nil {
		return err
	}
	key, err := kr.Get(signer)
	if err != nil {
		return err
	}
	// Prepare bytes to sign
	bytesToSign, err := getBytesToSign(tx, chainID, account, sequence)
	if err != nil {
		return err
	}
	// fmt.Println("BYTESTOSIGN", string(bytesToSign))

	// Sign those bytes
	privKey, err := key.GetPrivKey()
	if err != nil {
		return err
	}
	signature, err := privKey.Sign(bytesToSign)
	if err != nil {
		return err
	}

	// Update tx with signature and signer infos
	tx.Signatures = [][]byte{signature}
	pubKey := privKey.PubKey()
	tx.AuthInfo.SignerInfos = []SignerInfo{{
		PublicKey: map[string]any{
			"@type": codectypes.MsgTypeURL(pubKey),
			"key":   pubKey.Bytes(),
		},
		Sequence: fmt.Sprint(sequence),
	}}
	tx.AuthInfo.SignerInfos[0].ModeInfo.Single.Mode = "SIGN_MODE_LEGACY_AMINO_JSON"

	// Output tx
	bz, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(bz))

	return nil
}

var protoToAminoTypeMap = map[string]string{
	"/cosmos.bank.v1beta1.MsgSend": "cosmos-sdk/MsgSend",
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
		protoType := msg["@type"].(string)
		aminoType, ok := protoToAminoTypeMap[protoType]
		if !ok {
			return nil, fmt.Errorf("Can't find amino mapping for proto @type=%q", protoType)
		}
		aminoValue := make(map[string]any)
		for k, v := range msg {
			if k != "@type" {
				aminoValue[k] = v
			}
		}
		aminoMsg := map[string]any{
			"type":  aminoType,
			"value": aminoValue,
		}

		// TODO try to use stdlib json encoder (and then remove call to mustSortJSON)
		// bz, err := codec.Amino.MarshalJSON(aminoMsg)
		bz, err := json.Marshal(aminoMsg)
		if err != nil {
			return nil, fmt.Errorf("marshalling aminoMsg: %v", err)
		}
		msgsBytes = append(msgsBytes, mustSortJSON(bz))
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
	PublicKey any `json:"public_key"`
	ModeInfo  struct {
		Single struct {
			Mode string `json:"mode"`
		} `json:"single"`
	} `json:"mode_info"`
	Sequence string `json:"sequence"`
}

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

func readTxFile(txFile string) (Tx, error) {
	f, err := os.Open(txFile)
	if err != nil {
		return Tx{}, err
	}
	defer f.Close()
	var tx Tx
	if err := json.NewDecoder(f).Decode(&tx); err != nil {
		return Tx{}, fmt.Errorf("JSON decode %s: %v", txFile, err)
	}
	return tx, nil
}
