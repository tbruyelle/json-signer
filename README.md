# amino-signer

Sign any cosmos transaction using the amino-json sign mode.

Usage:
```
$ amino-signer --from=KEY --keyring-dir=~/.gaia --account=42 --sequence=43 -chain-id=cosmoshub-4  tx.json
```

Where `tx.json` is the file generated by the blockchain binary with the
`--generate-only` flag.


## Why amino-json

Amino-json allows to build the bytes-to-sign without the needs for protobuf
types registrations.

For example, this transaction is easy to translate to amino-json bytes-to-sign,
while it's impossible with the direct sign-mode because it requires the
`MsgSend` proto generated type to be available.

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

`amino-signer` also uses a alternate keyring package that is able to read a
keyring (only file backend for now), whether it has been created with a
cosmos-sdk application version <0.46 (amino encoded key), or with version >=0.46
(proto encoded key). And of course it doesn't automatically migrate amino
encoded keys to proto keys, like cosmos-sdk applications >=0.46 are doing.
