package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/99designs/keyring"
	"github.com/bgentry/speakeasy"
	"github.com/davecgh/go-spew/spew"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
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
	aminoCodec.RegisterInterface((*cosmoskeyring.LegacyInfo)(nil), nil)
	aminoCodec.RegisterConcrete(hd.BIP44Params{}, "crypto/keys/hd/BIP44Params", nil)
	aminoCodec.RegisterConcrete(legacyLocalInfo{}, "crypto/keys/localInfo", nil)
	aminoCodec.RegisterConcrete(legacyLedgerInfo{}, "crypto/keys/ledgerInfo", nil)
	aminoCodec.RegisterConcrete(legacyOfflineInfo{}, "crypto/keys/offlineInfo", nil)
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
			errProto := protocodec.Unmarshal(item.Data, &record)
			if errProto == nil {
				fmt.Printf("%q (proto encoded)-> %s\n", key, spew.Sdump(record))
				// Turn record to legacyInfo
				info, err := legacyInfoFromRecord(record)
				if err != nil {
					panic(err)
				}
				// amino marshal legacyInfo
				bz, err := aminoCodec.MarshalLengthPrefixed(info)
				if err != nil {
					panic(err)
				}
				addr, err := record.GetAddress()
				if err != nil {
					panic(err)
				}
				// record in new keyring
				aminoKeyringDir := filepath.Join(keyringDir, "amino")
				aminoKr, err := keyring.Open(keyring.Config{
					AllowedBackends: []keyring.BackendType{keyring.FileBackend},
					ServiceName:     "govgen",
					FileDir:         aminoKeyringDir,
					FilePasswordFunc: func(prompt string) (string, error) {
						return speakeasy.Ask(fmt.Sprintf("Enter password for amino keyring %q: ", aminoKeyringDir))
					},
				})
				if err := aminoKr.Set(keyring.Item{Key: key, Data: bz}); err != nil {
					panic(err)
				}
				item = keyring.Item{
					Key:  addrHexKeyAsString(addr),
					Data: []byte(key),
				}

				if err := aminoKr.Set(item); err != nil {
					panic(err)
				}
				fmt.Printf("%q re-encoded to amino keyring %q\n", key, aminoKeyringDir)
				continue
			}
			var info cosmoskeyring.LegacyInfo
			errAmino := aminoCodec.UnmarshalLengthPrefixed(item.Data, &info)
			if errAmino == nil {
				fmt.Printf("%q (amino encoded)-> %s\n", key, spew.Sdump(info))
				continue
			}
			fmt.Printf("%q cannot be decoded: errProto=%v, errAmino=%v\n", key, errProto, errAmino)
		}
	}
}

func addrHexKeyAsString(address sdk.Address) string {
	return fmt.Sprintf("%s.address", hex.EncodeToString(address.Bytes()))
}

// legacyInfoFromLegacyInfo turns a Record into a LegacyInfo.
func legacyInfoFromRecord(record cosmoskeyring.Record) (cosmoskeyring.LegacyInfo, error) {
	switch record.GetType() {
	case cosmoskeyring.TypeLocal:
		pk, err := record.GetPubKey()
		if err != nil {
			return nil, err
		}
		privBz, err := protocodec.Marshal(record.GetLocal().PrivKey)
		if err != nil {
			return nil, err
		}
		return legacyLocalInfo{
			Name:         record.Name,
			PubKey:       pk,
			PrivKeyArmor: string(privBz),
			Algo:         hd.PubKeyType(pk.Type()),
		}, nil

	case cosmoskeyring.TypeLedger:
		// TODO

	case cosmoskeyring.TypeMulti:
		panic("record type TypeMulti unhandled")

	case cosmoskeyring.TypeOffline:
		panic("record type TypeOffline unhandled")
	}
	panic(fmt.Sprintf("record type %s unhandled", record.GetType()))
}
