package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
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

	var (
		privkey = ed25519.GenPrivKeyFromSecret([]byte("secret"))
		pubkey  = privkey.PubKey()
	)
	record, err := cosmoskeyring.NewLocalRecord("local", privkey, pubkey)
	if err != nil {
		t.Fatalf("failed to create new local record: %v", err)
	}

	name := "proto.info"
	err = kr.AddProto(name, record)
	if err != nil {
		t.Fatalf("failed to add proto record: %v", err)
	}

	var buf bytes.Buffer
	err = printKeys(&buf, kr, "cosmos")
	assert.NoError(t, err)

	expected := []keyOutput{
		{
			Name:     name,
			Encoding: "proto",
			Address:  "cosmos182t3l5ptfgrlcg926xfk60936f3mjms0djnj6g",
			Type:     "local",
			PubKey:   `{"@type":"/cosmos.crypto.ed25519.PubKey","key":"XQNqhYzon4REkXYuuJ4r+9UKSgoNpljksmKLJbEXrgk="}`,
		},
	}
	var result []keyOutput
	err = yaml.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
