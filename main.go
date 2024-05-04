package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
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
	return &ffcli.Command{
		Name:       "sign-tx",
		ShortUsage: "legacykey sign-tx <tx.json> -from <key> -keyring-dir <dir> -chain-id <chainID> -sequence <sequence> -account <account>",
		ShortHelp:  "Sign transaction",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			fs.Parse(args)
			if fs.NArg() != 1 || flag.Lookup("keyring-dir") == nil ||
				flag.Lookup("from") == nil || flag.Lookup("sequence") == nil ||
				flag.Lookup("account") == nil || flag.Lookup("chain-id") == nil {
				return flag.ErrHelp
			}
			txFile := fs.Arg(0)
			return signTx(txFile, *keyringDir, *signer, *chainID, *account, *sequence)
		},
	}
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
