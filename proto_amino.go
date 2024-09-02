package main

import (
	"fmt"
	"maps"
	"reflect"
)

type aminoType struct {
	name         string
	fieldRenames map[string]string
	enums        map[string]map[string]int
	// If filled, the serialization will inline the named field.
	// Useful for secp256k1 and ed25519 keys from cosmos-sdk/crypto/keys, for
	// which the marshalling inlines the Key field instead of an object
	// containing that Key field.
	inlineField string
	// allowEmpty ensures the named field isn't omitted if it's empty (default
	// behavior)
	allowEmpty string
	// unregistered indicates the type is not registered in amino, therefore it
	// should be marshalled as pure JSON.
	// Useful for some proposals embeded in MsgExecLegacyContent.Content, like
	// ClientUpdate.
	// For some reasons those kinds of proposals are marshalled in pure JSON in
	// MsgExecLegacyContent.Content, while it's not the case for other like
	// SoftwareUpgradeProposal.
	unregistered bool
}

var voteOptionsEnum = map[string]int{
	"VOTE_OPTION_UNSPECIFIED":  0,
	"VOTE_OPTION_YES":          1,
	"VOTE_OPTION_ABSTAIN":      2,
	"VOTE_OPTION_NO":           3,
	"VOTE_OPTION_NO_WITH_VETO": 4,
}

