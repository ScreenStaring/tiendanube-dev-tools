package metafields

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/ScreenStaring/tiendanube-dev-tools/api"
)

func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestParseIDArgsOnly(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    []int64
		wantErr bool
	}{
		{
			name: "single id",
			args: []string{"123"},
			want: []int64{123},
		},
		{
			name: "multiple ids",
			args: []string{"123", "456"},
			want: []int64{123, 456},
		},
		{
			name: "empty",
			args: nil,
			want: []int64{},
		},
		{
			name:    "invalid id",
			args:    []string{"abc"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIDArgsOnly(tt.args, "Location")
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseIDArgsOnly(%v) = %v, want error", tt.args, got)
				}
				return
			}
			if err != nil {
				t.Errorf("parseIDArgsOnly(%v) unexpected error: %v", tt.args, err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseIDArgsOnly(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestParseMetafieldID(t *testing.T) {
	tests := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"12345", 12345, false},
		{"12345@custom.details", 12345, false},
		{" 9876 @ns.key ", 9876, false},
		{"invalid", 0, true},
		{"@ns.key", 0, true},
	}

	for _, tt := range tests {
		got, err := parseMetafieldID(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseMetafieldID(%q) = %d, want error", tt.input, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseMetafieldID(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseMetafieldID(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestReverseMetafields(t *testing.T) {
	mfs := []api.Metafield{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}

	reverseMetafields(mfs)

	if mfs[0].ID != 3 || mfs[1].ID != 2 || mfs[2].ID != 1 {
		t.Errorf("unexpected reversed order: %+v", mfs)
	}
}

func TestPrintMetafieldsFormatted(t *testing.T) {
	mfs := []api.Metafield{
		{
			ID:          101,
			Namespace:   "details",
			Key:         "fabric",
			Value:       "Cotton",
			Description: "Fabric material",
			CreatedAt:   "2026-01-01T00:00:00Z",
			UpdatedAt:   "2026-01-01T00:00:00Z",
		},
	}

	out := captureOutput(func() {
		printFormatted(mfs)
	})

	if !strings.Contains(out, "details") || !strings.Contains(out, "fabric") || !strings.Contains(out, "Cotton") {
		t.Errorf("unexpected output: %s", out)
	}
	if !strings.Contains(out, "--------------------") {
		t.Errorf("expected separator: %s", out)
	}
}
