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
		expectedPanic string
	}{
		{
			name: "proto type not found",
			m: map[string]any{
				"@type": "xxx",
			},
			expectedPanic: "can't find amino mapping for proto @type='xxx'",
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
							// NOTE: VOTE_OPTION_UNSPECIFIED is not present because it's an empty
							// value (OK or KO? we'll see later, for now this vote option isn't
							// available from the cli).
							// "option": 0,
							"x": "_",
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
							// NOTE: VOTE_OPTION_UNSPECIFIED is not present because it's an empty
							// value (OK or KO? we'll see later, for now this vote option isn't
							// available from the cli).
							// "option": 0,
							"x": "_",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)
			if tt.expectedPanic != "" {
				defer func() {
					r := recover()
					if r == nil || r.(string) != tt.expectedPanic {
						require.Failf("expected panic", "want %q got %q", tt.expectedPanic, r)
					}
				}()
			}
			orig := fmt.Sprint(tt.m)

			amino := protoToAminoJSON(tt.m)

			assert.Equal(tt.expectedAmino, amino)
			assert.Equal(orig, fmt.Sprint(tt.m), "input parameter has been altered")
		})
	}
}