// TODO put this in a config file?
var protoToAminoTypeMap = map[string]aminoType{
	// cosmos-sdk bank module
	"/cosmos.bank.v1beta1.MsgSend":      {name: "cosmos-sdk/MsgSend"},
	"/cosmos.bank.v1beta1.MsgMultiSend": {name: "cosmos-sdk/MsgMultiSend"},

	// cosmos-sdk distribution module
	"/cosmos.distribution.v1beta1.MsgCommunityPoolSpend":                   {name: "cosmos-sdk/distr/MsgCommunityPoolSpend"},
	"/cosmos.distribution.v1beta1.MsgFundCommunityPool":                    {name: "cosmos-sdk/MsgFundCommunityPool"},
	"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress":                   {name: "cosmos-sdk/MsgModifyWithdrawAddress"},
	"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward":              {name: "cosmos-sdk/MsgWithdrawDelegationReward"},
	"/cosmos.distribution.v1beta1.MsgWithdrawTokenizeShareRecordReward":    {name: "cosmos-sdk/MsgWithdrawTokenizeReward"},
	"/cosmos.distribution.v1beta1.MsgWithdrawAllTokenizeShareRecordReward": {name: "cosmos-sdk/MsgWithdrawAllTokenizeReward"},

	// cosmos-sdk auth module
	"/cosmos.auth.v1beta1.MsgUpdateParams": {name: "cosmos-sdk/x/auth/MsgUpdateParams"},

	// cosmos-sdk slashing module
	"/cosmos.slashing.v1beta1.MsgUnjail": {
		name: "cosmos-sdk/MsgUnjail",
		fieldRenames: map[string]string{
			"/validator_addr": "address",
		},
	},

	// ibc module
	"/ibc.core.client.v1.ClientUpdateProposal": {unregistered: true},

	// cosmos-sdk params module
	"/cosmos.params.v1beta1.ParameterChangeProposal": {name: "cosmos-sdk/ParameterChangeProposal"},

	// cosmos-sdk upgrade module
	"/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal":       {name: "cosmos-sdk/SoftwareUpgradeProposal"},
	"/cosmos.upgrade.v1beta1.CancelSoftwareUpgradeProposal": {name: "cosmos-sdk/CancelSoftwareUpgradeProposal"},

	// cosmos-sdk gov module
	"/cosmos.gov.v1beta1.MsgSubmitProposal": {
		name:       "cosmos-sdk/MsgSubmitProposal",
		allowEmpty: "/initial_deposit",
	},
	"/cosmos.gov.v1beta1.MsgDeposit": {name: "cosmos-sdk/MsgDeposit"},
	"/cosmos.gov.v1beta1.MsgVote": {
		name: "cosmos-sdk/MsgVote",
		enums: map[string]map[string]int{
			"/option": voteOptionsEnum,
		},
	},
	"/cosmos.gov.v1beta1.MsgVoteWeighted": {
		name: "cosmos-sdk/MsgVoteWeighted",
		enums: map[string]map[string]int{
			"/options/option": voteOptionsEnum,
		},
	},
	"/cosmos.gov.v1beta1.TextProposal": {name: "cosmos-sdk/TextProposal"},
	"/cosmos.gov.v1.MsgSubmitProposal": {name: "cosmos-sdk/v1/MsgSubmitProposal"},
	"/cosmos.gov.v1.MsgDeposit":        {name: "cosmos-sdk/v1/MsgDeposit"},
	"/cosmos.gov.v1.MsgVote": {
		name: "cosmos-sdk/v1/MsgVote",
		enums: map[string]map[string]int{
			"/option": voteOptionsEnum,
		},
	},
	"/cosmos.gov.v1.MsgVoteWeighted": {
		name: "cosmos-sdk/v1/MsgVoteWeighted",
		enums: map[string]map[string]int{
			"/options/option": voteOptionsEnum,
		},
	},
	"/cosmos.gov.v1.MsgExecLegacyContent": {name: "cosmos-sdk/v1/MsgExecLegacyContent"},

	// cosmos-sdk staking module
	"/cosmos.staking.v1beta1.MsgCreateValidator":             {name: "cosmos-sdk/MsgCreateValidator"},
	"/cosmos.staking.v1beta1.MsgEditValidator":               {name: "cosmos-sdk/MsgEditValidator"},
	"/cosmos.staking.v1beta1.MsgDelegate":                    {name: "cosmos-sdk/MsgDelegate"},
	"/cosmos.staking.v1beta1.MsgUndelegate":                  {name: "cosmos-sdk/MsgUndelegate"},
	"/cosmos.staking.v1beta1.MsgBeginRedelegate":             {name: "cosmos-sdk/MsgBeginRedelegate"},
	"/cosmos.staking.v1beta1.MsgCancelUnbondingDelegation":   {name: "cosmos-sdk/MsgCancelUnbondingDelegation"},
	"/cosmos.staking.v1beta1.MsgValidatorBond":               {name: "cosmos-sdk/MsgValidatorBond"},
	"/cosmos.staking.v1beta1.MsgUnbondValidator":             {name: "cosmos-sdk/MsgUnbondValidator"},
	"/cosmos.staking.v1beta1.MsgTokenizeShares":              {name: "cosmos-sdk/MsgTokenizeShares"},
	"/cosmos.staking.v1beta1.MsgTransferTokenizeShareRecord": {name: "cosmos-sdk/MsgTransferTokenizeRecord"},
	"/cosmos.staking.v1beta1.MsgEnableTokenizeShares":        {name: "cosmos-sdk/MsgEnableTokenizeShares"},
	"/cosmos.staking.v1beta1.MsgDisableTokenizeShares":       {name: "cosmos-sdk/MsgDisableTokenizeShares"},
	"/cosmos.staking.v1beta1.MsgRedeemTokensForShares":       {name: "cosmos-sdk/MsgRedeemTokensForShares"},

	// Govgen gov module
	"/govgen.gov.v1beta1.MsgSubmitProposal": {
		name:       "govgen/MsgSubmitProposal",
		allowEmpty: "/initial_deposit",
	},
	"/govgen.gov.v1beta1.MsgDeposit": {name: "govgen/MsgDeposit"},
	"/govgen.gov.v1beta1.MsgVote": {
		name: "govgen/MsgVote",
		enums: map[string]map[string]int{
			"/option": voteOptionsEnum,
		},
	},
	"/govgen.gov.v1beta1.MsgVoteWeighted": {
		name: "govgen/MsgVoteWeighted",
		enums: map[string]map[string]int{
			"/options/option": voteOptionsEnum,
		},
	},
	"/govgen.gov.v1beta1.TextProposal": {name: "govgen/TextProposal"},

	// misc mapping
	"/cosmos.crypto.secp256k1.PubKey": {
		name:        "tendermint/PubKeySecp256k1",
		inlineField: "key",
	},
	"/cosmos.crypto.ed25519.PubKey": {
		name:        "tendermint/PubKeyEd25519",
		inlineField: "key",
	},
}

