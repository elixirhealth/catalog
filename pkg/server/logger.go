package server

import (
	"encoding/hex"

	"github.com/drausin/libri/libri/common/id"
	"github.com/elixirhealth/catalog/pkg/catalogapi"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	logEnvelopeKey         = "envelope_key"
	logEntryKeyFilterShort = "entry_key_filter_short"
	logAuthorPubKeyFilter  = "author_pub_key_filter"
	logReaderPubKeyFilter  = "reader_pub_key_filter"
	logAuthorEntityID      = "author_entity_id"
	logReaderEntityID      = "reader_entity_id"
	logBeforeTimeFilter    = "before_time_filter"
	logAfterTimeFilter     = "after_time_filter"
	logNResults            = "n_results"
	logStorage             = "storage"
	logGCPProjectID        = "gcp_project_id"
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
	fs := []zapcore.Field{
		zap.String(logEntryKeyFilterShort, hex.EncodeToString(rq.EntryKey)),
		zap.String(logAuthorPubKeyFilter, hex.EncodeToString(rq.AuthorPublicKey)),
		zap.String(logAuthorEntityID, rq.AuthorEntityId),
		zap.String(logReaderPubKeyFilter, hex.EncodeToString(rq.ReaderPublicKey)),
		zap.String(logReaderEntityID, rq.ReaderEntityId),
		zap.Time(logBeforeTimeFilter, catalogapi.FromEpochMicros(rq.Before)),
		zap.Time(logAfterTimeFilter, catalogapi.FromEpochMicros(rq.After)),
	}
	return fs
}
