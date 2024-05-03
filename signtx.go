package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/99designs/keyring"
	"github.com/bgentry/speakeasy"
	"github.com/davecgh/go-spew/spew"
)

func signTx(txFile, keyringDir, signer string, account, sequence uint64) error {
	tx, err := readTxFile(txFile)
	if err != nil {
		return err
	}
	spew.Dump(tx)
	kr, err := keyring.Open(keyring.Config{
		AllowedBackends: []keyring.BackendType{keyring.FileBackend},
		FileDir:         keyringDir,
		FilePasswordFunc: func(prompt string) (string, error) {
			return speakeasy.Ask(prompt + ": ")
		},
	})
	if err != nil {
		return err
	}
	key, err := kr.Get(signer)
	if err != nil {
		return err
	}
	_ = tx
	_ = key

	return nil
}

type Tx struct {
	Body struct {
		Messages      []map[string]any
		Memo          string
		TimeoutHeight uint64
	}
	AuthInfo struct {
		SignerInfos []struct {
			PublicKey any `json:"public_key"`
			ModeInfo  struct {
				Single struct {
					Mode string
				}
			} `json:"mode_info"`
			Sequence string
		} `json:"signer_infos"`
		Fee struct {
			Amount []struct {
				Denom  string
				Amount string
			}
			GasLimit string `json:"gas_limit"`
			Payer    string
			Granter  string
		}
	} `json:"auth_info"`
	Signatures []string
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
