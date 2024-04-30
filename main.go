package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/99designs/keyring"
	"github.com/bgentry/speakeasy"
)

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
	fmt.Println("KEYS", keys)
	for _, key := range keys {
		if !strings.HasSuffix(key, ".info") {
			continue
		}
		item, err := kr.Get(key)
		if err != nil {
			panic(err)
		}
		fmt.Println("KEY", key, item)
	}
}
