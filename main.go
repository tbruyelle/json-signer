package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
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
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func listKeysCmd() *ffcli.Command {
	fs := flag.NewFlagSet("list-keys", flag.ContinueOnError)
	keyringDir := fs.String("keyring-dir", "", "Keyring directory")
	keyringBackend := fs.String("keyring-backend", "", "Keyring backend, which can be one of 'keychain' (macos), 'pass', 'kwallet' (linux), or 'file'")
	prefix := fs.String("prefix", "cosmos", "Bech32 address prefix")
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
			return PrintKeys(os.Stdout, kr, *prefix)
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
	sigOnly := fs.Bool("signature-only", false, "Outputs only the signature data")
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

			signedTx, bytesToSign, err := signTx(tx, kr, *signer, *chainID, *account, *sequence)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Bytes to sign:\n%s\n", string(bytesToSign))

			var output any = signedTx
			if *sigOnly {
				// Output signature only
				output, err = signedTx.GetSignaturesData()
				if err != nil {
					return err
				}
			}
			bz, err := json.Marshal(output)
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
