// SPDX-License-Identifier: EUPL-1.2

// JSON helpers for the Core framework.
// Wraps encoding/json so consumers don't import it directly.
// Same guardrail pattern as string.go wraps strings.
//
// Usage:
//
//	data := core.JSONMarshal(myStruct)
//	if data.OK { json := data.Value.([]byte) }
//
//	r := core.JSONUnmarshal(jsonBytes, &target)
//	if !r.OK { /* handle error */ }
package core

import "encoding/json"

// RawMessage is an alias for json.RawMessage — a deferred-decode JSON
// fragment. Lets HTTP / MCP / IPC handlers accept JSON parameters
// without committing to a concrete struct, then decode when the shape
// is known.
//
//	func HandleBridgeCall(args core.RawMessage) core.Result {
//	    var req DeployRequest
//	    if r := core.JSONUnmarshal(args, &req); !r.OK { return r }
//	    return core.Ok(req)
//	}
type RawMessage = json.RawMessage

// JSONMarshal serialises a value to JSON bytes.
//
//	r := core.JSONMarshal(myStruct)
//	if r.OK { data := r.Value.([]byte) }
func JSONMarshal(v any) Result {
	data, err := json.Marshal(v)
	if err != nil {
		return Result{err, false}
	}
	return Result{data, true}
}

// JSONMarshalIndent serialises a value to indented JSON bytes.
//
//	r := core.JSONMarshalIndent(report, "", "  ")
//	if r.OK { data := r.Value.([]byte) }
func JSONMarshalIndent(v any, prefix, indent string) Result {
	data, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		return Result{err, false}
	}
	return Result{data, true}
}

// JSONMarshalString serialises a value to a JSON string.
//
//	s := core.JSONMarshalString(myStruct)
func JSONMarshalString(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// JSONUnmarshal deserialises JSON bytes into a target.
//
//	var cfg Config
//	r := core.JSONUnmarshal(data, &cfg)
func JSONUnmarshal(data []byte, target any) Result {
	if err := json.Unmarshal(data, target); err != nil {
		return Result{err, false}
	}
	return Result{OK: true}
}

// JSONUnmarshalString deserialises a JSON string into a target.
//
//	var cfg Config
//	r := core.JSONUnmarshalString(`{"port":8080}`, &cfg)
func JSONUnmarshalString(s string, target any) Result {
	return JSONUnmarshal([]byte(s), target)
}
