package keyring

import (
	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

type Key struct {
	Name string
	// Record is not nil if the key is proto-encoded
	Record *cosmoskeyring.Record
	// Info is not nil if the key is amino-encoded
	Info cosmoskeyring.LegacyInfo
}

func (k Key) GetPubKey() (cryptotypes.PubKey, error) {
	if k.Record != nil {
		return k.Record.GetPubKey()
	}
	return k.Info.GetPubKey(), nil
}

func (k Key) IsAmino() bool {
	return k.Info != nil
}
