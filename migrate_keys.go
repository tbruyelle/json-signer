package main

import (
	"fmt"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/tbruyelle/legacykey/codec"
	"github.com/tbruyelle/legacykey/keyring"
)

func migrateKeys(keyringDir string) error {
	kr, err := keyring.New(keyringDir, "")
	if err != nil {
		return err
	}
	// new keyring for migrated keys
	aminoKeyringDir := filepath.Join(keyringDir, "amino")
	aminoKr, err := keyring.New(aminoKeyringDir,
		fmt.Sprintf("Enter password for amino keyring %q: ", aminoKeyringDir))
	if err != nil {
		return err
	}
	keys, err := kr.Keys()
	if err != nil {
		return err
	}
	for _, key := range keys {
		if key.IsAmino() {
			// this is a amino-encoded key  no migration just display
			fmt.Printf("%q (amino encoded)-> %s\n", key.Name, spew.Sdump(key.Info))
			continue
		}
		// this is a proto-encoded key let's migrate it back to amino
		fmt.Printf("%q (proto encoded)-> %s\n", key.Name, spew.Sdump(key.Record))
		// Turn record to legacyInfo
		info, err := codec.LegacyInfoFromRecord(key.Record)
		if err != nil {
			return err
		}
		// Register new amino key_name.info -> amino encoded LegacyInfo
		bz, err := codec.Amino.MarshalLengthPrefixed(info)
		if err != nil {
			return err
		}
		if err := aminoKr.Set(key.Name, bz); err != nil {
			return err
		}
		// TODO create keyring-dir/keyhash file
		fmt.Printf("%q re-encoded to amino keyring %q\n", key, aminoKeyringDir)
	}
	return nil
}
