package instrumentation

import "testing"

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Level
		wantErr bool
	}{
		{
			name:  "empty defaults to none",
			input: "",
			want:  None,
		},
		{
			name:  "none",
			input: "none",
			want:  None,
		},
		{
			name:  "low",
			input: "low",
			want:  Low,
		},
		{
			name:  "medium",
			input: "medium",
			want:  Medium,
		},
		{
			name:  "high",
			input: "high",
			want:  High,
		},
		{
			name:  "critical",
			input: "critical",
			want:  Critical,
		},
		{
			name:  "case insensitive",
			input: "MeDiUm",
			want:  Medium,
		},
		{
			name:    "invalid",
			input:   "verbose",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLevel(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
