package main

import (
	"fmt"
	"maps"
	"reflect"
)

type aminoType struct {
	name  string
	enums map[string]map[string]int
	// If filled, the serialization will inline the named field.
	// Useful for secp256k1 and ed25519 keys from cosmos-sdk/crypto/keys, for
	// which the marshalling inlines the Key field instead of an object
	// containing that Key field.
	inlineField string
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
	// cosmos-sdk mapping
	"/cosmos.bank.v1beta1.MsgSend":          {name: "cosmos-sdk/MsgSend"},
	"/cosmos.gov.v1beta1.MsgSubmitProposal": {name: "cosmos-sdk/MsgSubmitProposal"},
	"/cosmos.gov.v1beta1.MsgDeposit":        {name: "cosmos-sdk/MsgDeposit"},
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
	// TODO test other kind of proposal
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
	"/cosmos.staking.v1beta1.MsgCreateValidator": {name: "cosmos-sdk/MsgCreateValidator"},
	"/cosmos.staking.v1beta1.MsgEditValidator":   {name: "cosmos-sdk/MsgEditValidator"},
	"/cosmos.staking.v1beta1.MsgDelegate":        {name: "cosmos-sdk/MsgDelegate"},
	"/cosmos.staking.v1beta1.MsgUndelegate":      {name: "cosmos-sdk/MsgUndelegate"},
	"/cosmos.crypto.secp256k1.PubKey": {
		name:        "tendermint/PubKeySecp256k1",
		inlineField: "key",
	},
	"/cosmos.crypto.ed25519.PubKey": {
		name:        "tendermint/PubKeyEd25519",
		inlineField: "key",
	},

	// Govgen mapping
	"/govgen.gov.v1beta1.MsgSubmitProposal": {name: "cosmos-sdk/MsgSubmitProposal"},
	"/govgen.gov.v1beta1.MsgDeposit":        {name: "cosmos-sdk/MsgDeposit"},
	"/govgen.gov.v1beta1.MsgVote":           {name: "cosmos-sdk/MsgVote"},
	"/govgen.gov.v1beta1.TextProposal":      {name: "cosmos-sdk/TextProposal"},
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
	protoType string
	path      string
}

// handy method to return a new Context with p appended to c.path.
func (c Context) appendPath(p string) Context {
	c.path += "/" + p
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
				"value": _protoToAminoJSON(Context{protoType: protoType}, aminoValue),
			}
		}
		// m has no @type field
		for k, v := range m {
			// fmt.Fprintf(os.Stderr, "MAP ITEM %s\n", k)
			if isEmptyValue(reflect.ValueOf(v)) {
				// fmt.Fprintf(os.Stderr, "EMPTY\n")
				delete(m, k)
				continue
			}
			// append path
			m[k] = _protoToAminoJSON(ctx.appendPath(k), v)
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
		if aminoType, ok := protoToAminoTypeMap[ctx.protoType]; ok {
			// map proto enums if some are configured
			if enumMap, ok := aminoType.enums[ctx.path]; ok {
				enumVal, ok := enumMap[v.(string)]
				if !ok {
					panic(fmt.Sprintf("can't find enum value for type '%s', path '%s' and key '%v'", ctx.protoType, ctx.path, v))
				}
				// fmt.Fprintln(os.Stderr, "ENUM", ctx.path, v, enumVal)
				return enumVal
			}
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
	}
	return false
}
