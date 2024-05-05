package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/tbruyelle/legacykey/codec"
	"github.com/tbruyelle/legacykey/keyring"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
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
	signBytes, err := getSignBytes(tx, chainID, account, sequence)
	if err != nil {
		return err
	}
	// Sign those bytes
	privKey, err := key.GetPrivKey()
	if err != nil {
		return err
	}
	signature, err := privKey.Sign(signBytes)
	if err != nil {
		return err
	}

	// Construct the SignatureV2 struct
	sigData := signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: signature,
	}

	sigV2 = signing.SignatureV2{
		PubKey:   priv.PubKey(),
		Data:     &sigData,
		Sequence: accSeq,
	}

	return nil
}

func getSignBytes(tx Tx, chainID string, account, sequence uint64) ([]byte, error) {
	feeBytes, err := codec.Amino.MarshalJSON(tx.AuthInfo.Fee)
	if err != nil {
		return nil, err
	}
	msgsBytes := make([]json.RawMessage, 0, len(tx.Body.Messages))
	for _, msg := range tx.Body.Messages {
		bz := legacytx.RegressionTestingAminoCodec.MustMarshalJSON(msg)
		msgsBytes = append(msgsBytes, mustSortJSON(bz))
	}
	bz, err := codec.Amino.MarshalJSON(legacytx.StdSignDoc{
		AccountNumber: account,
		ChainID:       chainID,
		Fee:           json.RawMessage(feeBytes),
		Memo:          tx.Body.Memo,
		Msgs:          msgsBytes,
		Sequence:      sequence,
		TimeoutHeight: tx.Body.TimeoutHeight,
	})
	if err != nil {
		return nil, err
	}
	return mustSortJSON(bz)
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
		Messages      []json.RawMessage
		Memo          string
		TimeoutHeight string `json:"timeout_height"`
	}
	AuthInfo struct {
		SignerInfos []SignerInfo `json:"signer_infos"`
		Fee         struct {
			Amount   []Coin
			GasLimit string `json:"gas_limit"`
			Payer    string
			Granter  string
		}
	} `json:"auth_info"`
	Signatures []string
}

type SignerInfo struct {
	PublicKey any `json:"public_key"`
	ModeInfo  struct {
		Single struct {
			Mode string
		}
	} `json:"mode_info"`
	Sequence string
}

type Coin struct {
	Denom  string
	Amount string
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
