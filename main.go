package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/tbruyelle/keyring-compat"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func main() {
	rootCmd := &ffcli.Command{
		ShortUsage: "json-signer <subcommand>",
		Subcommands: []*ffcli.Command{
			listKeysCmd(), importKeyHexCmd(), migrateKeysCmd(), signTxCmd(), batchSignTxCmd(),
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

func importKeyHexCmd() *ffcli.Command {
	fs := flag.NewFlagSet("import-key-hex", flag.ContinueOnError)
	keyringDir := fs.String("keyring-dir", "", "Keyring directory (mandatory with -keyring-backend=file)")
	keyringBackend := fs.String("keyring-backend", "", "Keyring backend, which can be one of 'keychain' (macos), 'pass', 'kwallet' (linux), or 'file'")
	keyringEncoding := fs.String("keyring-encoding", "proto", "Keyring encoding, must be one of 'amino' or 'proto'")
	return &ffcli.Command{
		Name:       "import-key-hex",
		ShortUsage: "json-signer import-key-hex -keyring-backend=<keychain|pass|kwallet|file> -keyring-dir=<path> -keyring-encoding=<amino|proto> <key-name> <key-hex>",
		ShortHelp:  "Import private key in hex format",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if err := fs.Parse(args); err != nil {
				return err
			}
			if fs.NArg() != 2 {
				return flag.ErrHelp
			}
			kr, err := keyring.New(keyring.BackendType(*keyringBackend), *keyringDir, nil)
			if err != nil {
				return err
			}
			var (
				name   = fs.Arg(0)
				keyHex = fs.Arg(1)
			)
			// Decode the private key hex string
			privateKeyBz, err := hex.DecodeString(keyHex)
			if err != nil {
				return fmt.Errorf("failed to decode private key: %v", err)
			}
			// Generate types.PrivKey from bytes
			privKey := hd.Secp256k1.Generate()(privateKeyBz)
			// Create proto record from privKey
			record, err := cosmoskeyring.NewLocalRecord(name, privKey, privKey.PubKey())
			if err != nil {
				return fmt.Errorf("error NewLocalRecord: %v", err)
			}

			switch *keyringEncoding {
			case "amino":
				// Derive amino info from proto record
				info, err := keyring.LegacyInfoFromRecord(record)
				if err != nil {
					return fmt.Errorf("error LegacyInfoFromRecord: %v", err)
				}
				// record key with amino encoding
				err = kr.AddAmino(name, info)
				if err != nil {
					return fmt.Errorf("error AddAmino: %v", err)
				}
			case "proto":
				// record key with proto encoding
				err = kr.AddProto(name, record)
				if err != nil {
					return fmt.Errorf("error AddProto: %v", err)
				}
			default:
				return fmt.Errorf("unsuported encoding %s: must be one of amino or proto", *keyringEncoding)
			}
			return nil
		},
	}
}

func signTxCmd() *ffcli.Command {
	fs := flag.NewFlagSet("sign-tx", flag.ContinueOnError)
	keyringBackend, keyringDir, signer, chainID, account, sequence, sigOnly := setCommonFlags(fs)

	return &ffcli.Command{
		Name:       "sign-tx",
		ShortUsage: "json-signer sign-tx -from=<key> -keyring-backend=<keychain|pass|kwallet|file> -chain-id=<chainID> -sequence=<sequence> -account=<account-number> <tx.json>",
		ShortHelp:  "Sign transaction",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if err := parseAndCheckCommonFlags(fs, args, keyringDir, keyringBackend, signer, sequence, account, chainID, true); err != nil {
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

			switch {
			case *sigOnly:
				{
					// Output signature only
					return print(signedTx.GetSignaturesData())
				}
			default:
				{
					return print(signedTx)
				}
			}
		},
	}
}

func batchSignTxCmd() *ffcli.Command {
	fs := flag.NewFlagSet("sign-tx-batch", flag.ContinueOnError)
	keyringBackend, keyringDir, signer, chainID, account, sequence, sigOnly := setCommonFlags(fs)

	return &ffcli.Command{
		Name:       "sign-tx-batch",
		ShortUsage: "json-signer sign-tx-batch -from=<key> -keyring-backend=<keychain|pass|kwallet|file> -chain-id=<chainID> -sequence=<sequence> -account=<account-number> [file] ([file2]...)",
		ShortHelp:  "Sign batch transaction(s)",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if err := parseAndCheckCommonFlags(fs, args, keyringDir, keyringBackend, signer, sequence, account, chainID, false); err != nil {
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
						if err = print(signedTx.GetSignaturesData()); err != nil {
							return err
						}
					}

					return nil
				}
			default:
				{
					for _, signedTx := range signedTxs {
						if err = print(signedTx); err != nil {
							return err
						}
					}

					return nil
				}
			}
		},
	}
}

func print(x any) error {
	bz, err := json.Marshal(x)
	if err != nil {
		return err
	}
	fmt.Println(string(bz))
	return nil
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

func setCommonFlags(f *flag.FlagSet) (*string, *string, *string, *string, *string, *string, *bool) {
	if f == nil {
		panic("flag.FlagSet is nil")
	}

	keyringDir := f.String("keyring-dir", "", "Keyring directory (mandatory with -keyring-backend=file)")
	keyringBackend := f.String("keyring-backend", "", "Keyring backend, which can be one of 'keychain' (macos), 'pass', 'kwallet' (linux), or 'file'")
	signer := f.String("from", "", "Signer key name")
	chainID := f.String("chain-id", "", "Chain identifier")
	account := f.String("account", "", "Account number")
	sequence := f.String("sequence", "", "Sequence number")
	sigOnly := f.Bool("signature-only", false, "Outputs only the signature data")

	return keyringBackend, keyringDir, signer, chainID, account, sequence, sigOnly
}

func parseAndCheckCommonFlags(f *flag.FlagSet, args []string, keyringDir, keyringBackend, signer, sequence, account, chainID *string, checkNArg bool) error {
	if f == nil {
		return fmt.Errorf("flag.FlagSet is nil")
	}

	if err := f.Parse(args); err != nil {
		return err
	}
	if checkNArg {
		if f.NArg() != 1 {
			return flag.ErrHelp
		}
	}
	if *keyringBackend == "" || *signer == "" ||
		*sequence == "" || *account == "" || *chainID == "" {
		return flag.ErrHelp
	}
	if *keyringBackend == "file" && *keyringDir == "" {
		return fmt.Errorf("-keyring-backend=file requires -keyring-dir flag")
	}

	return nil
}
