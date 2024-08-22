package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/stretchr/testify/assert"
	keyring "github.com/tbruyelle/keyring-compat"
	"gopkg.in/yaml.v2"
)

func TestListKeys(t *testing.T) {
	keyringDir, err := os.MkdirTemp("", fmt.Sprintf("tmpKeyringDir-%s", time.Now().Format("20060102-150405")))
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}

	defer func() {
		if err := os.RemoveAll(keyringDir); err != nil {
			t.Fatalf("failed to remove temp directory (%s): %v", keyringDir, err)
		}
	}()

	kr, err := keyring.New(
		keyring.BackendType("file"),
		keyringDir,
		func(_ string) (string, error) { return "", nil },
	)
	if err != nil {
		t.Fatalf("failed to create keyring: %v", err)
	}

	cdc := testutilmod.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}, gov.AppModuleBasic{})
	cosmosKr := cosmoskeyring.NewInMemory(cdc.Codec)

	name := "newAccount"
	valAcc, _, err := cosmosKr.NewMnemonic(name, cosmoskeyring.English, sdk.FullFundraiserPath, cosmoskeyring.DefaultBIP39Passphrase, hd.Secp256k1)
	if err != nil {
		t.Fatalf("failed to create new account: %v", err)
	}
	if valAcc == nil {
		t.Fatal("failed to create new account: account is nil")
	}

	addr, err := valAcc.GetAddress()
	if err != nil {
		t.Fatalf("failed to get address: %v", err)
	}
	t.Logf("address: %s", addr.String())

	err = kr.AddProto(name, valAcc)
	if err != nil {
		t.Fatalf("failed to add key to keyring: %v", err)
	}
	keys, err := kr.Keys()
	if err != nil {
		t.Fatalf("failed to get keys: %v", err)
	}
	t.Logf("keys: %v", keys)

	var buf bytes.Buffer
	err = printKeys(&buf, kr, "cosmos")
	assert.NoError(t, err)

	expected := []keyOutput{
		{
			Name:     name,
			Encoding: "proto",
			Address:  addr.String(),
			Type:     "local",
			PubKey:   `{"type":"tendermint/PubKeySecp256k1","value":"A+"}`,
		},
	}
	var result []keyOutput
	err = yaml.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
