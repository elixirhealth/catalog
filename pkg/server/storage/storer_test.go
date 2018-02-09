package storage

import (
	"container/heap"
	"math/rand"
	"testing"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/drausin/libri/libri/common/id"
	libriapi "github.com/drausin/libri/libri/librarian/api"
	"github.com/elxirhealth/catalog/pkg/catalogapi"
	"github.com/elxirhealth/service-base/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultParameters(t *testing.T) {
	p := NewDefaultParameters()
	assert.NotEmpty(t, p.SearchQueryTimeout)
}

func TestValidateSearchFilters_ok(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	fs := map[string]*SearchFilters{
		"entry key only": {
			EntryKey: util.RandBytes(rng, id.Length),
		},
		"author public key only": {
			AuthorPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
		},
		"reader public key only": {
			ReaderPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
		},
	}
	for info, f := range fs {
		err := validateSearchFilters(f)
		assert.Nil(t, err, info)
	}
}

func TestValidateSearchFilters_err(t *testing.T) {
	fs := map[string]*SearchFilters{
		"before filter in epoch secs rather than micro-secs": {
			Before: time.Now().Unix(),
		},
		"before filter in epoch milli-secs rather than micro-secs": {
			Before: time.Now().Unix() * 1E3,
		},
		"bad entry key": {
			EntryKey: []byte{1, 2, 3},
		},
		"bad author public key": {
			AuthorPublicKey: []byte{1, 2, 3},
		},
		"bad reader public key": {
			ReaderPublicKey: []byte{1, 2, 3},
		},
		"no equality filters": {
			Before: catalogapi.ToEpochMicros(time.Now()),
		},
	}
	for info, f := range fs {
		err := validateSearchFilters(f)
		assert.NotNil(t, err, info)
	}
}

func TestPublicationReceipts(t *testing.T) {
	now := time.Now()
	prs := &publicationReceipts{
		{
			EnvelopeKey:  datastore.NameKey(publicationReceiptKind, "key 1", nil),
			ReceivedTime: now.Add(-1 * time.Second),
		},
		{
			EnvelopeKey:  datastore.NameKey(publicationReceiptKind, "key 2", nil),
			ReceivedTime: now.Add(-3 * time.Second),
		},
		{
			EnvelopeKey:  datastore.NameKey(publicationReceiptKind, "key 3", nil),
			ReceivedTime: now.Add(-2 * time.Second),
		},
		{
			EnvelopeKey:  datastore.NameKey(publicationReceiptKind, "key 4", nil),
			ReceivedTime: now.Add(-4 * time.Second),
		},
	}
	heap.Init(prs)
	heap.Push(prs, &PublicationReceipt{
		EnvelopeKey:  datastore.NameKey(publicationReceiptKind, "key 5", nil),
		ReceivedTime: now.Add(-500 * time.Millisecond),
	})
	assert.Equal(t, 5, prs.Len())

	pr0 := prs.Peak()
	pr1 := heap.Pop(prs).(*PublicationReceipt)
	pr2 := heap.Pop(prs).(*PublicationReceipt)
	pr3 := heap.Pop(prs).(*PublicationReceipt)
	pr4 := heap.Pop(prs).(*PublicationReceipt)
	pr5 := heap.Pop(prs).(*PublicationReceipt)

	// should be ordered from earliest to latest
	assert.Equal(t, "key 4", pr0.EnvelopeKey.Name)
	assert.Equal(t, "key 4", pr1.EnvelopeKey.Name)
	assert.Equal(t, "key 2", pr2.EnvelopeKey.Name)
	assert.Equal(t, "key 3", pr3.EnvelopeKey.Name)
	assert.Equal(t, "key 1", pr4.EnvelopeKey.Name)
	assert.Equal(t, "key 5", pr5.EnvelopeKey.Name)
}
