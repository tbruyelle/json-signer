package main

import (
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
			expectedPanic: "can't find amino mapping for proto @type=\"xxx\"",
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
