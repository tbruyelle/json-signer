package main

import (
	"fmt"
	"maps"
	"reflect"
)

var protoToAminoTypeMap = map[string]string{
	"/cosmos.bank.v1beta1.MsgSend":          "cosmos-sdk/MsgSend",
	"/cosmos.gov.v1beta1.MsgSubmitProposal": "cosmos-sdk/MsgSubmitProposal",
	"/cosmos.gov.v1beta1.TextProposal":      "cosmos-sdk/TextProposal",
	"/cosmos.gov.v1.MsgSubmitProposal":      "cosmos-sdk/v1/MsgSubmitProposal",
	"/govgen.gov.v1beta1.MsgSubmitProposal": "cosmos-sdk/MsgSubmitProposal",
	"/govgen.gov.v1beta1.TextProposal":      "cosmos-sdk/TextProposal",
}

// protoToAminoJSON turns proto json to amino json.
// It works by mapping the proto `@type` into amino `type`, and then
// encapsulate the other fields in a amino `value` field.
// TODO add parameter proto-to-amino map to extend the global map.
func protoToAminoJSON(m map[string]any) map[string]any {
	m = maps.Clone(m)
	for k, v := range m {
		if isEmptyValue(reflect.ValueOf(v)) {
			delete(m, k)
		}
	}
	if protoType, ok := m["@type"]; ok {
		aminoType, ok := protoToAminoTypeMap[protoType.(string)]
		if !ok {
			panic(fmt.Sprintf("can't find amino mapping for proto @type=%q", protoType))
		}
		delete(m, "@type")
		return map[string]any{
			"type":  aminoType,
			"value": protoToAminoJSON(m),
		}
	}
	for k, v := range m {
		if mm, ok := v.(map[string]any); ok {
			m[k] = protoToAminoJSON(mm)
		}
	}
	return m
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
