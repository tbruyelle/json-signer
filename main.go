package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/99designs/keyring"
	"github.com/bgentry/speakeasy"
	"github.com/davecgh/go-spew/spew"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		switch {

		case strings.HasSuffix(key, ".address"):
			item, err := kr.Get(key)
			if err != nil {
				panic(err)
			}
			bz, err := hex.DecodeString(strings.TrimSuffix(key, ".address"))
			if err != nil {
				panic(err)
			}
			addr := sdk.AccAddress(bz)
			fmt.Printf("%s -> %s - %s\n", key, addr.String(), string(item.Data))

		case strings.HasSuffix(key, ".info"):
			item, err := kr.Get(key)
			if err != nil {
				panic(err)
			}
			var record cosmoskeyring.Record
			if err := protocodec.Unmarshal(item.Data, &record); err == nil {
				fmt.Printf("%s (proto encoded)-> %s\n", key, spew.Sdump(record))
				continue
			}
			var info cosmoskeyring.LegacyInfo
			if err := aminoCodec.UnmarshalLengthPrefixed(item.Data, &info); err == nil {
				fmt.Printf("%s (amino encoded)-> %s\n", key, spew.Sdump(info))
				continue
			}
			fmt.Printf("%s cannot be decoded\n", key)
		}
	}
}

// TODO reverse this convert function
/*
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
*/
