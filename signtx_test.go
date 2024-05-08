package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tbruyelle/legacykey/keyring"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func TestGetBytesToSign(t *testing.T) {
	tests := []struct {
		name     string
		tx       Tx
		expected string
	}{
		{
			name: "/cosmos.bank.v1beta1.MsgSend",
			tx: Tx{
				Body: Body{
					Messages: []map[string]any{{
						"@type": "/cosmos.bank.v1beta1.MsgSend",
						"amount": []Coin{{
							Amount: "1000",
							Denom:  "token",
						}},
						"from_address": "cosmos1shzsqakdakzwhvy05cvjlt9acwf3hfjksy0ht5",
						"to_address":   "cosmos18lu8k4n7nmqhz2z3y9a5y39fzgapchfq6mvaeg",
					}},
					Memo:          "a memo",
					TimeoutHeight: "42",
				},
				AuthInfo: AuthInfo{
					Fee: Fee{
						Amount: []Coin{{
							Amount: "10",
							Denom:  "token",
						}},
						GasLimit: "200000",
					},
				},
			},
			expected: `{"account_number":"1","chain_id":"chainid-1","fee":{"amount":[{"amount":"10","denom":"token"}],"gas":"200000"},"memo":"a memo","msgs":[{"type":"cosmos-sdk/MsgSend","value":{"amount":[{"amount":"1000","denom":"token"}],"from_address":"cosmos1shzsqakdakzwhvy05cvjlt9acwf3hfjksy0ht5","to_address":"cosmos18lu8k4n7nmqhz2z3y9a5y39fzgapchfq6mvaeg"}}],"sequence":"2","timeout_height":"42"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			bytesToSign, err := getBytesToSign(tt.tx, "chainid-1", 1, 2)

			require.NoError(err)
			require.JSONEq(tt.expected, string(bytesToSign))
		})
	}
}

func TestSignTx(t *testing.T) {
	kr, err := keyring.New(t.TempDir(),
		func(_ string) (string, error) { return "test", nil })
	require.NoError(t, err)
	// Generate a local private key
	var (
		privkey = ed25519.GenPrivKey()
		pubkey  = privkey.PubKey()
	)
	record, err := cosmoskeyring.NewLocalRecord("local", privkey, pubkey)
	require.NoError(t, err)
	kr.AddProto("local.info", record)
	bankSendTx := Tx{
		Body: Body{
			Messages: []map[string]any{{
				"@type": "/cosmos.bank.v1beta1.MsgSend",
				"amount": []Coin{{
					Amount: "1000",
					Denom:  "token",
				}},
				"from_address": "cosmos1shzsqakdakzwhvy05cvjlt9acwf3hfjksy0ht5",
				"to_address":   "cosmos18lu8k4n7nmqhz2z3y9a5y39fzgapchfq6mvaeg",
			}},
			Memo:          "a memo",
			TimeoutHeight: "42",
		},
		AuthInfo: AuthInfo{
			Fee: Fee{
				Amount: []Coin{{
					Amount: "10",
					Denom:  "token",
				}},
				GasLimit: "200000",
			},
		},
	}
	tests := []struct {
		name             string
		keyname          string
		tx               Tx
		expectedSignedTx func(Tx) Tx
	}{
		{
			name:    "local key",
			keyname: "local",
			tx:      bankSendTx,
			expectedSignedTx: func(tx Tx) Tx {
				tx.AuthInfo.SignerInfos = []SignerInfo{{
					PublicKey: map[string]any{
						"@type": codectypes.MsgTypeURL(pubkey),
						"key":   pubkey.Bytes(),
					},
					Sequence: "1",
				}}
				tx.AuthInfo.SignerInfos[0].ModeInfo.Single.Mode = "SIGN_MODE_LEGACY_AMINO_JSON"
				tx.Signatures = [][]byte{
					0xcb, 0x7e, 0x60, 0x7, 0x15, 0xc2, 0xfc, 0x28, 0x2f, 0xcd, 0xd1, 0xa0, 0x93, 0x83, 0x92, 0x33, 0xbe, 0xe, 0xca, 0x84, 0x45, 0x51, 0x28, 0x4c, 0x1e, 0x9e, 0x7f, 0x72, 0xca, 0xfb, 0xa8, 0x99, 0xc8, 0x9a, 0x1e, 0x59, 0x1, 0xc0, 0x7e, 0xe9, 0x2c, 0x6e, 0x23, 0xd, 0x2e, 0xe8, 0x4c, 0x3a, 0x58, 0x69, 0x92, 0xe2, 0x79, 0xd6, 0x85, 0x47, 0xac, 0x35, 0x2c, 0xe8, 0xf1, 0xcd, 0xbb, 0xd,
				}
				return tx
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			signedTx, err := signTx(tt.tx, kr, tt.keyname, "chain-id", 42, 1)

			require.NoError(err)
			fmt.Println(string(signedTx.Signatures[0]))
			require.Equal(tt.expectedSignedTx(tt.tx), signedTx)
		})
	}

	// TODO ensure that SignerInfo & signatures are properly filled
	// Must create a fake keyring with a forged priv key
}
