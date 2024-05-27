package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProtoToAminoJSON(t *testing.T) {
	tests := []struct {
		name          string
		m             map[string]any
		expectedAmino map[string]any
		expectedError string
	}{
		{
			name: "proto type not found",
			m: map[string]any{
				"@type": "xxx",
			},
			expectedError: "can't find amino type for proto @type='xxx'",
		},
		{
			name: "empty fields are omitted",
			m: map[string]any{
				"empty-string": "",
				"zero-int":     0,
				"empty-slice":  []int{},
				"a": map[string]any{
					"a1":           1,
					"empty-string": "",
					"zero-int":     0,
					"empty-slice":  []string{},
				},
			},
			expectedAmino: map[string]any{
				"a": map[string]any{
					"a1": 1,
				},
			},
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
		{
			name: "proto types in array",
			m: map[string]any{
				"a": []map[string]any{
					{
						"@type": "/cosmos.bank.v1beta1.MsgSend",
						"a1":    42,
					},
					{
						"@type": "/cosmos.bank.v1beta1.MsgSend",
						"a1":    44,
					},
				},
			},
			expectedAmino: map[string]any{
				"a": []map[string]any{
					{
						"type": "cosmos-sdk/MsgSend",
						"value": map[string]any{
							"a1": 42,
						},
					},
					{
						"type": "cosmos-sdk/MsgSend",
						"value": map[string]any{
							"a1": 44,
						},
					},
				},
			},
		},
		{
			name: "MsgVote.option enum",
			m: map[string]any{
				"a": []map[string]any{
					{
						"@type":  "/cosmos.gov.v1.MsgVote",
						"option": "VOTE_OPTION_UNSPECIFIED",
						"x":      "_",
					},
					{
						"@type":  "/cosmos.gov.v1.MsgVote",
						"option": "VOTE_OPTION_YES",
						"x":      "yes",
					},
					{
						"@type":  "/cosmos.gov.v1.MsgVote",
						"option": "VOTE_OPTION_ABSTAIN",
						"x":      "abstain",
					},
					{
						"@type":  "/cosmos.gov.v1.MsgVote",
						"option": "VOTE_OPTION_NO",
						"x":      "no",
					},
					{
						"@type":  "/cosmos.gov.v1.MsgVote",
						"option": "VOTE_OPTION_NO_WITH_VETO",
						"x":      "nwv",
					},
					{
						"@type":  "/cosmos.gov.v1beta1.MsgVote",
						"option": "VOTE_OPTION_UNSPECIFIED",
						"x":      "_",
					},
					{
						"@type":  "/cosmos.gov.v1beta1.MsgVote",
						"option": "VOTE_OPTION_YES",
						"x":      "yes",
					},
					{
						"@type":  "/cosmos.gov.v1beta1.MsgVote",
						"option": "VOTE_OPTION_ABSTAIN",
						"x":      "abstain",
					},
					{
						"@type":  "/cosmos.gov.v1beta1.MsgVote",
						"option": "VOTE_OPTION_NO",
						"x":      "no",
					},
					{
						"@type":  "/cosmos.gov.v1beta1.MsgVote",
						"option": "VOTE_OPTION_NO_WITH_VETO",
						"x":      "nwv",
					},
				},
			},
			expectedAmino: map[string]any{
				"a": []map[string]any{
					{
						"type": "cosmos-sdk/v1/MsgVote",
						"value": map[string]any{
							"option": 0,
							"x":      "_",
						},
					},
					{
						"type": "cosmos-sdk/v1/MsgVote",
						"value": map[string]any{
							"option": 1,
							"x":      "yes",
						},
					},
					{
						"type": "cosmos-sdk/v1/MsgVote",
						"value": map[string]any{
							"option": 2,
							"x":      "abstain",
						},
					},
					{
						"type": "cosmos-sdk/v1/MsgVote",
						"value": map[string]any{
							"option": 3,
							"x":      "no",
						},
					},
					{
						"type": "cosmos-sdk/v1/MsgVote",
						"value": map[string]any{
							"option": 4,
							"x":      "nwv",
						},
					},
					{
						"type": "cosmos-sdk/MsgVote",
						"value": map[string]any{
							"option": 0,
							"x":      "_",
						},
					},
					{
						"type": "cosmos-sdk/MsgVote",
						"value": map[string]any{
							"option": 1,
							"x":      "yes",
						},
					},
					{
						"type": "cosmos-sdk/MsgVote",
						"value": map[string]any{
							"option": 2,
							"x":      "abstain",
						},
					},
					{
						"type": "cosmos-sdk/MsgVote",
						"value": map[string]any{
							"option": 3,
							"x":      "no",
						},
					},
					{
						"type": "cosmos-sdk/MsgVote",
						"value": map[string]any{
							"option": 4,
							"x":      "nwv",
						},
					},
				},
			},
		},
		{
			name: "MsgVoteWeighted.options[].option enum",
			m: map[string]any{
				"a": []map[string]any{
					{
						"@type": "/cosmos.gov.v1.MsgVoteWeighted",
						"options": []map[string]any{
							{
								"option": "VOTE_OPTION_UNSPECIFIED",
								"weight": "0.1",
							},
							{
								"option": "VOTE_OPTION_YES",
								"weight": "0.2",
							},
							{
								"option": "VOTE_OPTION_ABSTAIN",
								"weight": "0.3",
							},
							{
								"option": "VOTE_OPTION_NO",
								"weight": "0.4",
							},
							{
								"option": "VOTE_OPTION_NO_WITH_VETO",
								"weight": "0.5",
							},
						},
					},
					{
						"@type": "/cosmos.gov.v1beta1.MsgVoteWeighted",
						"options": []map[string]any{
							{
								"option": "VOTE_OPTION_UNSPECIFIED",
								"weight": "0.1",
							},
							{
								"option": "VOTE_OPTION_YES",
								"weight": "0.2",
							},
							{
								"option": "VOTE_OPTION_ABSTAIN",
								"weight": "0.3",
							},
							{
								"option": "VOTE_OPTION_NO",
								"weight": "0.4",
							},
							{
								"option": "VOTE_OPTION_NO_WITH_VETO",
								"weight": "0.5",
							},
						},
					},
				},
			},
			expectedAmino: map[string]any{
				"a": []map[string]any{
					{
						"type": "cosmos-sdk/v1/MsgVoteWeighted",
						"value": map[string]any{
							"options": []map[string]any{
								{
									"option": 0,
									"weight": "0.1",
								},
								{
									"option": 1,
									"weight": "0.2",
								},
								{
									"option": 2,
									"weight": "0.3",
								},
								{
									"option": 3,
									"weight": "0.4",
								},
								{
									"option": 4,
									"weight": "0.5",
								},
							},
						},
					},
					{
						"type": "cosmos-sdk/MsgVoteWeighted",
						"value": map[string]any{
							"options": []map[string]any{
								{
									"option": 0,
									"weight": "0.1",
								},
								{
									"option": 1,
									"weight": "0.2",
								},
								{
									"option": 2,
									"weight": "0.3",
								},
								{
									"option": 3,
									"weight": "0.4",
								},
								{
									"option": 4,
									"weight": "0.5",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "inline field",
			m: map[string]any{
				"pubkey": map[string]any{
					"@type": "/cosmos.crypto.secp256k1.PubKey",
					"key":   "AjbjaJ/tXxhwPLxsg+bZSiNsn/Ony6af7cOa+QULXCn3",
				},
			},
			expectedAmino: map[string]any{
				"pubkey": map[string]any{
					"type":  "tendermint/PubKeySecp256k1",
					"value": "AjbjaJ/tXxhwPLxsg+bZSiNsn/Ony6af7cOa+QULXCn3",
				},
			},
		},
		{
			name: "rename field",
			m: map[string]any{
				"msg": map[string]any{
					"@type":          "/cosmos.slashing.v1beta1.MsgUnjail",
					"validator_addr": "xxx",
				},
			},
			expectedAmino: map[string]any{
				"msg": map[string]any{
					"type": "cosmos-sdk/MsgUnjail",
					"value": map[string]any{
						"address": "xxx",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)
			orig := fmt.Sprint(tt.m)

			amino, err := protoToAminoJSON(tt.m)

			if tt.expectedError != "" {
				require.EqualError(err, tt.expectedError)
				return
			}
			require.NoError(err)
			assert.Equal(tt.expectedAmino, amino)
			assert.Equal(orig, fmt.Sprint(tt.m), "input parameter has been altered")
		})
	}
}
