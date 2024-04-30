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
	registry   = codectypes.NewInterfaceRegistry()
	protocodec *codec.ProtoCodec
)

func init() {
	cryptocodec.RegisterInterfaces(registry)
	protocodec = codec.NewProtoCodec(registry)
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
		var k cosmoskeyring.Record
		if err := protocodec.Unmarshal(item.Data, &k); err != nil {
			panic(err)
		}
		fmt.Println("RECORD", k)
	}
}
