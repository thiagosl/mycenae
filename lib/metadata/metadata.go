package metadata

import (
	"github.com/uol/gobol"
	"github.com/uol/gobol/solar"
	"github.com/uol/logh"
	"github.com/uol/mycenae/lib/constants"
	"github.com/uol/mycenae/lib/memcached"
	tlmanager "github.com/uol/timelinemanager"
)

// Backend hides the underlying implementation of the metadata storage
type Backend interface {
	// CreateKeyset creates a keyset in the metadata storage
	CreateKeyset(name string) gobol.Error

	// DeleteKeyset deletes a keyset in the metadata storage
	DeleteKeyset(name string) gobol.Error

	// ListKeyset - list all keyset
	ListKeysets() []string

	// CheckKeyset - verifies if a keyset exists
	CheckKeyset(keyset string) bool

	// FilterTagValues - filter tag values from a collection
	FilterTagValues(collection, prefix string, maxResults int) ([]string, int, gobol.Error)

	// FilterTagKeys - filter tag keys from a collection
	FilterTagKeys(collection, prefix string, maxResults int) ([]string, int, gobol.Error)

	// FilterMetrics - filter metrics from a collection
	FilterMetrics(collection, prefix string, maxResults int) ([]string, int, gobol.Error)

	// FilterMetadata - list all metas from a collection
	// Returns: results, total and gobol.Error
	FilterMetadata(collection string, query *Query, from, maxResults int) ([]Metadata, int, gobol.Error)

	// AddDocument - add/update a document
	AddDocument(collection string, metadata *Metadata) gobol.Error

	// CheckMetadata - verifies if a metadata exists
	CheckMetadata(collection, tsType, tsid string, tsidBytes []byte) (bool, gobol.Error)

	// SetRegexValue - add slashes to the value
	SetRegexValue(value string) string

	// HasRegexPattern - check if the value has a regular expression
	HasRegexPattern(value string) bool

	// DeleteDocumentByID - delete a document by ID and its child documents
	DeleteDocumentByID(collection, tsType, id string) gobol.Error

	// FilterTagKeysByMetric - filter tag values from a collection given its metric
	FilterTagKeysByMetric(collection, tsType, metric, prefix string, maxResults int) ([]string, int, gobol.Error)

	// FilterTagValuesByMetricAndTag - filter tag values from a collection given its metric and tag
	FilterTagValuesByMetricAndTag(collection, tsType, metric, tag, prefix string, maxResults int) ([]string, int, gobol.Error)
}

// Storage is a storage for metadata
type Storage struct {
	logger *logh.ContextualLogger

	// Backend is the thing that actually does the specific work in the storage
	Backend
}

// Settings for the metadata package
type Settings struct {
	NumShards                     int
	ReplicationFactor             int
	IDCacheTTL                    int
	QueryCacheTTL                 int
	MaxReturnedMetadata           int
	ZookeeperConfig               string
	BlacklistedKeysets            []string
	CacheKeyHashSize              int
	KeysetCacheAutoUpdateInterval string
	solar.Configuration
}

// Metadata document
type Metadata struct {
	ID       string   `json:"id"`
	Metric   string   `json:"metric"`
	TagKey   []string `json:"tagKey"`
	TagValue []string `json:"tagValue"`
	MetaType string   `json:"type"`
	Keyset   string   `json:"keyset"`
}

// Query - query
type Query struct {
	Metric   string     `json:"metric"`
	MetaType string     `json:"type"`
	Regexp   bool       `json:regexp`
	Tags     []QueryTag `json:"tags"`
}

// QueryTag - tags for query
type QueryTag struct {
	Key    string   `json:"key"`
	Values []string `json:value`
	Negate bool     `json:negate`
	Regexp bool     `json:regexp`
}

// Create creates a metadata handler
func Create(settings *Settings, mc *tlmanager.Instance, memcached *memcached.Memcached) (*Storage, error) {

	backend, err := NewSolrBackend(settings, mc, memcached)
	if err != nil {
		return nil, err
	}

	return &Storage{
		logger:  logh.CreateContextualLogger(constants.StringsPKG, "metadata"),
		Backend: backend,
	}, nil
}
