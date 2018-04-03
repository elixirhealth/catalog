package postgres

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/drausin/libri/libri/common/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	logEnvelopeKey = "envelope_key"
	nResults       = "n_results"
	logSQL         = "sql"
	logArgs        = "args"
)

func logPutInsert(q sq.InsertBuilder) []zapcore.Field {
	qSQL, args, err := q.ToSql()
	errors.MaybePanic(err)
	return []zapcore.Field{
		zap.String(logSQL, qSQL),
		zap.Array(logArgs, queryArgs(args)),
	}
}

func logSearchSelect(q sq.SelectBuilder) []zapcore.Field {
	qSQL, args, err := q.ToSql()
	errors.MaybePanic(err)
	return []zapcore.Field{
		zap.String(logSQL, qSQL),
		zap.Array(logArgs, queryArgs(args)),
	}
}

type queryArgs []interface{}

func (qas queryArgs) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, qa := range qas {
		switch val := qa.(type) {
		case string:
			enc.AppendString(val)
		default:
			if err := enc.AppendReflected(qa); err != nil {
				return err
			}
		}
	}
	return nil
}
