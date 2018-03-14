package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/drausin/libri/libri/common/id"
	libriapi "github.com/drausin/libri/libri/librarian/api"
	api "github.com/elxirhealth/catalog/pkg/catalogapi"
	bstorage "github.com/elxirhealth/service-base/pkg/server/storage"
	"go.uber.org/zap/zapcore"
)

var (
	// ErrNoEqualityFilters denotes when search filters have no equality filters.
	ErrNoEqualityFilters = errors.New("no equality search filters defined")

	// ErrUnexpectedEntryKeyLength denotes when the entry key filter has an unexpected length.
	ErrUnexpectedEntryKeyLength = errors.New("unexpected entry key filter length")

	// ErrUnexpectedAuthorPubKeyLength denotes when the author public key filter has an
	// unexpected length.
	ErrUnexpectedAuthorPubKeyLength = errors.New("unexpected author public key filter length")

	// ErrUnexpectedReaderPubKeyLength denotes when the reader public key filter has an
	// unexpected length.
	ErrUnexpectedReaderPubKeyLength = errors.New("unexpected author public key filter length")

	// ErrEarlierTimeMin denotes when a time filter is before the minimum date.
	ErrEarlierTimeMin = fmt.Errorf("time filter earlier than %s", minFilterTime.String())

	// DefaultStorage is the default storage type.
	DefaultStorage = bstorage.Memory

	// DefaultSearchQueryTimeout is the default timeout for search queries.
	DefaultSearchQueryTimeout = 3 * time.Second

	// DefaultTimeout is the default timeout for DataStore operations (e.g., Get, Put).
	DefaultTimeout = 1 * time.Second

	minFilterTime = time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
)

// Storer stores and searches for *PublicationReceipts.
type Storer interface {
	Put(pub *api.PublicationReceipt) error
	Search(filters *SearchFilters, limit uint32) ([]*api.PublicationReceipt, error)
}

// Parameters defines the parameters of the Storer.
type Parameters struct {
	Type               bstorage.Type
	SearchQueryTimeout time.Duration
	GetTimeout         time.Duration
	PutTimeout         time.Duration
}

// NewDefaultParameters returns a *Parameters object with default values.
func NewDefaultParameters() *Parameters {
	return &Parameters{
		Type:               DefaultStorage,
		SearchQueryTimeout: DefaultSearchQueryTimeout,
		GetTimeout:         DefaultTimeout,
		PutTimeout:         DefaultTimeout,
	}
}

// MarshalLogObject writes the parameters to the given object encoder.
func (p *Parameters) MarshalLogObject(oe zapcore.ObjectEncoder) error {
	oe.AddString(logStorageType, p.Type.String())
	oe.AddDuration(logSearchQueryTimeout, p.SearchQueryTimeout)
	return nil
}

// SearchFilters represents a set of filters for a search. If fields are non-null/non-empty, an
// equality constraint with its value is added to the search. The Before field is an epoch
// timestamp (micro-seconds since Jan 1, 1970) and denotes an exclusive bound.
type SearchFilters struct {
	EntryKey        []byte
	AuthorPublicKey []byte
	AuthorEntityID  string
	ReaderPublicKey []byte
	ReaderEntityID  string
	After           int64
	Before          int64
}

func validateSearchFilters(f *SearchFilters) error {
	if !validateTimeFilter(f.Before) || !validateTimeFilter(f.After) {
		return ErrEarlierTimeMin
	}
	hasFilts := f.AuthorEntityID != "" || f.ReaderEntityID != ""
	hasFilts, valid := validateBytesFilters(f.EntryKey, id.Length, hasFilts)
	if !valid {
		return ErrUnexpectedEntryKeyLength
	}

	hasFilts, valid = validateBytesFilters(f.AuthorPublicKey, libriapi.ECPubKeyLength, hasFilts)
	if !valid {
		return ErrUnexpectedAuthorPubKeyLength
	}

	hasFilts, valid = validateBytesFilters(f.ReaderPublicKey, libriapi.ECPubKeyLength, hasFilts)
	if !valid {
		return ErrUnexpectedReaderPubKeyLength
	}

	if !hasFilts {
		return ErrNoEqualityFilters
	}
	return nil
}

func validateTimeFilter(filter int64) (valid bool) {
	return filter == 0 || api.FromEpochMicros(filter).After(minFilterTime)
}

func validateBytesFilters(filter []byte, length int, hasFilters bool) (bool, bool) {
	if filter != nil {
		if len(filter) != length {
			return true, false
		}
		return true, true
	}
	return hasFilters, true
}

// publicationReceipts is a min-heap of PublicationReceipt objects sorted ascending by ReceivedTime
type publicationReceipts []*PublicationReceipt

func (prs publicationReceipts) Len() int {
	return len(prs)
}

func (prs publicationReceipts) Less(i, j int) bool {
	return prs[i].ReceivedTime.Before(prs[j].ReceivedTime)
}

func (prs publicationReceipts) Swap(i, j int) {
	prs[i], prs[j] = prs[j], prs[i]
}

func (prs *publicationReceipts) Push(x interface{}) {
	*prs = append(*prs, x.(*PublicationReceipt))
}

func (prs *publicationReceipts) Pop() interface{} {
	old := *prs
	n := len(old)
	x := old[n-1]
	*prs = old[0 : n-1]
	return x
}

func (prs publicationReceipts) Peak() *PublicationReceipt {
	return prs[0]
}
