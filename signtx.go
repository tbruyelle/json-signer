package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/tbruyelle/legacykey/codec"
	"github.com/tbruyelle/legacykey/keyring"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

func signTx(txFile, keyringDir, signer, chainID string, account, sequence uint64) error {
	tx, err := readTxFile(txFile)
	if err != nil {
		return err
	}
	spew.Dump(tx)
	kr, err := keyring.New(keyringDir, "")
	if err != nil {
		return err
	}
	key, err := kr.Get(signer)
	if err != nil {
		return err
	}
	pubKey, err := key.GetPubKey()
	if err != nil {
		return err
	}
	addr := sdk.AccAddress(pubKey.Address())
	signInfo := SignerInfo{
		PublicKey: pubKey,
	}
	_ = addr
	_ = signInfo
	// Prepare bytes to sign
	bytesToSign, err := getBytesToSign(tx, chainID, account, sequence)
	if err != nil {
		return err
	}
	fmt.Println("BYTESTOSIGN", base64.StdEncoding.EncodeToString(bytesToSign))
	fmt.Println(string(bytesToSign))

	// Sign those bytes
	privKey, err := key.GetPrivKey()
	if err != nil {
		return err
	}
	signature, err := privKey.Sign(bytesToSign)
	if err != nil {
		return err
	}
	signerInfo := SignerInfo{
		PublicKey: pubKey,
		Sequence:  fmt.Sprint(sequence),
	}
	signerInfo.ModeInfo.Single.Mode = "SIGN_MODE_LEGACY_AMINO_JSON"
	tx.AuthInfo.SignerInfos = []SignerInfo{signerInfo}
	tx.Signatures = [][]byte{signature}

	bz, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(bz))

	return nil
}

func getBytesToSign(tx Tx, chainID string, account, sequence uint64) ([]byte, error) {
	feeBytes, err := codec.Amino.MarshalJSON(tx.AuthInfo.Fee)
	if err != nil {
		return nil, err
	}
	msgsBytes := make([]json.RawMessage, 0, len(tx.Body.Messages))
	for _, msg := range tx.Body.Messages {
		bz, err := codec.Amino.MarshalJSON(msg)
		if err != nil {
			return nil, err
		}
		msgsBytes = append(msgsBytes, mustSortJSON(bz))
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
	Body struct {
		Messages      []json.RawMessage `json:"messages"`
		Memo          string            `json:"memo"`
		TimeoutHeight string            `json:"timeout_height"`
	} `json:"body"`
	AuthInfo struct {
		SignerInfos []SignerInfo `json:"signer_infos"`
		Fee         struct {
			Amount   []Coin `json:"amount,omitempty"`
			GasLimit string `json:"gas_limit,omitempty"`
			Payer    string `json:"payer,omitempty"`
			Granter  string `json:"granter,omitempty"`
		} `json:"fee"`
	} `json:"auth_info"`
	Signatures [][]byte `json:"signatures"`
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
