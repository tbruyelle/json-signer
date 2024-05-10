package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/tbruyelle/legacykey/keyring"
)

func main() {
	rootCmd := &ffcli.Command{
		ShortUsage: "legacykey <subcommand>",
		Subcommands: []*ffcli.Command{
			migrateKeysCmd(), signTxCmd(),
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

func signTxCmd() *ffcli.Command {
	fs := flag.NewFlagSet("sign-tx", flag.ContinueOnError)
	keyringDir := fs.String("keyring-dir", "", "Keyring directory")
	signer := fs.String("from", "", "Signer key name")
	chainID := fs.String("chain-id", "", "Chain identifier")
	account := fs.Uint64("account", 0, "Account number")
	sequence := fs.Uint64("sequence", 0, "Sequence number")
	bytesToSignFile := fs.String("bytes-to-sign-file", "", "Path to file that contains the bytes to sign in base64 std format")
	return &ffcli.Command{
		Name:       "sign-tx",
		ShortUsage: "legacykey sign-tx --from <key> --keyring-dir <dir> --chain-id <chainID> --sequence <sequence> --account <account> <tx.json>",
		ShortHelp:  "Sign transaction",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if err := fs.Parse(args); err != nil {
				return err
			}
			if fs.NArg() != 1 || fs.Lookup("keyring-dir") == nil ||
				fs.Lookup("from") == nil || fs.Lookup("sequence") == nil ||
				fs.Lookup("account") == nil || fs.Lookup("chain-id") == nil {
				return flag.ErrHelp
			}
			tx, err := readTxFile(fs.Arg(0))
			if err != nil {
				return err
			}
			kr, err := keyring.New(*keyringDir, nil)
			if err != nil {
				return err
			}
			var bytesToSign string
			if *bytesToSignFile != "" {
				bz, err := os.ReadFile(*bytesToSignFile)
				if err != nil {
					return err
				}
				bytesToSign = string(bz)
			}

			signedTx, err := signTx(tx, kr, *signer, *chainID, *account, *sequence, bytesToSign)
			if err != nil {
				return err
			}

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
		ShortUsage: "legacykey migrate-keys <keyring_path>",
		ShortHelp:  "Migrate keys from proto to amino",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			fs.Parse(args)
			if fs.NArg() != 1 {
				return flag.ErrHelp
			}
			return migrateKeys(fs.Arg(0))
		},
	}
}
