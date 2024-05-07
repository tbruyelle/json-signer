package keyring

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/99designs/keyring"
	"github.com/bgentry/speakeasy"
	"github.com/tbruyelle/legacykey/codec"

	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
)

type Keyring struct {
	k keyring.Keyring
}

func New(keyringDir, alternatePrompt string) (Keyring, error) {
	k, err := keyring.Open(keyring.Config{
		AllowedBackends: []keyring.BackendType{keyring.FileBackend},
		FileDir:         keyringDir,
		FilePasswordFunc: func(prompt string) (string, error) {
			if alternatePrompt != "" {
				prompt = alternatePrompt
			}
			return speakeasy.FAsk(os.Stderr, prompt)
		},
	})
	if err != nil {
		return Keyring{}, err
	}
	return Keyring{k: k}, nil
}

func (k Keyring) Keys() ([]Key, error) {
	var keys []Key
	names, err := k.k.Keys()
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		if !strings.HasSuffix(name, ".info") {
			continue
		}
		key, err := k.Get(name)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (k Keyring) GetByAddress(addr string) (Key, error) {
	item, err := k.k.Get(hex.EncodeToString([]byte(addr)) + ".address")
	if err != nil {
		return Key{}, err
	}
	return k.Get(string(item.Data) + ".info")
}

func (k Keyring) Get(name string) (Key, error) {
	item, err := k.k.Get(name)
	if err != nil {
		return Key{}, err
	}

	// try proto decode
	var record cosmoskeyring.Record
	errProto := codec.Proto.Unmarshal(item.Data, &record)
	if errProto == nil {
		return Key{Name: name, Record: &record}, nil
	}
	// try amino decode
	var info cosmoskeyring.LegacyInfo
	errAmino := codec.Amino.UnmarshalLengthPrefixed(item.Data, &info)
	if errAmino == nil {
		return Key{Name: name, Info: info}, nil
	}
	return Key{}, fmt.Errorf("cannot decode key %s: decodeProto=%v decodeAmino=%v", name, errProto, errAmino)
}

func (k Keyring) Set(name string, bz []byte) error {
	return k.k.Set(keyring.Item{Key: name, Data: bz})
}
