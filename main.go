package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/tbruyelle/keyring-compat"
)

var (
	keyringDir,
	keyringBackend,
	signer,
	chainID,
	account,
	sequence *string

	sigOnly *bool
)

func main() {
	rootCmd := &ffcli.Command{
		ShortUsage: "json-signer <subcommand>",
		Subcommands: []*ffcli.Command{
			listKeysCmd(), migrateKeysCmd(), signTxCmd(), batchSignTxCmd(),
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
	keyringDir := fs.String("keyring-dir", "", "Keyring directory (mandatory with -keyring-backend=file)")
	keyringBackend := fs.String("keyring-backend", "", "Keyring backend, which can be one of 'keychain' (macos), 'pass', 'kwallet' (linux), or 'file'")
	prefix := fs.String("prefix", "cosmos", "Bech32 address prefix")
	return &ffcli.Command{
		Name:       "list-keys",
		ShortUsage: "json-signer list-keys -keyring-backend=<keychain|pass|kwallet|file> -prefix=<prefix>",
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
			return printKeys(os.Stdout, kr, *prefix)
		},
	}
}

func signTxCmd() *ffcli.Command {
	fs := flag.NewFlagSet("sign-tx", flag.ContinueOnError)
	setCommonFlags(fs)

	return &ffcli.Command{
		Name:       "sign-tx",
		ShortUsage: "json-signer sign-tx -from=<key> -keyring-backend=<keychain|pass|kwallet|file> -chain-id=<chainID> -sequence=<sequence> -account=<account-number> <tx.json>",
		ShortHelp:  "Sign transaction",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if err := parseAndCheckCommonFlags(fs, args); err != nil {
				return err
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
			fmt.Fprintf(os.Stderr, "Bytes to sign: %s\n", string(bytesToSign))

			var output any = signedTx
			if *sigOnly {
				// Output signature only
				output = signedTx.GetSignaturesData()
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

func batchSignTxCmd() *ffcli.Command {
	fs := flag.NewFlagSet("sign-tx-batch", flag.ContinueOnError)
	setCommonFlags(fs)

	return &ffcli.Command{
		Name:       "sign-tx-batch",
		ShortUsage: "json-signer sign-tx-batch -from=<key> -keyring-backend=<keychain|pass|kwallet|file> -chain-id=<chainID> -sequence=<sequence> -account=<account-number> [file] ([file2]...)",
		ShortHelp:  "Sign batch transaction(s)",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if err := parseAndCheckCommonFlags(fs, args); err != nil {
				return err
			}

			kr, err := keyring.New(keyring.BackendType(*keyringBackend), *keyringDir, nil)
			if err != nil {
				return err
			}

			signedTxs, err := batchSignTxs(fs.Args(), kr, *signer, *chainID, *account, *sequence)
			if err != nil {
				return err
			}

			switch {
			case *sigOnly:
				{
					// Output signature only
					for _, signedTx := range signedTxs {
						tmpOutput, err := json.Marshal(signedTx.GetSignaturesData())
						if err != nil {
							return err
						}

						fmt.Println(string(tmpOutput))
					}

					return nil
				}
			default:
				{
					for _, signedTx := range signedTxs {
						bz, err := json.Marshal(signedTx)
						if err != nil {
							return err
						}

						fmt.Println(string(bz))
					}

					return nil
				}
			}
		},
	}
}

func batchSignTxs(files []string, kr keyring.Keyring, signer, chainID, account, sequence string) ([]Tx, error) {
	var (
		signedTxs, unsignedTxs []Tx
		unsignedTx, signedTx   Tx
		i                      int
	)

	seq, err := strconv.ParseInt(sequence, 10, 64)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		unsignedTxs, err = readTxs(file)
		if err != nil {
			return nil, err
		}

		// Sign each tx
		for _, unsignedTx = range unsignedTxs {
			sequence = strconv.FormatInt(seq+int64(i), 10)
			i++

			signedTx, _, err = signTx(unsignedTx, kr, signer, chainID, account, sequence)
			if err != nil {
				return nil, err
			}

			signedTxs = append(signedTxs, signedTx)
		}
	}

	return signedTxs, nil
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

func readTxs(fileLoc string) ([]Tx, error) {
	f, err := os.Open(fileLoc)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Create a new scanner for the file
	scanner := bufio.NewScanner(f)

	// Read and append each line of the file
	var txs []Tx
	for scanner.Scan() {
		line := scanner.Text()
		var tx Tx
		if err := json.Unmarshal([]byte(line), &tx); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}

	// Check for errors during the scan
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return txs, nil
}

func migrateKeysCmd() *ffcli.Command {
	fs := flag.NewFlagSet("migrate-keys", flag.ContinueOnError)
	return &ffcli.Command{
		Name:       "migrate-keys",
		ShortUsage: "json-signer migrate-keys <keyring_path>",
		ShortHelp:  "Migrate keys from proto to amino (only file backend is supported for now)",
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

func setCommonFlags(f *flag.FlagSet) {
	if f == nil {
		return
	}

	keyringDir = f.String("keyring-dir", "", "Keyring directory (mandatory with -keyring-backend=file)")
	keyringBackend = f.String("keyring-backend", "", "Keyring backend, which can be one of 'keychain' (macos), 'pass', 'kwallet' (linux), or 'file'")
	signer = f.String("from", "", "Signer key name")
	chainID = f.String("chain-id", "", "Chain identifier")
	account = f.String("account", "", "Account number")
	sequence = f.String("sequence", "", "Sequence number")
	sigOnly = f.Bool("signature-only", false, "Outputs only the signature data")
}

func parseAndCheckCommonFlags(f *flag.FlagSet, args []string) error {
	if f == nil {
		return fmt.Errorf("flag.FlagSet is nil")

	}

	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 1 || *keyringBackend == "" || *signer == "" ||
		*sequence == "" || *account == "" || *chainID == "" {
		return flag.ErrHelp
	}
	if *keyringBackend == "file" && *keyringDir == "" {
		return fmt.Errorf("-keyring-backend=file requires -keyring-dir flag")
	}

	return nil
}
