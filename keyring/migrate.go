package keyring

import (
	"fmt"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
)

func (kr Keyring) MigrateProtoKeysToAmino() error {
	// new keyring for migrated keys
	aminoKeyringDir := filepath.Join(kr.dir, "amino")
	aminoKr, err := New(aminoKeyringDir, nil)
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
		info, err := key.RecordToInfo()
		if err != nil {
			return err
		}
		// Register new amino key_name.info -> amino encoded LegacyInfo
		if err := aminoKr.AddAmino(key.Name, info); err != nil {
			return err
		}
		// TODO create keyring-dir/keyhash file
		fmt.Printf("%q re-encoded to amino keyring %q\n", key, aminoKeyringDir)
	}
	return nil
}
