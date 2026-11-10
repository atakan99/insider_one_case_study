package content

import "testing"

func TestParseVars(t *testing.T) {
	tests := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"", 0, false},
		{"Hi {{first_name}}", 1, false},
		{"{{first_name}} and {{first_name}}", 2, false},
		{"{{last_name}}", 0, true},
		{"{{ first_name }}", 0, true},
		{"{{first_name}", 0, true},
		{"{{", 0, true},
		{"}}", 0, true},
		{"{first_name}", 0, true},
		{"{{FirstName}}", 0, true},
		{"a { b", 0, true},
		{"a } b", 0, true},
	}
	for _, tc := range tests {
		got, err := ParseVars(tc.in)
		if (err != nil) != tc.wantErr {
			t.Errorf("ParseVars(%q) err = %v, wantErr %v", tc.in, err, tc.wantErr)
			continue
		}
		if err == nil && got != tc.want {
			t.Errorf("ParseVars(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestRender(t *testing.T) {
	tests := []struct{ in, name, want string }{
		{"Hi {{first_name}}!", "Ada", "Hi Ada!"},
		{"{{first_name}} {{first_name}}", "Ada", "Ada Ada"},
		{"No variable", "Ada", "No variable"},
		{"Hi {{first_name}}", "<b>x</b>", "Hi <b>x</b>"},
	}
	for _, tc := range tests {
		if got := Render(tc.in, tc.name); got != tc.want {
			t.Errorf("Render(%q, %q) = %q, want %q", tc.in, tc.name, got, tc.want)
		}
	}
}
