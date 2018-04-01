package storage

import (
	"container/heap"
	"math/rand"
	"testing"
	"time"

	"github.com/drausin/libri/libri/common/id"
	libriapi "github.com/drausin/libri/libri/librarian/api"
	"github.com/elixirhealth/catalog/pkg/catalogapi"
	api "github.com/elixirhealth/catalog/pkg/catalogapi"
	"github.com/elixirhealth/service-base/pkg/util"
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
		"author entity ID only": {
			AuthorEntityID: "some entity ID",
		},
		"reader public key only": {
			ReaderPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
		},
		"reader entity ID only": {
			ReaderEntityID: "some entity ID",
		},
	}
	for info, f := range fs {
		err := ValidateSearchFilters(f)
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
		"after filter in epoch secs rather than micro-secs": {
			After: time.Now().Unix(),
		},
		"after filter in epoch milli-secs rather than micro-secs": {
			After: time.Now().Unix() * 1E3,
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
		err := ValidateSearchFilters(f)
		assert.NotNil(t, err, info)
	}
}

func TestPublicationReceipts(t *testing.T) {
	now := time.Now()
	prs := &PublicationReceipts{
		{
			EnvelopeKey:  []byte{1},
			ReceivedTime: now.Add(-1*time.Second).UnixNano() / 1E3,
		},
		{
			EnvelopeKey:  []byte{2},
			ReceivedTime: now.Add(-3*time.Second).UnixNano() / 1E3,
		},
		{
			EnvelopeKey:  []byte{3},
			ReceivedTime: now.Add(-2*time.Second).UnixNano() / 1E3,
		},
		{
			EnvelopeKey:  []byte{4},
			ReceivedTime: now.Add(-4*time.Second).UnixNano() / 1E3,
		},
	}
	heap.Init(prs)
	heap.Push(prs, &api.PublicationReceipt{
		EnvelopeKey:  []byte{5},
		ReceivedTime: now.Add(-500*time.Millisecond).UnixNano() / 1E3,
	})
	assert.Equal(t, 5, prs.Len())

	pr0 := prs.Peak()
	pr1 := heap.Pop(prs).(*api.PublicationReceipt)
	pr2 := heap.Pop(prs).(*api.PublicationReceipt)
	pr3 := heap.Pop(prs).(*api.PublicationReceipt)
	pr4 := heap.Pop(prs).(*api.PublicationReceipt)
	pr5 := heap.Pop(prs).(*api.PublicationReceipt)

	// should be ordered from earliest to latest
	assert.Equal(t, []byte{4}, pr0.EnvelopeKey)
	assert.Equal(t, []byte{4}, pr1.EnvelopeKey)
	assert.Equal(t, []byte{2}, pr2.EnvelopeKey)
	assert.Equal(t, []byte{3}, pr3.EnvelopeKey)
	assert.Equal(t, []byte{1}, pr4.EnvelopeKey)
	assert.Equal(t, []byte{5}, pr5.EnvelopeKey)
}
