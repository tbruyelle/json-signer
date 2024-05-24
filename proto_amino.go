package main

import (
	"fmt"
	"maps"
	"reflect"
)

type aminoType struct {
	name  string
	enums map[string]map[string]int
}

var protoToAminoTypeMap = map[string]aminoType{
	"/cosmos.bank.v1beta1.MsgSend":          {name: "cosmos-sdk/MsgSend"},
	"/cosmos.gov.v1beta1.MsgSubmitProposal": {name: "cosmos-sdk/MsgSubmitProposal"},
	"/cosmos.gov.v1beta1.MsgDeposit":        {name: "cosmos-sdk/MsgDeposit"},
	"/cosmos.gov.v1beta1.MsgVote": {
		name: "cosmos-sdk/MsgVote",
		enums: map[string]map[string]int{
			"option": {
				"VOTE_OPTION_UNSPECIFIED":  0,
				"VOTE_OPTION_YES":          1,
				"VOTE_OPTION_ABSTAIN":      2,
				"VOTE_OPTION_NO":           3,
				"VOTE_OPTION_NO_WITH_VETO": 4,
			},
		},
	},
	"/cosmos.gov.v1beta1.TextProposal": {name: "cosmos-sdk/TextProposal"},
	"/cosmos.gov.v1.MsgSubmitProposal": {name: "cosmos-sdk/v1/MsgSubmitProposal"},
	"/cosmos.gov.v1.MsgDeposit":        {name: "cosmos-sdk/v1/MsgDeposit"},
	"/cosmos.gov.v1.MsgVote": {
		name: "cosmos-sdk/v1/MsgVote",
		enums: map[string]map[string]int{
			"option": {
				"VOTE_OPTION_UNSPECIFIED":  0,
				"VOTE_OPTION_YES":          1,
				"VOTE_OPTION_ABSTAIN":      2,
				"VOTE_OPTION_NO":           3,
				"VOTE_OPTION_NO_WITH_VETO": 4,
			},
		},
	},
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
func protoToAminoJSON(v any) any {
	// fmt.Fprintf(os.Stderr, "TYPE %s %T\n", v, v)
	x := reflect.ValueOf(v)
	switch x.Kind() {
	case reflect.Map:
		// fmt.Fprintf(os.Stderr, "MAP\n")
		m := maps.Clone(v.(map[string]any))
		// Check if it's a proto @type
		if protoType, ok := m["@type"]; ok {
			aminoType, ok := protoToAminoTypeMap[protoType.(string)]
			if !ok {
				panic(fmt.Sprintf("can't find amino mapping for proto @type='%s'", protoType))
			}
			// fmt.Fprintf(os.Stderr, "@TYPE %v %v\n", protoType, aminoType)
			// remove field @type
			delete(m, "@type")
			// map proto enums if some are configured
			for field, enum := range aminoType.enums {
				val, ok := enum[m[field].(string)]
				// fmt.Fprintln(os.Stderr, "ENUM", field, m[field], val, ok)
				if !ok {
					panic(fmt.Sprintf("can't find enum value for type '%s' and key '%v'", protoType, m[field]))
				}
				m[field] = val
			}
			return map[string]any{
				"type":  aminoType.name,
				"value": protoToAminoJSON(m),
			}
		}
		for k, v := range m {
			// fmt.Fprintf(os.Stderr, "MAP ITEM %s\n", k)
			if isEmptyValue(reflect.ValueOf(v)) {
				// fmt.Fprintf(os.Stderr, "EMPTY\n")
				delete(m, k)
				continue
			}
			m[k] = protoToAminoJSON(v)
		}
		return m
	case reflect.Slice:
		// fmt.Fprintf(os.Stderr, "SLICE\n")
		s := reflect.MakeSlice(x.Type(), x.Len(), x.Cap())
		for i := 0; i < x.Len(); i++ {
			s.Index(i).Set(reflect.ValueOf(protoToAminoJSON(x.Index(i).Interface())))
		}
		return s.Interface()
	default:
		// fmt.Fprintf(os.Stderr, "DEFAULT\n")
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
