package main

import (
	"fmt"
	"maps"
	"reflect"
)

var protoToAminoTypeMap = map[string]string{
	"/cosmos.bank.v1beta1.MsgSend":          "cosmos-sdk/MsgSend",
	"/cosmos.gov.v1beta1.MsgSubmitProposal": "cosmos-sdk/MsgSubmitProposal",
	"/cosmos.gov.v1beta1.MsgDeposit":        "cosmos-sdk/MsgDeposit",
	"/cosmos.gov.v1beta1.TextProposal":      "cosmos-sdk/TextProposal",
	"/cosmos.gov.v1.MsgSubmitProposal":      "cosmos-sdk/v1/MsgSubmitProposal",
	"/cosmos.gov.v1.MsgDeposit":             "cosmos-sdk/v1/MsgDeposit",
	"/govgen.gov.v1beta1.MsgSubmitProposal": "cosmos-sdk/MsgSubmitProposal",
	"/govgen.gov.v1beta1.TextProposal":      "cosmos-sdk/TextProposal",
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
		for k, v := range m {
			// fmt.Fprintf(os.Stderr, "MAP ITEM %s %v\n", k, v)
			if isEmptyValue(reflect.ValueOf(v)) {
				// fmt.Fprintf(os.Stderr, "EMPTY\n")
				delete(m, k)
				continue
			}
			if k == "@type" {
				aminoType, ok := protoToAminoTypeMap[v.(string)]
				if !ok {
					panic(fmt.Sprintf("can't find amino mapping for proto @type=%q", v))
				}
				delete(m, "@type")
				return map[string]any{
					"type":  aminoType,
					"value": protoToAminoJSON(m),
				}
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
