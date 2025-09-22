package migration

import "time"

// Migration represents a database migration
type Migration struct {
	Version   string
	Name      string
	Filepath  string
	Checksum  string
	AppliedAt *time.Time
	IsApplied bool
}

// Config holds migration configuration
type Config struct {
	Enabled bool
	Path    string
	Timeout time.Duration
}
