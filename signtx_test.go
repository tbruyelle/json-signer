package main

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tbruyelle/json-signer/keyring"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func TestProtoToAminoJSON(t *testing.T) {
	tests := []struct {
		name          string
		m             map[string]any
		expectedAmino map[string]any
		expectedPanic string
	}{
		{
			name: "proto type not found",
			m: map[string]any{
				"@type": "xxx",
			},
			expectedPanic: "can't find amino mapping for proto @type=\"xxx\"",
		},
		{
			name: "no proto type",
			m: map[string]any{
				"a": 1,
				"b": map[string]any{
					"b1": "xxx",
					"b2": []int{1},
					"b3": map[string]any{
						"b31": "x",
					},
				},
				"c": []string{"x", "xx"},
			},
			expectedAmino: map[string]any{
				"a": 1,
				"b": map[string]any{
					"b1": "xxx",
					"b2": []int{1},
					"b3": map[string]any{
						"b31": "x",
					},
				},
				"c": []string{"x", "xx"},
			},
		},
		{
			name: "one proto type",
			m: map[string]any{
				"@type": "/cosmos.bank.v1beta1.MsgSend",
				"a":     1,
				"b": map[string]any{
					"b1": "xxx",
					"b2": []int{1},
					"b3": map[string]any{
						"b31": "x",
					},
				},
				"c": []string{"x", "xx"},
			},
			expectedAmino: map[string]any{
				"type": "cosmos-sdk/MsgSend",
				"value": map[string]any{
					"a": 1,
					"b": map[string]any{
						"b1": "xxx",
						"b2": []int{1},
						"b3": map[string]any{
							"b31": "x",
						},
					},
					"c": []string{"x", "xx"},
				},
			},
		},
		{
			name: "multiple proto types",
			m: map[string]any{
				"@type": "/cosmos.bank.v1beta1.MsgSend",
				"a":     1,
				"b": map[string]any{
					"b1": "xxx",
					"b2": []int{1},
					"b3": map[string]any{
						"@type": "/cosmos.bank.v1beta1.MsgSend",
						"b31":   "x",
					},
				},
				"c": []string{"x", "xx"},
				"d": map[string]any{
					"@type": "/cosmos.bank.v1beta1.MsgSend",
					"c1":    42,
					"c2": map[string]any{
						"c21": []int{1, 2},
					},
				},
			},
			expectedAmino: map[string]any{
				"type": "cosmos-sdk/MsgSend",
				"value": map[string]any{
					"a": 1,
					"b": map[string]any{
						"b1": "xxx",
						"b2": []int{1},
						"b3": map[string]any{
							"type": "cosmos-sdk/MsgSend",
							"value": map[string]any{
								"b31": "x",
							},
						},
					},
					"c": []string{"x", "xx"},
					"d": map[string]any{
						"type": "cosmos-sdk/MsgSend",
						"value": map[string]any{
							"c1": 42,
							"c2": map[string]any{
								"c21": []int{1, 2},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)
			if tt.expectedPanic != "" {
				defer func() {
					r := recover()
					if r == nil || r.(string) != tt.expectedPanic {
						require.Fail("expected panic %q got %q", tt.expectedPanic, r)
					}
				}()
			}

			amino := protoToAminoJSON(tt.m)

			assert.Equal(tt.expectedAmino, amino)
		})
	}
}

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

			bytesToSign, err := getBytesToSign(tt.tx, "chainid-1", "1", "2")

			require.NoError(err)
			require.JSONEq(tt.expected, string(bytesToSign))
		})
	}
}

func TestSignTx(t *testing.T) {
	kr, err := keyring.New(keyring.BackendType("file"), t.TempDir(),
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
						Mode: signModeAminoJSON,
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

			signedTx, err := signTx(tt.tx, kr, tt.keyname, "chain-id", "42", "1")

			require.NoError(err)
			assert.Equal(tt.expectedSignerInfos, signedTx.AuthInfo.SignerInfos)
			for i := 0; i < len(tt.expectedSignatures); i++ {
				signature := base64.StdEncoding.EncodeToString(signedTx.Signatures[i])
				assert.Equal(tt.expectedSignatures[i], signature)
			}
		})
	}
}
