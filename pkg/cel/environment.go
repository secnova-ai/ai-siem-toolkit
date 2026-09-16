// Package cel implements the Tool Center authoring expression environment.
package cel

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	celgo "github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"reflect"
)

func newCELEnv(exec *httpExecutor) (*celgo.Env, error) {
	httpResult := celgo.MapType(celgo.StringType, celgo.DynType)

	return celgo.NewEnv(
		celgo.Variable("params", celgo.MapType(celgo.StringType, celgo.DynType)),
		celgo.Variable("creds", celgo.MapType(celgo.StringType, celgo.DynType)),

		// get(url: string) → {body: string, status: int, ok: bool}
		celgo.Function("get",
			celgo.Overload("get_string", []*celgo.Type{celgo.StringType}, httpResult,
				celgo.UnaryBinding(func(v ref.Val) ref.Val {
					rawURL, ok := v.Value().(string)
					if !ok {
						return types.NewErr("get: expected string URL")
					}
					m, err := exec.get(rawURL)
					if err != nil {
						return types.NewErr("get: %v", err)
					}
					return nativeToVal(m)
				}),
			),
		),

		// post(url: string, body: string) → {body, status, ok}
		celgo.Function("post",
			celgo.Overload("post_string_string", []*celgo.Type{celgo.StringType, celgo.StringType}, httpResult,
				celgo.BinaryBinding(func(u, b ref.Val) ref.Val {
					rawURL, ok1 := u.Value().(string)
					body, ok2 := b.Value().(string)
					if !ok1 || !ok2 {
						return types.NewErr("post: expected string args")
					}
					m, err := exec.post(rawURL, body)
					if err != nil {
						return types.NewErr("post: %v", err)
					}
					return nativeToVal(m)
				}),
			),
		),

		// put(url: string, body: string) → {body, status, ok}
		celgo.Function("put",
			celgo.Overload("put_string_string", []*celgo.Type{celgo.StringType, celgo.StringType}, httpResult,
				celgo.BinaryBinding(func(u, b ref.Val) ref.Val {
					rawURL, ok1 := u.Value().(string)
					body, ok2 := b.Value().(string)
					if !ok1 || !ok2 {
						return types.NewErr("put: expected string args")
					}
					m, err := exec.put(rawURL, body)
					if err != nil {
						return types.NewErr("put: %v", err)
					}
					return nativeToVal(m)
				}),
			),
		),

		// delete(url: string) → {body, status, ok}
		celgo.Function("delete",
			celgo.Overload("delete_string", []*celgo.Type{celgo.StringType}, httpResult,
				celgo.UnaryBinding(func(v ref.Val) ref.Val {
					rawURL, ok := v.Value().(string)
					if !ok {
						return types.NewErr("delete: expected string URL")
					}
					m, err := exec.delete(rawURL)
					if err != nil {
						return types.NewErr("delete: %v", err)
					}
					return nativeToVal(m)
				}),
			),
		),

		// get_h(url: string, headers: map<string, dyn>) → {body, status, ok}
		celgo.Function("get_h",
			celgo.Overload("get_h_string_map", []*celgo.Type{celgo.StringType, celgo.MapType(celgo.StringType, celgo.DynType)}, httpResult,
				celgo.BinaryBinding(func(u, h ref.Val) ref.Val {
					rawURL, ok := u.Value().(string)
					if !ok {
						return types.NewErr("get_h: expected string URL")
					}
					m, err := exec.getH(rawURL, celValToMap(h))
					if err != nil {
						return types.NewErr("get_h: %v", err)
					}
					return nativeToVal(m)
				}),
			),
		),

		// post_h(url: string, body: string, headers: map<string, dyn>) → {body, status, ok}
		celgo.Function("post_h",
			celgo.Overload("post_h_string_string_map", []*celgo.Type{celgo.StringType, celgo.StringType, celgo.MapType(celgo.StringType, celgo.DynType)}, httpResult,
				celgo.FunctionBinding(func(args ...ref.Val) ref.Val {
					rawURL, ok1 := args[0].Value().(string)
					body, ok2 := args[1].Value().(string)
					if !ok1 || !ok2 {
						return types.NewErr("post_h: expected string args")
					}
					m, err := exec.postH(rawURL, body, celValToMap(args[2]))
					if err != nil {
						return types.NewErr("post_h: %v", err)
					}
					return nativeToVal(m)
				}),
			),
		),

		// put_h(url: string, body: string, headers: map<string, dyn>) → {body, status, ok}
		celgo.Function("put_h",
			celgo.Overload("put_h_string_string_map", []*celgo.Type{celgo.StringType, celgo.StringType, celgo.MapType(celgo.StringType, celgo.DynType)}, httpResult,
				celgo.FunctionBinding(func(args ...ref.Val) ref.Val {
					rawURL, ok1 := args[0].Value().(string)
					body, ok2 := args[1].Value().(string)
					if !ok1 || !ok2 {
						return types.NewErr("put_h: expected string args")
					}
					m, err := exec.putH(rawURL, body, celValToMap(args[2]))
					if err != nil {
						return types.NewErr("put_h: %v", err)
					}
					return nativeToVal(m)
				}),
			),
		),

		// delete_h(url: string, headers: map<string, dyn>) → {body, status, ok}
		celgo.Function("delete_h",
			celgo.Overload("delete_h_string_map", []*celgo.Type{celgo.StringType, celgo.MapType(celgo.StringType, celgo.DynType)}, httpResult,
				celgo.BinaryBinding(func(u, h ref.Val) ref.Val {
					rawURL, ok := u.Value().(string)
					if !ok {
						return types.NewErr("delete_h: expected string URL")
					}
					m, err := exec.deleteH(rawURL, celValToMap(h))
					if err != nil {
						return types.NewErr("delete_h: %v", err)
					}
					return nativeToVal(m)
				}),
			),
		),

		// x.encode_json() → string   (member function, works on any type)
		celgo.Function("encode_json",
			celgo.MemberOverload("dyn_encode_json", []*celgo.Type{celgo.DynType}, celgo.StringType,
				celgo.UnaryBinding(func(v ref.Val) ref.Val {
					native, err := v.ConvertToNative(reflect.TypeOf((*interface{})(nil)).Elem())
					if err != nil {
						return types.NewErr("encode_json: convert: %v", err)
					}
					b, err := json.Marshal(normalizeForJSON(native))
					if err != nil {
						return types.NewErr("encode_json: marshal: %v", err)
					}
					return types.String(string(b))
				}),
			),
		),

		// decode_json — available both as a free function and as a member call.
		//   decode_json(s)   → works on string literals
		//   x.decode_json()  → works on dyn (e.g. get(url).body.decode_json())
		celgo.Function("decode_json",
			celgo.Overload("decode_json_string",
				[]*celgo.Type{celgo.StringType}, celgo.DynType,
				celgo.UnaryBinding(decodeJSONVal),
			),
			celgo.MemberOverload("decode_json_dyn_member",
				[]*celgo.Type{celgo.DynType}, celgo.DynType,
				celgo.UnaryBinding(decodeJSONVal),
			),
		),

		// base64_encode(s: string) → string   (e.g. Basic-auth header construction)
		celgo.Function("base64_encode",
			celgo.Overload("base64_encode_string",
				[]*celgo.Type{celgo.StringType}, celgo.StringType,
				celgo.UnaryBinding(func(v ref.Val) ref.Val {
					s, ok := v.Value().(string)
					if !ok {
						return types.NewErr("base64_encode: expected string")
					}
					return types.String(base64.StdEncoding.EncodeToString([]byte(s)))
				}),
			),
		),
	)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// decodeJSONVal is the shared binding for the decode_json CEL function.
// It accepts string or dyn (e.g. result of map field access).
func decodeJSONVal(v ref.Val) ref.Val {
	var s string
	switch val := v.Value().(type) {
	case string:
		s = val
	case []byte:
		s = string(val)
	default:
		// dyn case: convert to native then to string
		native, err := v.ConvertToNative(reflect.TypeOf((*interface{})(nil)).Elem())
		if err != nil {
			return types.NewErr("decode_json: convert: %v", err)
		}
		switch sv := native.(type) {
		case string:
			s = sv
		default:
			return types.NewErr("decode_json: expected string, got %T", native)
		}
	}
	var result interface{}
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return types.NewErr("decode_json: %v", err)
	}
	return nativeToVal(result)
}

