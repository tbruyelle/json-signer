package main

import (
	"testing"

	"github.com/stretchr/testify/require"
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
	// TODO ensure that SignerInfo & signatures are properly filled
	// Must create a fake keyring with a forged priv key
}
