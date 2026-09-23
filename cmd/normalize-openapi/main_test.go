package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNormalizeCollapsesNullableUnions(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "optional header parameter",
			in:   `{"anyOf": [{"type": "string"}, {"type": "null"}], "title": "X-Api-Key"}`,
			want: `{"type": "string", "nullable": true, "title": "X-Api-Key"}`,
		},
		{
			name: "nullable date-time property",
			in:   `{"anyOf": [{"type": "string", "format": "date-time"}, {"type": "null"}], "title": "Expires At"}`,
			want: `{"type": "string", "format": "date-time", "nullable": true, "title": "Expires At"}`,
		},
		{
			name: "nullable reference",
			in:   `{"anyOf": [{"$ref": "#/components/schemas/GovernanceRecordIdentity"}, {"type": "null"}]}`,
			want: `{"$ref": "#/components/schemas/GovernanceRecordIdentity", "nullable": true}`,
		},
		{
			name: "nullable const",
			in:   `{"anyOf": [{"type": "string", "const": "mcp.tool.call"}, {"type": "null"}]}`,
			want: `{"type": "string", "enum": ["mcp.tool.call"], "nullable": true}`,
		},
		{
			name: "nullable oneOf",
			in:   `{"oneOf": [{"type": "integer"}, {"type": "null"}]}`,
			want: `{"type": "integer", "nullable": true}`,
		},
		{
			name: "nullable union of two schemas keeps both",
			in:   `{"anyOf": [{"type": "string"}, {"type": "integer"}, {"type": "null"}], "title": "Policy Version"}`,
			want: `{"anyOf": [{"type": "string"}, {"type": "integer"}], "nullable": true, "title": "Policy Version"}`,
		},
		{
			name: "non-nullable union is unchanged",
			in:   `{"anyOf": [{"type": "string"}, {"type": "integer"}]}`,
			want: `{"anyOf": [{"type": "string"}, {"type": "integer"}]}`,
		},
		{
			name: "nested property",
			in:   `{"type": "object", "properties": {"record": {"anyOf": [{"$ref": "#/components/schemas/GovernanceRecordIdentity"}, {"type": "null"}]}}}`,
			want: `{"type": "object", "properties": {"record": {"$ref": "#/components/schemas/GovernanceRecordIdentity", "nullable": true}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalize(decodeJSON(t, tt.in))
			want := decodeJSON(t, tt.want)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("normalize(%s)\n got: %s\nwant: %s", tt.in, encodeJSON(t, got), tt.want)
			}
		})
	}
}

func decodeJSON(t *testing.T, data string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		t.Fatalf("decode %s: %v", data, err)
	}
	return v
}

func encodeJSON(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	return string(data)
}
