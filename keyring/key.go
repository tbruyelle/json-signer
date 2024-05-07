package keyring

import (
	"fmt"

	"github.com/tbruyelle/legacykey/codec"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
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

func (k Key) ToProto() ([]byte, error) {
	// Turn record to legacyInfo
	info, err := legacyInfoFromRecord(k.Record)
	if err != nil {
		return nil, err
	}
	return codec.Amino.MarshalLengthPrefixed(info)
}

func extractPrivKeyFromLocal(rl *cosmoskeyring.Record_Local) (cryptotypes.PrivKey, error) {
	if rl.PrivKey == nil {
		return nil, cosmoskeyring.ErrPrivKeyNotAvailable
	}

	priv, ok := rl.PrivKey.GetCachedValue().(cryptotypes.PrivKey)
	if !ok {
		return nil, cosmoskeyring.ErrCastAny
	}

	return priv, nil
}

func privKeyFromRecord(record *cosmoskeyring.Record) (cryptotypes.PrivKey, error) {
	switch record.GetType() {
	case cosmoskeyring.TypeLocal:
		return extractPrivKeyFromLocal(record.GetLocal())
	}
	return nil, fmt.Errorf("unhandled Record type %q", record.GetType())
}

func privKeyFromInfo(info cosmoskeyring.LegacyInfo) (privKey cryptotypes.PrivKey, err error) {
	switch info.GetType() {
	case cosmoskeyring.TypeLocal:
		err = codec.Amino.Unmarshal([]byte(info.(legacyLocalInfo).GetPrivKeyArmor()), &privKey)
		return
	}
	return nil, fmt.Errorf("unhandled Info type %q", info.GetType())
}

// legacyInfoFromLegacyInfo turns a Record into a LegacyInfo.
func legacyInfoFromRecord(record *cosmoskeyring.Record) (cosmoskeyring.LegacyInfo, error) {
	switch record.GetType() {
	case cosmoskeyring.TypeLocal:
		pk, err := record.GetPubKey()
		if err != nil {
			return nil, err
		}
		privKey, err := extractPrivKeyFromLocal(record.GetLocal())
		if err != nil {
			return nil, err
		}
		privBz, err := codec.Amino.Marshal(privKey)
		if err != nil {
			return nil, err
		}
		return legacyLocalInfo{
			Name:         record.Name,
			PubKey:       pk,
			Algo:         hd.PubKeyType(pk.Type()),
			PrivKeyArmor: string(privBz),
		}, nil

	case cosmoskeyring.TypeLedger:
		pk, err := record.GetPubKey()
		if err != nil {
			return nil, err
		}
		return legacyLedgerInfo{
			Name:   record.Name,
			PubKey: pk,
			Algo:   hd.PubKeyType(pk.Type()),
			Path:   *record.GetLedger().Path,
		}, nil

	case cosmoskeyring.TypeMulti:
		panic("record type TypeMulti unhandled")

	case cosmoskeyring.TypeOffline:
		panic("record type TypeOffline unhandled")
	}
	panic(fmt.Sprintf("record type %s unhandled", record.GetType()))
}
