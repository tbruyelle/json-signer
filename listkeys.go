package main

import (
	"fmt"
	"io"

	"github.com/tbruyelle/keyring-compat"
	"gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

type keyOutput struct {
	Name     string `json:"name" yaml:"name"`
	Encoding string `json:"encoding" yaml:"encoding"`
	Type     string `json:"type" yaml:"type"`
	Address  string `json:"address" yaml:"address"`
	PubKey   string `json:"pubkey" yaml:"pubkey"`
}

func PrintKeys(w io.Writer, kr keyring.Keyring, prefix string) error {
	keys, err := kr.Keys()
	if err != nil {
		return err
	}
	var list []keyOutput
	for _, key := range keys {
		encoding := "proto"
		if key.IsAminoEncoded() {
			encoding = "amino"
		}
		pk, err := key.PubKey()
		if err != nil {
			return err
		}
		addr, err := bech32.ConvertAndEncode(prefix, pk.Address())
		if err != nil {
			return err
		}

		apk, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			return err
		}
		bz, err := codec.ProtoMarshalJSON(apk, nil)
		if err != nil {
			return err
		}
		list = append(list, keyOutput{
			Name:     key.Name(),
			Encoding: encoding,
			Address:  addr,
			Type:     key.Type().String(),
			PubKey:   string(bz),
		})
	}
	out, err := yaml.Marshal(list)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(out))
	return nil
}
