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
			migrateKeysCmd(),
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

func migrateKeysCmd() *ffcli.Command {
	fs := flag.NewFlagSet("migrate-keys", flag.ContinueOnError)
	return &ffcli.Command{
		Name:       "migrate-keys",
		ShortUsage: "legacytx migrate-keys <keyring_path>",
		ShortHelp:  "Migrate keys from proto to amino",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return flag.ErrHelp
			}
			fs.Parse(args)
			return migrateKeys(fs.Arg(0))
		},
	}
}
