package main

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/99designs/keyring"
	"github.com/bgentry/speakeasy"
	"github.com/davecgh/go-spew/spew"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func migrateKeys(keyringDir string) error {
	kr, err := keyring.Open(keyring.Config{
		AllowedBackends: []keyring.BackendType{keyring.FileBackend},
		FileDir:         keyringDir,
		FilePasswordFunc: func(prompt string) (string, error) {
			return speakeasy.Ask(prompt + ": ")
		},
	})
	if err != nil {
		return err
	}
	// new keyring for migrated keys
	aminoKeyringDir := filepath.Join(keyringDir, "amino")
	aminoKr, err := keyring.Open(keyring.Config{
		AllowedBackends: []keyring.BackendType{keyring.FileBackend},
		FileDir:         aminoKeyringDir,
		FilePasswordFunc: func(prompt string) (string, error) {
			return speakeasy.Ask(fmt.Sprintf("Enter password for amino keyring %q: ", aminoKeyringDir))
		},
	})
	if err != nil {
		return err
	}
	keys, err := kr.Keys()
	if err != nil {
		return err
	}
	for _, key := range keys {
		switch {

		case strings.HasSuffix(key, ".address"):
			item, err := kr.Get(key)
			if err != nil {
				return err
			}
			bz, err := hex.DecodeString(strings.TrimSuffix(key, ".address"))
			if err != nil {
				return err
			}
			addr := sdk.AccAddress(bz)
			fmt.Printf("%s -> %s - %s\n", key, addr.String(), string(item.Data))

		case strings.HasSuffix(key, ".info"):
			item, err := kr.Get(key)
			if err != nil {
				return err
			}
			var record cosmoskeyring.Record
			errProto := protocodec.Unmarshal(item.Data, &record)
			if errProto == nil {
				fmt.Printf("%q (proto encoded)-> %s\n", key, spew.Sdump(record))
				// Turn record to legacyInfo
				info, err := legacyInfoFromRecord(record)
				if err != nil {
					return err
				}
				// amino marshal legacyInfo
				bz, err := aminoCodec.MarshalLengthPrefixed(info)
				if err != nil {
					return err
				}
				addr, err := record.GetAddress()
				if err != nil {
					return err
				}
				if err := aminoKr.Set(keyring.Item{Key: key, Data: bz}); err != nil {
					return err
				}
				item = keyring.Item{
					Key:  addrHexKeyAsString(addr),
					Data: []byte(key),
				}

				if err := aminoKr.Set(item); err != nil {
					return err
				}
				// TODO create keyring-dir/keyhash file
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
	return nil
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
		privKey, err := extractPrivKeyFromLocal(record.GetLocal())
		if err != nil {
			return nil, err
		}
		privBz, err := aminoCodec.Marshal(privKey)
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
