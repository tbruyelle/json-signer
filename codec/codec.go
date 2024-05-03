package codec

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
)

var (
	Proto *codec.ProtoCodec
	Amino *codec.LegacyAmino
)

func init() {
	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	Proto = codec.NewProtoCodec(registry)

	Amino = codec.NewLegacyAmino()
	cryptocodec.RegisterCrypto(Amino)
	Amino.RegisterInterface((*cosmoskeyring.LegacyInfo)(nil), nil)
	Amino.RegisterConcrete(hd.BIP44Params{}, "crypto/keys/hd/BIP44Params", nil)
	Amino.RegisterConcrete(legacyLocalInfo{}, "crypto/keys/localInfo", nil)
	Amino.RegisterConcrete(legacyLedgerInfo{}, "crypto/keys/ledgerInfo", nil)
	Amino.RegisterConcrete(legacyOfflineInfo{}, "crypto/keys/offlineInfo", nil)
}
