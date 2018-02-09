package server

import (
	"github.com/drausin/libri/libri/common/id"
	"github.com/elxirhealth/catalog/pkg/catalogapi"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	logEnvelopeKey             = "envelope_key"
	logEntryKeyFilterShort     = "entry_key_filter_short"
	logAuthorPubKeyFilterShort = "author_pub_key_filter_short"
	logReaderPubKeyFilterShort = "reader_pub_key_filter_short"
	logBeforeTimeFilter        = "before_time_filter"
	logNResults                = "n_results"
	logStorage                 = "storage"
)

func logPutRequestFields(rq *catalogapi.PutRequest) []zapcore.Field {
	if rq.Value == nil {
		return []zapcore.Field{}
	}
	return []zapcore.Field{
		zap.String(logEnvelopeKey, id.Hex(rq.Value.EnvelopeKey)),
	}
}

func logSearchRequestFields(rq *catalogapi.SearchRequest) []zapcore.Field {
	return []zapcore.Field{
		zap.String(logEntryKeyFilterShort, id.ShortHex(rq.EntryKey)),
		zap.String(logAuthorPubKeyFilterShort, id.ShortHex(rq.AuthorPublicKey)),
		zap.String(logReaderPubKeyFilterShort, id.ShortHex(rq.ReaderPublicKey)),
		zap.Time(logBeforeTimeFilter, catalogapi.FromEpochMicros(rq.Before)),
	}
}
