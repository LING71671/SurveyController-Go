package answer

import (
	"reflect"
	"testing"
)

func TestMergeRequiredOptions(t *testing.T) {
	got := MergeRequiredOptions(
		[]string{"b", "a", "b", ""},
		[]string{"c", "a", ""},
	)
	want := SelectionResult{OptionIDs: []string{"a", "b", "c"}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MergeRequiredOptions() = %+v, want %+v", got, want)
	}
}
