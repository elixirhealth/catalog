package catalogapi

import (
	"math/rand"
	"testing"
	"time"

	"github.com/drausin/libri/libri/common/id"
	"github.com/drausin/libri/libri/librarian/api"
	"github.com/elxirhealth/service-base/pkg/util"
	"github.com/magiconair/properties/assert"
)

func TestValidatePublicationReceipt(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	cases := map[string]struct {
		pr          *PublicationReceipt
		expectedErr error
	}{
		"ok": {
			pr: &PublicationReceipt{
				EnvelopeKey:     util.RandBytes(rng, id.Length),
				EntryKey:        util.RandBytes(rng, id.Length),
				AuthorPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReaderPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReceivedTime:    ToEpochMicros(time.Now()),
			},
		},
		"missing envelope key": {
			pr: &PublicationReceipt{
				EntryKey:        util.RandBytes(rng, id.Length),
				AuthorPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReaderPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReceivedTime:    ToEpochMicros(time.Now()),
			},
			expectedErr: ErrMissingEnvelopeKey,
		},
		"missing entry key": {
			pr: &PublicationReceipt{
				EnvelopeKey:     util.RandBytes(rng, id.Length),
				AuthorPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReaderPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReceivedTime:    ToEpochMicros(time.Now()),
			},
			expectedErr: ErrMissingEntryKey,
		},
		"missing author public key": {
			pr: &PublicationReceipt{
				EnvelopeKey:     util.RandBytes(rng, id.Length),
				EntryKey:        util.RandBytes(rng, id.Length),
				ReaderPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReceivedTime:    ToEpochMicros(time.Now()),
			},
			expectedErr: ErrMissingAuthorPublicKey,
		},
		"missing reader public key": {
			pr: &PublicationReceipt{
				EnvelopeKey:     util.RandBytes(rng, id.Length),
				EntryKey:        util.RandBytes(rng, id.Length),
				AuthorPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReceivedTime:    ToEpochMicros(time.Now()),
			},
			expectedErr: ErrMissingReaderPublicKey,
		},
		"missing received time": {
			pr: &PublicationReceipt{
				EnvelopeKey:     util.RandBytes(rng, id.Length),
				EntryKey:        util.RandBytes(rng, id.Length),
				AuthorPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReaderPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
			},
			expectedErr: ErrMissingReceivedTime,
		},
		"bad envelope key": {
			pr: &PublicationReceipt{
				EnvelopeKey:     util.RandBytes(rng, id.Length-1),
				EntryKey:        util.RandBytes(rng, id.Length),
				AuthorPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReaderPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReceivedTime:    ToEpochMicros(time.Now()),
			},
			expectedErr: ErrUnexpectedEnvelopeKeyLength,
		},
		"bad entry key": {
			pr: &PublicationReceipt{
				EnvelopeKey:     util.RandBytes(rng, id.Length),
				EntryKey:        util.RandBytes(rng, id.Length-1),
				AuthorPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReaderPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReceivedTime:    ToEpochMicros(time.Now()),
			},
			expectedErr: ErrUnexpectedEntryKeyLength,
		},
		"bad author public key": {
			pr: &PublicationReceipt{
				EnvelopeKey:     util.RandBytes(rng, id.Length),
				EntryKey:        util.RandBytes(rng, id.Length),
				AuthorPublicKey: util.RandBytes(rng, api.ECPubKeyLength-1),
				ReaderPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReceivedTime:    ToEpochMicros(time.Now()),
			},
			expectedErr: ErrUnexpectedAuthorPubKeyLength,
		},
		"bad reader public key": {
			pr: &PublicationReceipt{
				EnvelopeKey:     util.RandBytes(rng, id.Length),
				EntryKey:        util.RandBytes(rng, id.Length),
				AuthorPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReaderPublicKey: util.RandBytes(rng, api.ECPubKeyLength-1),
				ReceivedTime:    ToEpochMicros(time.Now()),
			},
			expectedErr: ErrUnexpectedReaderPubKeyLength,
		},
		"bad received time": {
			pr: &PublicationReceipt{
				EnvelopeKey:     util.RandBytes(rng, id.Length),
				EntryKey:        util.RandBytes(rng, id.Length),
				AuthorPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReaderPublicKey: util.RandBytes(rng, api.ECPubKeyLength),
				ReceivedTime:    time.Now().Unix(),
			},
			expectedErr: ErrEarlierReceivedTime,
		},
	}
	for info, c := range cases {
		err := ValidatePublicationReceipt(c.pr)
		assert.Equal(t, c.expectedErr, err, info)
	}
}
