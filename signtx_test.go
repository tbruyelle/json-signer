package main

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
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
	// (with a secret so it generates the same private key, else we wouldn't be
	// able to assert the signtures).
	var (
		privkey = ed25519.GenPrivKeyFromSecret([]byte("secret"))
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
		name                string
		keyname             string
		tx                  Tx
		bytesToSign         string
		expectedSignerInfos []SignerInfo
		expectedSignatures  []string
	}{
		{
			name:    "local key",
			keyname: "local",
			tx:      bankSendTx,
			expectedSignerInfos: []SignerInfo{{
				PublicKey: map[string]any{
					"@type": codectypes.MsgTypeURL(pubkey),
					"key":   pubkey.Bytes(),
				},
				ModeInfo: ModeInfo{
					Single: Single{
						Mode: "SIGN_MODE_LEGACY_AMINO_JSON",
					},
				},
				Sequence: "1",
			}},
			expectedSignatures: []string{
				"NsA6KJYcBaMI9edMV4H0vKHDiOBzu4J2e3xQc0WuIqPt6O0UeJ0zsBcw4X+o+ZkiPsEZ5kOVF8AzC4O4XHViDA==",
			},
		},
		{
			name:        "local key with bytes-to-sign",
			keyname:     "local",
			tx:          bankSendTx,
			bytesToSign: "eyJhY2NvdW50X251bWJlciI6IjQyIiwiY2hhaW5faWQiOiJjaGFpbi1pZCIsImZlZSI6eyJhbW91bnQiOlt7ImFtb3VudCI6IjEwIiwiZGVub20iOiJ0b2tlbiJ9XSwiZ2FzIjoiMjAwMDAwIn0sIm1lbW8iOiJhIG1lbW8iLCJtc2dzIjpbeyJ0eXBlIjoiY29zbW9zLXNkay9Nc2dTZW5kIiwidmFsdWUiOnsiYW1vdW50IjpbeyJhbW91bnQiOiIxMDAwIiwiZGVub20iOiJ0b2tlbiJ9XSwiZnJvbV9hZGRyZXNzIjoiY29zbW9zMXNoenNxYWtkYWt6d2h2eTA1Y3ZqbHQ5YWN3ZjNoZmprc3kwaHQ1IiwidG9fYWRkcmVzcyI6ImNvc21vczE4bHU4azRuN25tcWh6MnozeTlhNXkzOWZ6Z2FwY2hmcTZtdmFlZyJ9fV0sInNlcXVlbmNlIjoiMSIsInRpbWVvdXRfaGVpZ2h0IjoiNDIifQ==",
			expectedSignerInfos: []SignerInfo{{
				PublicKey: map[string]any{
					"@type": codectypes.MsgTypeURL(pubkey),
					"key":   pubkey.Bytes(),
				},
				ModeInfo: ModeInfo{
					Single: Single{
						Mode: "SIGN_MODE_LEGACY_AMINO_JSON",
					},
				},
				Sequence: "1",
			}},
			expectedSignatures: []string{
				"NsA6KJYcBaMI9edMV4H0vKHDiOBzu4J2e3xQc0WuIqPt6O0UeJ0zsBcw4X+o+ZkiPsEZ5kOVF8AzC4O4XHViDA==",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			signedTx, err := signTx(tt.tx, kr, tt.keyname, "chain-id", 42, 1, tt.bytesToSign)

			require.NoError(err)
			assert.Equal(tt.expectedSignerInfos, signedTx.AuthInfo.SignerInfos)
			for i := 0; i < len(tt.expectedSignatures); i++ {
				signature := base64.StdEncoding.EncodeToString(signedTx.Signatures[i])
				assert.Equal(tt.expectedSignatures[i], signature)
			}
		})
	}
}
