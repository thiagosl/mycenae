package keyspace

import (
	"regexp"

	"github.com/uol/mycenae/lib/persistence"
	tlmanager "github.com/uol/timelinemanager"
)

var validKey = regexp.MustCompile(`^[0-9A-Za-z][0-9A-Za-z_]+$`)

// New creates a new keyspace manager
func New(
	timelineManager *tlmanager.Instance,
	storage *persistence.Storage,
	devMode bool,
	defaultTTL int,
	maxAllowedTTL int,
) *Keyspace {
	return &Keyspace{
		Storage:         storage,
		timelineManager: timelineManager,
		devMode:         devMode,
		defaultTTL:      defaultTTL,
		maxAllowedTTL:   maxAllowedTTL,
	}
}

// Keyspace is a structure that represents the functionality of this module
type Keyspace struct {
	*persistence.Storage
	timelineManager *tlmanager.Instance
	devMode         bool
	defaultTTL      int
	maxAllowedTTL   int
}
