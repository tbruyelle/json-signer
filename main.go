package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/tbruyelle/keyring-compat"
)

func main() {
	rootCmd := &ffcli.Command{
		ShortUsage: "json-signer <subcommand>",
		Subcommands: []*ffcli.Command{
			listKeysCmd(), migrateKeysCmd(), signTxCmd(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
	err := rootCmd.ParseAndRun(context.Background(), os.Args[1:])
	if err != nil && err != flag.ErrHelp {
		log.Fatal(err)
	}
}

func listKeysCmd() *ffcli.Command {
	fs := flag.NewFlagSet("list-keys", flag.ContinueOnError)
	keyringDir := fs.String("keyring-dir", "", "Keyring directory")
	keyringBackend := fs.String("keyring-backend", "", "Keyring backend, which can be one of 'keychain' (macos), 'pass', 'kwallet' (linux), or 'file'")
	return &ffcli.Command{
		Name:       "list-keys",
		ShortUsage: "json-signer list-keys --keyring-backend <keychain|pass|kwallet|file> --keyring-dir <dir>",
		ShortHelp:  "List keys from keyring",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if err := fs.Parse(args); err != nil {
				return err
			}
			kr, err := keyring.New(keyring.BackendType(*keyringBackend), *keyringDir, nil)
			if err != nil {
				return err
			}
			keys, err := kr.Keys()
			if err != nil {
				return err
			}
			if len(keys) == 0 {
				fmt.Println("No keys found in keyring")
				return nil
			}
			for i, key := range keys {
				encoding := "proto"
				if key.IsAminoEncoded() {
					encoding = "amino"
				}
				fmt.Printf("%d) %-20s encoding=%s\ttype=%s\tpubkey=%s\n", i+1, key.Name(), encoding, key.Type(), "TODO")
			}
			return nil
		},
	}
}

func signTxCmd() *ffcli.Command {
	fs := flag.NewFlagSet("sign-tx", flag.ContinueOnError)
	keyringDir := fs.String("keyring-dir", "", "Keyring directory")
	keyringBackend := fs.String("keyring-backend", "", "Keyring backend, which can be one of 'keychain' (macos), 'pass', 'kwallet' (linux), or 'file'")
	signer := fs.String("from", "", "Signer key name")
	chainID := fs.String("chain-id", "", "Chain identifier")
	account := fs.String("account", "", "Account number")
	sequence := fs.String("sequence", "", "Sequence number")
	return &ffcli.Command{
		Name:       "sign-tx",
		ShortUsage: "json-signer sign-tx --from <key> --keyring-dir <dir> --chain-id <chainID> --sequence <sequence> --account <account> <tx.json>",
		ShortHelp:  "Sign transaction",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if err := fs.Parse(args); err != nil {
				return err
			}
			if fs.NArg() != 1 ||
				fs.Lookup("keyring-dir") == nil || // FIXME not mandatory for backend than file
				fs.Lookup("keyring-backend") == nil ||
				fs.Lookup("from") == nil || fs.Lookup("sequence") == nil ||
				fs.Lookup("account") == nil || fs.Lookup("chain-id") == nil {
				return flag.ErrHelp
			}
			tx, err := readTxFile(fs.Arg(0))
			if err != nil {
				return err
			}
			kr, err := keyring.New(keyring.BackendType(*keyringBackend), *keyringDir, nil)
			if err != nil {
				return err
			}

			// TODO test with ledger
			signedTx, bytesToSign, err := signTx(tx, kr, *signer, *chainID, *account, *sequence)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Bytes to sign:\n%s\n", string(bytesToSign))

			// Output tx
			bz, err := json.MarshalIndent(signedTx, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(bz))
			return nil
		},
	}
}

func readTxFile(txFile string) (Tx, error) {
	f, err := os.Open(txFile)
	if err != nil {
		return Tx{}, err
	}
	defer f.Close()
	var tx Tx
	if err := json.NewDecoder(f).Decode(&tx); err != nil {
		return Tx{}, fmt.Errorf("JSON decode %s: %v", txFile, err)
	}
	return tx, nil
}

func migrateKeysCmd() *ffcli.Command {
	fs := flag.NewFlagSet("migrate-keys", flag.ContinueOnError)
	return &ffcli.Command{
		Name:       "migrate-keys",
		ShortUsage: "json-signer migrate-keys <keyring_path>",
		ShortHelp:  "Migrate keys from proto to amino",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			fs.Parse(args)
			if fs.NArg() != 1 {
				return flag.ErrHelp
			}
			kr, err := keyring.New(keyring.BackendType("file"), fs.Arg(0), nil)
			if err != nil {
				return err
			}
			return kr.MigrateProtoKeysToAmino()
		},
	}
}