// nativeToVal converts a native Go value to a CEL ref.Val.
func nativeToVal(v any) ref.Val {
	return types.DefaultTypeAdapter.NativeToValue(v)
}

// celValToMap converts a CEL map ref.Val to map[string]any.
// Returns nil if conversion fails.
func celValToMap(v ref.Val) map[string]any {
	native, err := v.ConvertToNative(reflect.TypeOf((*interface{})(nil)).Elem())
	if err != nil {
		return nil
	}
	if m, ok := normalizeForJSON(native).(map[string]interface{}); ok {
		return m
	}
	return nil
}

// valToJSON converts a CEL ref.Val to JSON.
func valToJSON(v ref.Val) (json.RawMessage, error) {
	if v == nil || v == types.NullValue {
		return json.RawMessage("null"), nil
	}
	native, err := v.ConvertToNative(reflect.TypeOf((*interface{})(nil)).Elem())
	if err != nil {
		// Fallback: marshal the raw Value()
		b, err2 := json.Marshal(v.Value())
		if err2 != nil {
			return json.RawMessage("null"), nil
		}
		return b, nil
	}
	b, err := json.Marshal(normalizeForJSON(native))
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return b, nil
}

// normalizeForJSON recursively converts map[interface{}]interface{} (CEL's
// internal map representation) to map[string]interface{} so json.Marshal works.
func normalizeForJSON(v interface{}) interface{} {
	switch vt := v.(type) {
	case map[interface{}]interface{}:
		out := make(map[string]interface{}, len(vt))
		for k, val := range vt {
			out[fmt.Sprintf("%v", k)] = normalizeForJSON(val)
		}
		return out
	case map[string]interface{}:
		for k, val := range vt {
			vt[k] = normalizeForJSON(val)
		}
		return vt
	case []interface{}:
		for i, elem := range vt {
			vt[i] = normalizeForJSON(elem)
		}
		return vt
	default:
		return v
	}
}
