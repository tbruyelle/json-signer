package main

import (
	"fmt"
	"io"

	"github.com/tbruyelle/keyring-compat"
	"gopkg.in/yaml.v2"
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
		return fmt.Errorf("read keyring keys: %w", err)
	}
	var list []keyOutput
	for _, key := range keys {
		encoding := "proto"
		if key.IsAminoEncoded() {
			encoding = "amino"
		}
		addr, err := key.Bech32Address(prefix)
		if err != nil {
			return fmt.Errorf("key.Bech32Address: %w", err)
		}
		bz, err := key.ProtoJSONPubKey()
		if err != nil {
			return fmt.Errorf("ProtoJSONPubKey: %w", err)
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
