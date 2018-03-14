package catalogapi

import (
	"errors"
	"fmt"
	"time"

	"github.com/drausin/libri/libri/common/id"
	"github.com/drausin/libri/libri/librarian/api"
)

var (
	// ErrMissingValue indicates when a PutRequest is missing the value field.
	ErrMissingValue = errors.New("missing Put request value")

	// ErrMissingEnvelopeKey denotes when the envelope key is missing.
	ErrMissingEnvelopeKey = errors.New("missing envelope key")

	// ErrMissingEntryKey denotes when the entry key is missing.
	ErrMissingEntryKey = errors.New("missing entry key")

	// ErrMissingAuthorPublicKey denotes when the author public key is missing.
	ErrMissingAuthorPublicKey = errors.New("missing author public key")

	// ErrMissingReaderPublicKey denotes when the reader public key is missing.
	ErrMissingReaderPublicKey = errors.New("missing reader public key")

	// ErrMissingReceivedTime denotes when the received time is missing.
	ErrMissingReceivedTime = errors.New("missing received time")

	// ErrUnexpectedEnvelopeKeyLength denotes when the envelope key length is not the expected
	// length.
	ErrUnexpectedEnvelopeKeyLength = errors.New("unexpected envelope key length")

	// ErrUnexpectedEntryKeyLength denotes when the entry key length is not the expected
	// length.
	ErrUnexpectedEntryKeyLength = errors.New("unexpected entry key length")

	// ErrUnexpectedAuthorPubKeyLength denotes when the author public key length is not the
	// expected length.
	ErrUnexpectedAuthorPubKeyLength = errors.New("unexpected author public key length")

	// ErrUnexpectedReaderPubKeyLength denotes when the reader public key length is not the
	// expected length.
	ErrUnexpectedReaderPubKeyLength = errors.New("unexpected reader public key length")

	// ErrEarlierReceivedTime denotes when the received time is earlier than the
	// minimum value (often because it is represening seconds or milliseconds instead of
	// microseconds from the epoch).
	ErrEarlierReceivedTime = fmt.Errorf("received time earlier than min %d",
		minReceivedTime)

	// ErrSearchZeroLimit denotes when a search request has a zero limit.
	ErrSearchZeroLimit = errors.New("search zero limit not allowed")

	// ErrAfterRequiresBefore denotes an a search request defines an after filter but no before
	// filter.
	ErrAfterRequiresBefore = errors.New("after time filter requires before filter to be set")

	minReceivedTime = ToEpochMicros(time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC))
)

// ValidatePutRequest checks that a PutRequest has the required value.
func ValidatePutRequest(rq *PutRequest) error {
	if rq.Value == nil {
		return ErrMissingValue
	}
	return ValidatePublicationReceipt(rq.Value)
}

// ValidateSearchRequest checks that a SearchRequest has the required values.
func ValidateSearchRequest(rq *SearchRequest) error {
	if rq.Limit == 0 {
		return ErrSearchZeroLimit
	}
	if rq.After != 0 && rq.Before == 0 { // if after is defined, before must be too
		return ErrAfterRequiresBefore
	}
	return nil
}

// ValidatePublicationReceipt validates that a *PublicationReceipt has the required fields
// populated with valid values.
func ValidatePublicationReceipt(pr *PublicationReceipt) error {
	if err := validatePresent(pr); err != nil {
		return err
	}
	return validateCorrect(pr)
}

func validatePresent(pr *PublicationReceipt) error {
	if pr.EnvelopeKey == nil {
		return ErrMissingEnvelopeKey
	}
	if pr.EntryKey == nil {
		return ErrMissingEntryKey
	}
	if pr.AuthorPublicKey == nil {
		return ErrMissingAuthorPublicKey
	}
	// ok for AuthorEntityID to be empty
	if pr.ReaderPublicKey == nil {
		return ErrMissingReaderPublicKey
	}
	// ok for ReaderEntityID to be empty
	if pr.ReceivedTime == 0 {
		return ErrMissingReceivedTime
	}
	return nil
}

func validateCorrect(pr *PublicationReceipt) error {
	if len(pr.EnvelopeKey) != id.Length {
		return ErrUnexpectedEnvelopeKeyLength
	}
	if len(pr.EntryKey) != id.Length {
		return ErrUnexpectedEntryKeyLength
	}
	if len(pr.AuthorPublicKey) != api.ECPubKeyLength {
		return ErrUnexpectedAuthorPubKeyLength
	}
	if len(pr.ReaderPublicKey) != api.ECPubKeyLength {
		return ErrUnexpectedReaderPubKeyLength
	}
	if pr.ReceivedTime < minReceivedTime {
		return ErrEarlierReceivedTime
	}
	return nil
}

// FromEpochMicros converts a epoch-time in microseconds to a time.Time.
func FromEpochMicros(epochMicros int64) time.Time {
	secs := epochMicros / 1E6
	nanos := (epochMicros % 1E6) * 1000
	return time.Unix(secs, nanos)
}

// ToEpochMicros converts a time.Tiem to epoch time microseconds.
func ToEpochMicros(t time.Time) int64 {
	return t.UnixNano() / 1000
}
