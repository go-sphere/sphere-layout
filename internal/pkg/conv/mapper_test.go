package conv

import (
	"bytes"
	"testing"
)

func TestMapStruct(t *testing.T) {
	type source struct {
		Name     string
		Age      string
		Raw      []byte
		Internal struct{ Name string }
	}
	type target struct {
		Name     *string
		Age      int
		Raw      []byte
		Internal *struct{ Name string }
	}

	if got := MapStruct[source, target](nil); got != nil {
		t.Fatalf("MapStruct(nil) = %#v, want nil", got)
	}

	input := source{Name: "Alice", Age: "25", Raw: []byte("raw")}
	input.Internal.Name = "internal"
	got := MapStruct[source, target](&input)
	if got == nil {
		t.Fatal("MapStruct() = nil")
	}
	if got.Name == nil || *got.Name != input.Name {
		t.Errorf("Name = %v, want %q", got.Name, input.Name)
	}
	if got.Age != 25 {
		t.Errorf("Age = %d, want 25", got.Age)
	}
	if !bytes.Equal(got.Raw, input.Raw) {
		t.Errorf("Raw = %q, want %q", got.Raw, input.Raw)
	}
	if got.Internal == nil || got.Internal.Name != input.Internal.Name {
		t.Errorf("Internal = %#v, want Name %q", got.Internal, input.Internal.Name)
	}
}
