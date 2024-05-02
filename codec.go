package main

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
)

var (
	protocodec *codec.ProtoCodec
	aminoCodec *codec.LegacyAmino
)

func init() {
	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	protocodec = codec.NewProtoCodec(registry)

	aminoCodec = codec.NewLegacyAmino()
	cryptocodec.RegisterCrypto(aminoCodec)
	aminoCodec.RegisterInterface((*cosmoskeyring.LegacyInfo)(nil), nil)
	aminoCodec.RegisterConcrete(hd.BIP44Params{}, "crypto/keys/hd/BIP44Params", nil)
	aminoCodec.RegisterConcrete(legacyLocalInfo{}, "crypto/keys/localInfo", nil)
	aminoCodec.RegisterConcrete(legacyLedgerInfo{}, "crypto/keys/ledgerInfo", nil)
	aminoCodec.RegisterConcrete(legacyOfflineInfo{}, "crypto/keys/offlineInfo", nil)
}
