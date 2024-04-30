package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/99designs/keyring"
	"github.com/bgentry/speakeasy"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
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
	cosmoskeyring.RegisterLegacyAminoCodec(aminoCodec)
}

func main() {
	keyringDir := os.Args[1]
	kr, err := keyring.Open(keyring.Config{
		AllowedBackends: []keyring.BackendType{keyring.FileBackend},
		ServiceName:     "govgen",
		FileDir:         keyringDir,
		FilePasswordFunc: func(prompt string) (string, error) {
			return speakeasy.Ask(prompt + ": ")
		},
	})
	if err != nil {
		panic(err)
	}
	keys, err := kr.Keys()
	if err != nil {
		panic(err)
	}
	for _, key := range keys {
		if !strings.HasSuffix(key, ".info") {
			continue
		}
		item, err := kr.Get(key)
		if err != nil {
			panic(err)
		}
		fmt.Println("KEY", key, item)
		var record cosmoskeyring.Record
		if err := protocodec.Unmarshal(item.Data, &record); err == nil {
			fmt.Println("PROTO ENCODED KEY", record)
		} else {
			var info cosmoskeyring.LegacyInfo
			if err := aminoCodec.UnmarshalLengthPrefixed(item.Data, &info); err != nil {
				panic(err)
			}
			fmt.Println("AMINO ENCODED KEY", info)
		}
	}
}

// TODO reverse this convert function
func convertFromLegacyInfo(info cosmoskeyring.LegacyInfo) (*cosmoskeyring.Record, error) {
	name := info.GetName()
	pk := info.GetPubKey()

	switch info.GetType() {
	case cosmoskeyring.TypeLocal:
		priv, err := privKeyFromLegacyInfo(info)
		if err != nil {
			return nil, err
		}

		return cosmoskeyring.NewLocalRecord(name, priv, pk)
	case cosmoskeyring.TypeOffline:
		return cosmoskeyring.NewOfflineRecord(name, pk)
	case cosmoskeyring.TypeMulti:
		return cosmoskeyring.NewMultiRecord(name, pk)
	case cosmoskeyring.TypeLedger:
		path, err := info.GetPath()
		if err != nil {
			return nil, err
		}

		return cosmoskeyring.NewLedgerRecord(name, pk, path)
	default:
		return nil, cosmoskeyring.ErrUnknownLegacyType

	}
}

// privKeyFromLegacyInfo exports a private key from LegacyInfo
func privKeyFromLegacyInfo(info cosmoskeyring.LegacyInfo) (cryptotypes.PrivKey, error) {
	switch linfo := info.(type) {
	case legacyLocalInfo:
		if linfo.PrivKeyArmor == "" {
			return nil, fmt.Errorf("private key not available")
		}
		priv, err := legacy.PrivKeyFromBytes([]byte(linfo.PrivKeyArmor))
		if err != nil {
			return nil, err
		}

		return priv, nil
	// case legacyLedgerInfo, legacyOfflineInfo, LegacyMultiInfo:
	default:
		return nil, fmt.Errorf("only works on local private keys, provided %s", linfo)
	}
}
