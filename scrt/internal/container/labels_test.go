package container

import "testing"

func TestLabels(t *testing.T) {
	t.Parallel()

	labels := Labels("myproject", "v1.0.0")

	tests := []struct {
		key  string
		want string
	}{
		{LabelAuthor, AuthorValue},
		{LabelProject, "myproject"},
		{LabelVersion, "v1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			got, ok := labels[tt.key]
			if !ok {
				t.Fatalf("label %q not found", tt.key)
			}
			if got != tt.want {
				t.Errorf("label %q = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestFilterLabels(t *testing.T) {
	t.Parallel()

	want := "author=fr3d"
	got := FilterLabels()
	if got != want {
		t.Errorf("FilterLabels() = %q, want %q", got, want)
	}
}
