package sqlite

import "testing"

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		pragma string
		want   string
	}{
		{
			name:   "simple path without pragma",
			path:   "db/data/forum.db",
			pragma: "",
			want:   "db/data/forum.db",
		},
		{
			name:   "path with pragma uses question mark",
			path:   "db/data/forum.db",
			pragma: "_foreign_keys=on&_journal_mode=WAL",
			want:   "db/data/forum.db?_foreign_keys=on&_journal_mode=WAL",
		},
		{
			name:   "path with existing query params uses ampersand",
			path:   "file:db/data/forum.db?mode=memory",
			pragma: "_foreign_keys=on",
			want:   "file:db/data/forum.db?mode=memory&_foreign_keys=on",
		},
		{
			name:   "empty pragma returns path unchanged",
			path:   "db/data/forum.db",
			pragma: "",
			want:   "db/data/forum.db",
		},
		{
			name:   "path with trailing question mark",
			path:   "db/data/forum.db?",
			pragma: "_journal_mode=WAL",
			want:   "db/data/forum.db?&_journal_mode=WAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDSN(tt.path, tt.pragma)
			if got != tt.want {
				t.Errorf("buildDSN(%q, %q) = %q, want %q", tt.path, tt.pragma, got, tt.want)
			}
		})
	}
}
