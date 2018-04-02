package storage

import (
	"container/heap"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/drausin/libri/libri/common/id"
	libriapi "github.com/drausin/libri/libri/librarian/api"
	api "github.com/elixirhealth/catalog/pkg/catalogapi"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

const (
	// MaxSearchLimit is the maximum number of search results that can be returned in a single
	// search query.
	MaxSearchLimit = uint32(128)
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

	// ErrSearchLimitTooLarge denotes when the search limit given is larger than MaxSearchLimit.
	ErrSearchLimitTooLarge = fmt.Errorf("search limit larger than maximum value %d",
		MaxSearchLimit)

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
	Type          bstorage.Type
	SearchTimeout time.Duration
	GetTimeout    time.Duration
	PutTimeout    time.Duration
}

// NewDefaultParameters returns a *Parameters object with default values.
func NewDefaultParameters() *Parameters {
	return &Parameters{
		Type:          DefaultStorage,
		SearchTimeout: DefaultSearchQueryTimeout,
		GetTimeout:    DefaultTimeout,
		PutTimeout:    DefaultTimeout,
	}
}

// MarshalLogObject writes the parameters to the given object encoder.
func (p *Parameters) MarshalLogObject(oe zapcore.ObjectEncoder) error {
	oe.AddString(logStorageType, p.Type.String())
	oe.AddDuration(logSearchQueryTimeout, p.SearchTimeout)
	oe.AddDuration(logPutTimeout, p.PutTimeout)
	oe.AddDuration(logGetTimeout, p.GetTimeout)
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

// ValidateSearchFilters checks that the populated search filters are valid.
func ValidateSearchFilters(f *SearchFilters) error {
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

// PublicationReceipts is a min-heap of PublicationReceipt objects sorted ascending by ReceivedTime
type PublicationReceipts []*api.PublicationReceipt

// Len returns the number of publication receipts.
func (prs PublicationReceipts) Len() int {
	return len(prs)
}

// Less returns whether pub receipt i was received before receipt j.
func (prs PublicationReceipts) Less(i, j int) bool {
	return prs[i].ReceivedTime < prs[j].ReceivedTime
}

// Swap swaps pub receipts i & j.
func (prs PublicationReceipts) Swap(i, j int) {
	prs[i], prs[j] = prs[j], prs[i]
}

// Push adds the pub receipt to the heap.
func (prs *PublicationReceipts) Push(x interface{}) {
	*prs = append(*prs, x.(*api.PublicationReceipt))
}

// Pop removes and returns the pub receipt from the root of the heap.
func (prs *PublicationReceipts) Pop() interface{} {
	old := *prs
	n := len(old)
	x := old[n-1]
	*prs = old[0 : n-1]
	return x
}

// Peak returns the pub receipt from the root of the heap.
func (prs PublicationReceipts) Peak() *api.PublicationReceipt {
	return prs[0]
}

// PopList pops all the pub receipts from the heap and returns them as a slice.
func (prs *PublicationReceipts) PopList() []*api.PublicationReceipt {
	result := make([]*api.PublicationReceipt, prs.Len())
	for i := prs.Len() - 1; i >= 0; i-- {
		result[i] = heap.Pop(prs).(*api.PublicationReceipt)
	}
	return result
}

// TestStorerPutSearch tests that the storer properly handles a number of Put
func TestStorerPutSearch(t *testing.T, s Storer) {
	now := time.Now().Unix() * 1E6
	envKey1 := append(make([]byte, id.Length-1), 1)
	envKey2 := append(make([]byte, id.Length-1), 2)
	envKey3 := append(make([]byte, id.Length-1), 3)
	envKey4 := append(make([]byte, id.Length-1), 4)
	entryKey1 := append(make([]byte, id.Length-1), 5)
	entryKey2 := append(make([]byte, id.Length-1), 6)
	authorPub1 := append(make([]byte, libriapi.ECPubKeyLength-1), 7)
	authorPub2 := append(make([]byte, libriapi.ECPubKeyLength-1), 8)
	authorEntityID1 := "author entity ID 1"
	authorEntityID2 := "author entity ID 2"
	readerPub1 := append(make([]byte, libriapi.ECPubKeyLength-1), 9)
	readerPub2 := append(make([]byte, libriapi.ECPubKeyLength-1), 10)
	readerPub3 := append(make([]byte, libriapi.ECPubKeyLength-1), 11)
	readerEntityID1 := "reader entity ID 1"
	readerEntityID2 := "reader entity ID 2"

	prs := []*api.PublicationReceipt{
		{
			EnvelopeKey:     envKey1,
			EntryKey:        entryKey1,
			AuthorPublicKey: authorPub1,
			AuthorEntityId:  authorEntityID1,
			ReaderPublicKey: readerPub1,
			ReaderEntityId:  readerEntityID1,
			ReceivedTime:    now - 5,
		},
		{
			EnvelopeKey:     envKey2,
			EntryKey:        entryKey1,
			AuthorPublicKey: authorPub1,
			AuthorEntityId:  authorEntityID1,
			ReaderPublicKey: readerPub2,
			ReaderEntityId:  readerEntityID2,
			ReceivedTime:    now - 4,
		},
		{
			EnvelopeKey:     envKey3,
			EntryKey:        entryKey1,
			AuthorPublicKey: authorPub1,
			AuthorEntityId:  authorEntityID1,
			ReaderPublicKey: readerPub3,
			ReaderEntityId:  readerEntityID2,
			ReceivedTime:    now - 3,
		},
		{
			EnvelopeKey:     envKey4,
			EntryKey:        entryKey2,
			AuthorPublicKey: authorPub2,
			AuthorEntityId:  authorEntityID2,
			ReaderPublicKey: readerPub1,
			ReaderEntityId:  readerEntityID1,
			ReceivedTime:    now - 2,
		},
	}
	for _, pr := range prs {
		err := s.Put(pr)
		assert.Nil(t, err)
	}

	// check entry key filter
	limit := MaxSearchLimit
	f := &SearchFilters{
		EntryKey: entryKey1,
	}
	results, err := s.Search(f, limit)
	assert.Nil(t, err)
	assert.Len(t, results, 3)
	assert.Equal(t, envKey3, results[0].EnvelopeKey)
	assert.Equal(t, envKey2, results[1].EnvelopeKey)
	assert.Equal(t, envKey1, results[2].EnvelopeKey)

	// check author pub key filter
	f = &SearchFilters{
		AuthorPublicKey: authorPub1,
	}
	results, err = s.Search(f, limit)
	assert.Nil(t, err)
	assert.Len(t, results, 3)
	assert.Equal(t, envKey3, results[0].EnvelopeKey)
	assert.Equal(t, envKey2, results[1].EnvelopeKey)
	assert.Equal(t, envKey1, results[2].EnvelopeKey)

	// check author entity ID filter
	f = &SearchFilters{
		AuthorEntityID: authorEntityID1,
	}
	results, err = s.Search(f, limit)
	assert.Nil(t, err)
	assert.Len(t, results, 3)
	assert.Equal(t, envKey3, results[0].EnvelopeKey)
	assert.Equal(t, envKey2, results[1].EnvelopeKey)
	assert.Equal(t, envKey1, results[2].EnvelopeKey)

	// check reader pub key filter
	f = &SearchFilters{
		ReaderPublicKey: readerPub1,
	}
	results, err = s.Search(f, limit)
	assert.Nil(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, envKey4, results[0].EnvelopeKey)
	assert.Equal(t, envKey1, results[1].EnvelopeKey)

	// check reader entity ID filter
	f = &SearchFilters{
		ReaderEntityID: readerEntityID2,
	}
	results, err = s.Search(f, limit)
	assert.Nil(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, envKey3, results[0].EnvelopeKey)
	assert.Equal(t, envKey2, results[1].EnvelopeKey)

	// check entry + author filter
	f = &SearchFilters{
		EntryKey:        entryKey1,
		ReaderPublicKey: readerPub1,
	}
	results, err = s.Search(f, limit)
	assert.Nil(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, envKey1, results[0].EnvelopeKey)

	// check before filter
	f = &SearchFilters{
		EntryKey: entryKey1,
		Before:   now - 3,
	}
	results, err = s.Search(f, limit)
	assert.Nil(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, envKey2, results[0].EnvelopeKey)
	assert.Equal(t, envKey1, results[1].EnvelopeKey)

	// check before + after filter
	f = &SearchFilters{
		EntryKey: entryKey1,
		Before:   now - 3,
		After:    now - 4,
	}
	results, err = s.Search(f, limit)
	assert.Nil(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, envKey2, results[0].EnvelopeKey)

	// check limit
	f = &SearchFilters{
		EntryKey: entryKey1,
	}
	results, err = s.Search(f, 1)
	assert.Nil(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, envKey3, results[0].EnvelopeKey)
}