// protoToAminoJSON turns proto json to amino json.
// It works by mapping the proto `@type` into amino `type`, and then
// encapsulate the other fields in a amino `value` field.
// TODO add parameter proto-to-amino map to extend the global map.
//
// Contract: should never alter v
func protoToAminoJSON(v any) (ret any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	ret = _protoToAminoJSON(Context{}, v)
	return
}

// Context helps to keep track of current protoType and path.
type Context struct {
	aminoType aminoType
	path      string
}

func (c Context) withPath(p string) Context {
	c.path = p
	return c
}

func _protoToAminoJSON(ctx Context, v any) any {
	// fmt.Fprintf(os.Stderr, "TYPE %s %T\n", v, v)
	x := reflect.ValueOf(v)
	switch x.Kind() {

	case reflect.Map:
		// fmt.Fprintf(os.Stderr, "MAP\n")
		m := maps.Clone(v.(map[string]any))
		// Check if it's a proto @type
		if typ, ok := m["@type"]; ok {
			protoType := typ.(string)
			aminoType, ok := protoToAminoTypeMap[protoType]
			if !ok {
				panic(fmt.Sprintf("can't find amino type for proto @type='%s'", protoType))
			}
			// fmt.Fprintf(os.Stderr, "@TYPE %v %v\n", typ, aminoType)
			// remove field @type
			delete(m, "@type")
			if aminoType.unregistered {
				// marshal as standard JSON
				return _protoToAminoJSON(ctx, m)
			}
			// return pseudo amino
			var aminoValue any = m
			if aminoType.inlineField != "" {
				// if inlineField is provided, then its value replace the aminoValue
				if aminoValue, ok = m[aminoType.inlineField]; !ok {
					panic(fmt.Sprintf(
						"amino type '%s' configured to inline field '%s', but field not found in '%v'",
						aminoType.name, aminoType.inlineField, m,
					))
				}
			}
			return map[string]any{
				"type":  aminoType.name,
				"value": _protoToAminoJSON(Context{aminoType: aminoType}, aminoValue),
			}
		}
		// m has no @type field
		for k, v := range m {
			path := ctx.path + "/" + k
			// fmt.Fprintf(os.Stderr, "MAP ITEM path=%s allowempty=%s\n", path, ctx.aminoType.allowEmpty)
			if isEmptyValue(reflect.ValueOf(v)) && ctx.aminoType.allowEmpty != path {
				// fmt.Fprintf(os.Stderr, "EMPTY\n")
				delete(m, k)
				continue
			}
			if newName, ok := ctx.aminoType.fieldRenames[path]; ok {
				delete(m, k)
				k = newName
			}
			m[k] = _protoToAminoJSON(ctx.withPath(path), v)
		}
		return m

	case reflect.Slice:
		// fmt.Fprintf(os.Stderr, "SLICE\n")
		// duplicate the slice, call _protoToAminoJSON for each item.
		s := reflect.MakeSlice(x.Type(), x.Len(), x.Cap())
		for i := 0; i < x.Len(); i++ {
			s.Index(i).Set(reflect.ValueOf(_protoToAminoJSON(ctx, x.Index(i).Interface())))
		}
		return s.Interface()

	default:
		// fmt.Fprintf(os.Stderr, "DEFAULT %s %v\n", ctx.path, v)
		// map proto enums if some are configured
		if enumMap, ok := ctx.aminoType.enums[ctx.path]; ok {
			enumVal, ok := enumMap[v.(string)]
			if !ok {
				panic(fmt.Sprintf(
					"can't find enum value for type '%s', path '%s' and key '%v'",
					ctx.aminoType.name, ctx.path, v,
				))
			}
			// fmt.Fprintln(os.Stderr, "ENUM", ctx.path, v, enumVal)
			return enumVal
		}
		return v
	}
}

// Scavanged from https://go-review.googlesource.com/c/go/+/482415
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return v.Bool() == false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	case reflect.Invalid:
		return true
	}
	return false
}
