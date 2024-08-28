package main

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tbruyelle/keyring-compat"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func TestSignBatchTx(t *testing.T) {
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
	err = kr.AddProto("local.info", record)
	require.NoError(t, err)
	key, err := kr.Get("local.info")
	require.NoError(t, err)
	pubKeyBz, err := key.ProtoJSONPubKey()
	require.NoError(t, err)
	tests := []struct {
		name                string
		keyname             string
		expectedSignerInfos []SignerInfo
		expectedSignatures  []string
	}{
		{
			name:    "local key",
			keyname: "local",
			expectedSignerInfos: []SignerInfo{
				{
					PublicKey: pubKeyBz,
					ModeInfo: ModeInfo{
						Single: Single{
							Mode: signModeAminoJSON,
						},
					},
					Sequence: "1",
				},
				{
					PublicKey: pubKeyBz,
					ModeInfo: ModeInfo{
						Single: Single{
							Mode: signModeAminoJSON,
						},
					},
					Sequence: "2",
				},
				{
					PublicKey: pubKeyBz,
					ModeInfo: ModeInfo{
						Single: Single{
							Mode: signModeAminoJSON,
						},
					},
					Sequence: "3",
				},
				{
					PublicKey: pubKeyBz,
					ModeInfo: ModeInfo{
						Single: Single{
							Mode: signModeAminoJSON,
						},
					},
					Sequence: "4",
				},
				{
					PublicKey: pubKeyBz,
					ModeInfo: ModeInfo{
						Single: Single{
							Mode: signModeAminoJSON,
						},
					},
					Sequence: "5",
				},
			},
			expectedSignatures: []string{
				"d1/D9FzsdsIE5TIqhG687MHEALakPIJOfT01ZSx2moSqNYsTetGpQjJ4f58W58AchO3mSz7x/yPBZpeFUPiFDA==",
				"UX9NiZlUYUAO+O/cogZBWJSdK7MQVtkGwz6PBilFkiG4qH9BMAKLiWCuaE5sK1mFsjl15qJj6jMk7xzoDm1/AQ==",
				"UqX2htnKSp4+lqmdr1LICRcs9t2ft2rqGAqfatHaHDni22ZWpUuZycCGQwLTYTy+N32b/DFo28Vx10QS99AzCQ==",
				"psqSiKaoTtDF2ANGjzdJ1Yq7WrTCj8kkiR8Id6UuTKlSpLg0t9fzdvU1aOlgNILFMMqlsiVcFT5zvSMevvz1Bg==",
				"ETzhWMnpsCj+F3FVNDDO34cZwwFmOhvwBns+peMBA0GIOlEy7y5PBFDfFzI7Ti2NcFLEiuaKLmAu78flcX7QBQ==",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			signedTxs, err := batchSignTxs([]string{"./testdata/batch_tx_1.txt", "./testdata/batch_tx_2.txt"}, kr, tt.keyname, "chain-id", "42", "1")
			require.NoError(err)
			assert.Equal(len(signedTxs), len(tt.expectedSignatures))

			for i, signedTx := range signedTxs {
				assert.Equal(tt.expectedSignerInfos[i], signedTx.AuthInfo.SignerInfos[0])

				signature := base64.StdEncoding.EncodeToString(signedTx.Signatures[0])
				assert.Equal(tt.expectedSignatures[i], signature)
			}
		})
	}
}
