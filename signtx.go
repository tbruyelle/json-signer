package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/tbruyelle/legacykey/keyring"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func signTx(txFile, keyringDir, signer string, account, sequence uint64) error {
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

	return nil
}

type Tx struct {
	Body struct {
		Messages      []map[string]any
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
