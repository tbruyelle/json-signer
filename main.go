package main

import (
	"fmt"
	"os"

	"github.com/99designs/keyring"
)

func main() {
	keyringDir := os.Args[1]
	// test keyring
	kr, err := keyring.Open(keyring.Config{
		AllowedBackends: []keyring.BackendType{keyring.FileBackend},
		ServiceName:     "govgen",
		FileDir:         keyringDir,
		FilePasswordFunc: func(_ string) (string, error) {
			return "test", nil
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(kr.Keys())

	{
		// os keyring - not working for now
		kr, err := keyring.Open(keyring.Config{
			ServiceName:              "govgen",
			FileDir:                  keyringDir,
			KeychainTrustApplication: true,
			FilePasswordFunc: func(_ string) (string, error) {
				return "test", nil
			},
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(kr.Keys())
	}
}
