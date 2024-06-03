# json-signer

Sign any cosmos transaction using the amino-json sign mode.

In the context of offline signing, it's preferable to rely on a single tool to
sign your transaction, instead of the blockchain binary itself, which is often
updated and not always audited on time. `json-signer` embraces this and aims to
deliver an audited tool that is able to sign any cosmos-sdk transaction.

## Example using gaia

In this example, there is an online computer and an offline computer. For
security reasons, only the offline computer contains the private key.

Let us create a transaction using the `gaiad` binary on the online computer:

> [!NOTE]
> This example follows the procedure described in this [guide], please refer to
> it for more details.

```sh
$ gaiad tx bank send [addr1] [addr2] 100000uatom --chain-id cosmoshub-4 \
    --fees 1000uatom --account-number 12345 --sequence 123 --gas auto \
    --generate-only > tx.json
```

Once executed, the `tx.json` will contain the unsigned transaction. This file
must be copied to the offline computer.

From the offline computer, use `json-signer` to sign the tx:

``sh
$ json-signer sign-tx --from=addr1 --keyring-dir=~/.gaia --account=12345 \
    --sequence=123 --chain-id=cosmoshub-4 tx.json > tx-signed.json
```

Copy `tx-signed.json` to the online computer. You can validate the signature by
running:

```sh
$ gaiad tx validate-signatures tx-signed.json
```

The command should not return any error. You can now broadcast the transaction:

```sh
$ gaiad tx broadcast tx-signed.json
```

Congratulations, your transaction should be on its way to be executed on the
chain.

## Why amino-json

Amino-json allows to build the bytes-to-sign without the needs for protobuf
types registrations. This is important because otherwise this tool would have
to rely on the latest version of the chain binaries to generate these bytes,
which would require frequent updates and thus defeat the purpose of the tool,
which is to be independent of the chain binaries.

For example, this transaction is easy to translate to amino-json bytes-to-sign,
while it's impossible with the direct sign-mode because it requires the
`MsgSend` proto generated type to be compiled within this tool.

```json
{
  "account_number": "682802",
  "chain_id": "cosmoshub-4",
  "fee": {
    "amount": [
      {
        "amount": "1000",
        "denom": "uatom"
      }
    ],
    "gas": "66701"
  },
  "memo": "",
  "msgs": [
    {
      "type": "cosmos-sdk/MsgSend",
      "value": {
        "amount": [
          {
            "amount": "100000",
            "denom": "uatom"
          }
        ],
        "from_address": "cosmos1shzsqakdakzwhvy05cvjlt9acwf3hfjksy0ht5",
        "to_address": "cosmos18lu8k4n7nmqhz2z3y9a5y39fzgapchfq6mvaeg"
      }
    }
  ],
  "sequence": "391"
}
```
Extracted bytes-to-sign:
```
{
  "account_number": "682802",
  "chain_id": "cosmoshub-4",
  "fee": {
    "amount": [
      {
        "amount": "1000",
        "denom": "uatom"
      }
    ],
    "gas": "66701"
  },
  "memo": "",
  "msgs": [
    {
      "type": "cosmos-sdk/MsgSend",
      "value": {
        "amount": [
          {
            "amount": "100000",
            "denom": "uatom"
          }
        ],
        "from_address": "cosmos1shzsqakdakzwhvy05cvjlt9acwf3hfjksy0ht5",
        "to_address": "cosmos18lu8k4n7nmqhz2z3y9a5y39fzgapchfq6mvaeg"
      }
    }
  ],
  "sequence": "391"
}
```

## Keyring

`json-signer` also uses a alternate [`keyring`] package that is able to read
a keyring, whether the keys are amino or proto encoded (before cosmos-sdk 0.46,
the keyring was amino encoded, then it has been migrated to protobuf encoding).

And of course it doesn't automatically migrate amino encoded keys to proto keys
like cosmos-sdk applications >=0.46 does.

[guide]: https://github.com/atomone-hub/govgen-proposals/blob/main/submit-tx-securely.md
[keyring]: https://github.com/tbruyelle/keyring-compat
