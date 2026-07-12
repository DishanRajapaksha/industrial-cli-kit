package safety

import "testing"

func TestResolve(t *testing.T) {
	tests := []struct {
		yes, dryRun bool
		want        Mode
		wantErr     bool
	}{
		{want: DryRun},
		{dryRun: true, want: DryRun},
		{yes: true, want: Execute},
		{yes: true, dryRun: true, wantErr: true},
	}
	for _, test := range tests {
		got, err := Resolve(test.yes, test.dryRun)
		if (err != nil) != test.wantErr {
			t.Fatalf("Resolve(%t, %t) error = %v", test.yes, test.dryRun, err)
		}
		if !test.wantErr && got != test.want {
			t.Fatalf("Resolve(%t, %t) = %v, want %v", test.yes, test.dryRun, got, test.want)
		}
	}
}
