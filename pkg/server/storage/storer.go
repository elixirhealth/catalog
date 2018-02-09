package storage

import (
	"time"

	"errors"

	"fmt"

	"github.com/drausin/libri/libri/common/id"
	libriapi "github.com/drausin/libri/libri/librarian/api"
	api "github.com/elxirhealth/catalog/pkg/catalogapi"
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

	// ErrEarlierBeforeMin denotes when the before time filter is before the minimum date.
	ErrEarlierBeforeMin = fmt.Errorf("before time filter earlier than %s",
		minBeforeTime.String())

	defaultSearchQueryTimeout = 5 * time.Second

	minBeforeTime = time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
)

// Storer stores and searches for *PublicationReceipts.
type Storer interface {
	Put(pub *api.PublicationReceipt) error
	Search(filters *SearchFilters, limit uint) ([]*api.PublicationReceipt, error)
}

// Parameters defines the parameters of the Storer.
type Parameters struct {
	SearchQueryTimeout time.Duration
}

// NewDefaultParameters returns a *Parameters object with default values.
func NewDefaultParameters() *Parameters {
	return &Parameters{
		SearchQueryTimeout: defaultSearchQueryTimeout,
	}
}

// SearchFilters represents a set of filters for a search. If fields are non-null, an equality
// constraint with its value is added to the search. The Before field is an epoch timestamp (
// micro-seconds since Jan 1, 1970) and denotes an exclusive bound.
type SearchFilters struct {
	EntryKey        []byte
	AuthorPublicKey []byte
	ReaderPublicKey []byte
	Before          int64
}

func validateSearchFilters(f *SearchFilters) error {
	if f.Before != 0 && api.FromEpochMicros(f.Before).Before(minBeforeTime) {
		return ErrEarlierBeforeMin
	}
	hasFilters := false
	if f.EntryKey != nil {
		hasFilters = true
		if len(f.EntryKey) != id.Length {
			return ErrUnexpectedEntryKeyLength
		}
	}
	if f.AuthorPublicKey != nil {
		hasFilters = true
		if len(f.AuthorPublicKey) != libriapi.ECPubKeyLength {
			return ErrUnexpectedAuthorPubKeyLength
		}
	}
	if f.ReaderPublicKey != nil {
		hasFilters = true
		if len(f.ReaderPublicKey) != libriapi.ECPubKeyLength {
			return ErrUnexpectedReaderPubKeyLength
		}
	}
	if !hasFilters {
		return ErrNoEqualityFilters
	}
	return nil
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
