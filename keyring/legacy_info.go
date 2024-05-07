package keyring

// imported from cosmos-sdk/crypto/keyring because of private types *LocalInfo

import (
	"fmt"

	"github.com/tbruyelle/legacykey/codec"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ cosmoskeyring.LegacyInfo = &legacyLocalInfo{}
	_ cosmoskeyring.LegacyInfo = &legacyLedgerInfo{}
	_ cosmoskeyring.LegacyInfo = &legacyOfflineInfo{}
)

func init() {
	codec.Amino.RegisterInterface((*cosmoskeyring.LegacyInfo)(nil), nil)
	codec.Amino.RegisterConcrete(hd.BIP44Params{}, "crypto/keys/hd/BIP44Params", nil)
	codec.Amino.RegisterConcrete(legacyLocalInfo{}, "crypto/keys/localInfo", nil)
	codec.Amino.RegisterConcrete(legacyLedgerInfo{}, "crypto/keys/ledgerInfo", nil)
	codec.Amino.RegisterConcrete(legacyOfflineInfo{}, "crypto/keys/offlineInfo", nil)
}

// legacyLocalInfo is the public information about a locally stored key
// Note: Algo must be last field in struct for backwards amino compatibility
type legacyLocalInfo struct {
	Name         string             `json:"name"`
	PubKey       cryptotypes.PubKey `json:"pubkey"`
	PrivKeyArmor string             `json:"privkey.armor"`
	Algo         hd.PubKeyType      `json:"algo"`
}

// GetType implements Info interface
func (i legacyLocalInfo) GetType() cosmoskeyring.KeyType {
	return cosmoskeyring.TypeLocal
}

// GetType implements Info interface
func (i legacyLocalInfo) GetName() string {
	return i.Name
}

// GetType implements Info interface
func (i legacyLocalInfo) GetPubKey() cryptotypes.PubKey {
	return i.PubKey
}

// GetType implements Info interface
func (i legacyLocalInfo) GetAddress() sdk.AccAddress {
	return i.PubKey.Address().Bytes()
}

// GetPrivKeyArmor
func (i legacyLocalInfo) GetPrivKeyArmor() string {
	return i.PrivKeyArmor
}

// GetType implements Info interface
func (i legacyLocalInfo) GetAlgo() hd.PubKeyType {
	return i.Algo
}

// GetType implements Info interface
func (i legacyLocalInfo) GetPath() (*hd.BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
}

// legacyLedgerInfo is the public information about a Ledger key
// Note: Algo must be last field in struct for backwards amino compatibility
type legacyLedgerInfo struct {
	Name   string             `json:"name"`
	PubKey cryptotypes.PubKey `json:"pubkey"`
	Path   hd.BIP44Params     `json:"path"`
	Algo   hd.PubKeyType      `json:"algo"`
}

// GetType implements Info interface
func (i legacyLedgerInfo) GetType() cosmoskeyring.KeyType {
	return cosmoskeyring.TypeLedger
}

// GetName implements Info interface
func (i legacyLedgerInfo) GetName() string {
	return i.Name
}

// GetPubKey implements Info interface
func (i legacyLedgerInfo) GetPubKey() cryptotypes.PubKey {
	return i.PubKey
}

// GetAddress implements Info interface
func (i legacyLedgerInfo) GetAddress() sdk.AccAddress {
	return i.PubKey.Address().Bytes()
}

// GetPath implements Info interface
func (i legacyLedgerInfo) GetAlgo() hd.PubKeyType {
	return i.Algo
}

// GetPath implements Info interface
func (i legacyLedgerInfo) GetPath() (*hd.BIP44Params, error) {
	tmp := i.Path
	return &tmp, nil
}

// legacyOfflineInfo is the public information about an offline key
// Note: Algo must be last field in struct for backwards amino compatibility
type legacyOfflineInfo struct {
	Name   string             `json:"name"`
	PubKey cryptotypes.PubKey `json:"pubkey"`
	Algo   hd.PubKeyType      `json:"algo"`
}

// GetType implements Info interface
func (i legacyOfflineInfo) GetType() cosmoskeyring.KeyType {
	return cosmoskeyring.TypeOffline
}

// GetName implements Info interface
func (i legacyOfflineInfo) GetName() string {
	return i.Name
}

// GetPubKey implements Info interface
func (i legacyOfflineInfo) GetPubKey() cryptotypes.PubKey {
	return i.PubKey
}

// GetAlgo returns the signing algorithm for the key
func (i legacyOfflineInfo) GetAlgo() hd.PubKeyType {
	return i.Algo
}

// GetAddress implements Info interface
func (i legacyOfflineInfo) GetAddress() sdk.AccAddress {
	return i.PubKey.Address().Bytes()
}

// GetPath implements Info interface
func (i legacyOfflineInfo) GetPath() (*hd.BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
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

// LegacyInfoFromLegacyInfo turns a Record into a LegacyInfo.
func LegacyInfoFromRecord(record *cosmoskeyring.Record) (cosmoskeyring.LegacyInfo, error) {
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
