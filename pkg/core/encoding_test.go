package core

import (
	"encoding/hex"
	"math"
	"testing"
)

var testVectors = []struct {
	name string
	val  Value
	hex  string
}{
	{"undefined", U(), "00"},
	{"null", V(nil), "01"},
	{"false", V(false), "02"},
	{"true", V(true), "03"},
	{"num_0", V(0.0), "040000000000000000"},
	{"num_1", V(1.0), "043ff0000000000000"},
	{"num_42", V(42.0), "044045000000000000"},
	{"num_neg1", V(-1.0), "04bff0000000000000"},
	{"num_half", V(0.5), "043fe0000000000000"},
	{"str_empty", V(""), "0500"},
	{"str_a", V("a"), "050161"},
	{"str_42", V("42"), "05023432"},
	{"str_unicode", V("héllo"), "050668c3a96c6c6f"}, // é is 2 UTF-8 bytes => length 6
	{"arr_empty", V([]any{}), "0600"},
	{"arr_nums", V([]any{1.0, 2.0}), "0602043ff0000000000000044000000000000000"},
	{"arr_strs", V([]any{"1", "2"}), "0602050131050132"},
	{"obj_empty", V(map[string]any{}), "0700"},
	{"obj_a1", V(map[string]any{"a": 1.0}), "07010161043ff0000000000000"},
	{"obj_a1_bx", V(map[string]any{"a": 1.0, "b": "x"}), "07020161043ff00000000000000162050178"},
	{"obj_nested", V(map[string]any{"u": map[string]any{"id": "x"}, "t": []any{"a"}}), "07020174060105016101750701026964050178"},
}

func TestCanonicalBytesOnTestVectors(t *testing.T) {
	for _, tv := range testVectors {
		got, err := tv.val.ToCanonicalBytes()
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tv.name, err)
			continue
		}
		if h := hex.EncodeToString(got); h != tv.hex {
			t.Errorf("%s: encoding drifted\n  got  %s\n  want %s", tv.name, h, tv.hex)
		}
	}
}

func TestCanonicalBytesInjective(t *testing.T) {
	pairs := []struct {
		name string
		a, b Value
	}{
		{"number_vs_string", V(42.0), V("42")},
		{"bool_vs_string", V(true), V("true")},
		{"null_vs_string", V(nil), V("null")},
		{"null_vs_undefined", V(nil), U()},
		{"array_vs_string", V([]any{"1", "2"}), V("[1 2]")},
		{"object_vs_string", V(map[string]any{"a": 1.0, "b": "x"}), V("map[a:1 b:x]")},
		{"empty_array_vs_object", V([]any{}), V(map[string]any{})},
		// object payload injectivity: {a:"1 b:x"} must not equal {a:1, b:"x"}
		{"object_field_forgery", V(map[string]any{"a": "1 b:x"}), V(map[string]any{"a": 1.0, "b": "x"})},
		// length-prefix boundary: ["a","bc"] must not equal ["ab","c"]
		{"array_boundary_forgery", V([]any{"a", "bc"}), V([]any{"ab", "c"})},
	}
	for _, p := range pairs {
		ea, err := p.a.ToCanonicalBytes()
		if err != nil {
			t.Fatalf("%s: a: %v", p.name, err)
		}
		eb, err := p.b.ToCanonicalBytes()
		if err != nil {
			t.Fatalf("%s: b: %v", p.name, err)
		}
		if hex.EncodeToString(ea) == hex.EncodeToString(eb) {
			t.Errorf("%s: distinct values share an encoding: %s", p.name, hex.EncodeToString(ea))
		}
	}
}

func TestCanonicalBytesEqual(t *testing.T) {
	z := 0.0
	pairs := []struct {
		name string
		a, b Value
	}{
		{"neg_zero", V(0.0), V(-z)},
		{"nan_payloads", V(math.NaN()), V(math.Float64frombits(0x7FF8000000000042))},
		{"object_key_order", V(ValueMap{"a": V(1.0), "b": V(2.0)}), V(ValueMap{"b": V(2.0), "a": V(1.0)})},
	}
	for _, p := range pairs {
		ea, err := p.a.ToCanonicalBytes()
		if err != nil {
			t.Fatalf("%s: a: %v", p.name, err)
		}
		eb, err := p.b.ToCanonicalBytes()
		if err != nil {
			t.Fatalf("%s: b: %v", p.name, err)
		}
		if hex.EncodeToString(ea) != hex.EncodeToString(eb) {
			t.Errorf("%s: equal values encode differently:\n  a %s\n  b %s",
				p.name, hex.EncodeToString(ea), hex.EncodeToString(eb))
		}
	}
}

func TestCanonicalBytesDeterministic(t *testing.T) {
	v := V(map[string]any{
		"user": map[string]any{"id": "u1", "score": 3.5},
		"tags": []any{"a", "b", "c"},
	})
	first, err := v.ToCanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	for i := range 10 {
		got, err := v.ToCanonicalBytes()
		if err != nil {
			t.Fatal(err)
		}
		if hex.EncodeToString(got) != hex.EncodeToString(first) {
			t.Fatalf("non-deterministic encoding on iteration %d", i)
		}
	}
}
