# json-signer

Sign any Cosmos-SDK transaction using the amino-json sign mode.

`json-signer` supports signing with local private key, ledger and multisig
accounts.

`json-signer` has [E2E tests] for the gaia and govgen chains, but it should
work just well with other Cosmos-SDK chains.

## Why a separate tool

Your private keys need to be secure, they shouldn't be read by an unaudited
tool.

Cosmos-sdk applications embed a CLI that can read private keys, but the binary
is frequently updated and not always audited in time. There's always a risk
that the lastest version will contain an exploit, a bug or a vulnerability that
could expose your private keys to malicious people. 

It is therefore preferable to rely on a single tool to sign your transactions.
`json-signer` embraces this and aims to deliver an audited tool that can sign
any Cosmos-SDK transaction.

## Why amino-json

Amino-json allows to build the bytes-to-sign without the needs for protobuf
types registrations. This is important because otherwise this tool would have
to rely on the latest version of the chain binaries to generate these bytes,
which would require frequent updates and thus defeat the purpose of the tool,
which is to be independent of the chain binaries.

## Usage

```sh
$ json-signer sign-tx 
DESCRIPTION
  Sign transaction

USAGE
  json-signer sign-tx -from=<key> \
    -keyring-backend=<keychain|pass|kwallet|file> \
    -chain-id=<chainID> \
    -sequence=<sequence> \
    -account=<account-number> \
    <tx.json>

FLAGS
  -account string          Account number
  -chain-id string         Chain identifier
  -from string             Signer key name
  -keyring-backend string  Keyring backend, which can be one of 'keychain'
                           (macos), 'pass', 'kwallet' (linux), or 'file'
  -keyring-dir string      Keyring directory
                           (mandatory with -keyring-backend=file)
  -sequence string         Sequence number
  -signature-only=false    Outputs only the signature data (useful for multisig)
```

### Keyring backend selection

Unlike Cosmos-SDK apps, there's no automatic selection of keyring backend, you
have to specify the exact keyring backend where your keys are stored. If you
don't know, use `json-signer list-keys` with different backends until you find
out where your keys are stored.

### Target the `test` keyring backend

To use the `test` keyring backend which is targeted with the
`--keyring-backend=test` flag in Cosmos-SDK app CLIs, use the following flags
in `json-signer` (using `gaia` as an example):

```sh
-keyring-backend=file -keyring-dir=~/.gaia/keyring-test
```

Unlike Cosmos-SDK app CLIs, a password is required, and the password is `test`.

## Caveats

Unlike Cosmos-SDK app, `json-signer` is not able to check if the signer is the
one expected by the message. For example, for a `MsgSend` message, the signer
must match the `from_address` field, and the Cosmos-SDK app CLI will reject the
attempt to sign such a transaction if it doesn't. The `json-signer` can't do
this because it doesn't depend on protobuf types like `MsgSend` and others.

But this is not a risk, if the signer is not the expected account, the error
will occur after the signing, when the signature is verified or when the
transaction is broadcasted.

## Example using gaia

> [!NOTE]
> This example follows the procedure described in this [guide], please refer to
> it for more details.

In this example, there is an online computer and an offline computer. For
security reasons, only the offline computer contains the private key.

Let us create a transaction using the `gaiad` binary on the online computer:

```sh
$ gaiad tx bank send [addr1] [addr2] 100000uatom --chain-id cosmoshub-4 \
    --fees 1000uatom --account-number 12345 --sequence 123 --gas auto \
    --generate-only > tx.json
```

Once executed, the `tx.json` will contain the unsigned transaction. This file
must be copied to the offline computer.

From the offline computer, use `json-signer` to sign the tx:

```sh
$ json-signer sign-tx --from=addr1 --keyring-dir=~/.gaia --keyring-backend=file \
    --account=12345 --sequence=123 --chain-id=cosmoshub-4 tx.json \
    > tx-signed.json
```

> [!NOTE]
> This command assumes that the keyring backend is `file` and is stored in the
> `~/.gaia` directory. Adjust the `--keyring-backend` flag to your needs.

Copy `tx-signed.json` to the online computer. You can validate the signature by
running:

```sh
$ gaiad tx validate-signatures tx-signed.json
```

The command should not return any error. You can now broadcast the transaction:

```sh
$ gaiad tx broadcast tx-signed.json
```

Congratulations, your transaction is on its way to being executed on the chain.

## Example with multisig account

`json-signer` supports signing transaction with multisig accounts. The
procedure is quite similar as signing with standard accounts, with just a few
changes in the flags.

Let's take `bob` and `alice` as two standard accounts, and `bob-alice` as a
multisig account between `bob` and `alice`.

Let's first generate the transaction with the blockchain binary:

```sh
$ gaiad tx bank send [bob-alice-addr] [other-addr] 100000uatom \
    --chain-id cosmoshub-4 --fees 1000uatom --account-number 12345 \
    --sequence 123 --gas auto --generate-only > tx.json
```

From the offline computer, use `json-signer` to sign the tx with `bob`:

```sh
$ json-signer sign-tx --from=bob --signature-only \
    --keyring-dir=~/.gaia --keyring-backend=file \
    --account=12345 --sequence=123 --chain-id=cosmoshub-4 tx.json \
    > tx-bob-signature.json
```

Similarly, let's create the signature for the `alice` account:

```sh
$ json-signer sign-tx --from=alice --signature-only \
    --keyring-dir=~/.gaia --keyring-backend=file \
    --account=12345 --sequence=123 --chain-id=cosmoshub-4 tx.json \
    > tx-alice-signature.json
```

> [!WARNING]
> - the `account-number` and `sequence` belong to the `bob-alice` multisig
>   account.
> - the `--signature-only` flag is necessary for `json-signer` to output only
>   the signature of the tx, which is needed for the next steps.

Once `tx-bob-signature.json` and `tx-alice-signature.json` files are ready,
let's copy them to the online computer and use the blockchain binary to
generate the final transaction:

```sh
$ gaiad tx multi-sign tx.json bob-alice tx-bob-signature.json tx-alice-signature.json \
  > tx-signed.json
```

The generated file `tx-signed.json` is multi-signed and ready to be
broadcasted.
 
## Keyring

`json-signer` also uses a alternate [keyring] package that is able to read
a keyring, whether the keys are amino or proto encoded (before Cosmos-SDK 0.46,
the keyring was amino encoded, then it has been migrated to protobuf encoding).

And of course it doesn't automatically migrate amino encoded keys to proto keys
like Cosmos-SDK applications >=0.46 does.

[guide]: https://github.com/atomone-hub/govgen-proposals/blob/main/submit-tx-securely.md
[keyring]: https://github.com/tbruyelle/keyring-compat
[E2E tests]: https://github.com/tbruyelle/json-signer/tree/master/testdata
