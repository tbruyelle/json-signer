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
	if k.IsAmino() {
		return k.Info.GetPubKey(), nil
	}
	return k.Record.GetPubKey()
}

func (k Key) GetPrivKey() (cryptotypes.PrivKey, error) {
	if k.IsAmino() {
		return privKeyFromInfo(k.Info)
	}
	return privKeyFromRecord(k.Record)
}

func (k Key) IsAmino() bool {
	return k.Info != nil
}
