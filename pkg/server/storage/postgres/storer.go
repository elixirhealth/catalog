package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	cerrors "github.com/drausin/libri/libri/common/errors"
	"github.com/drausin/libri/libri/common/id"
	api "github.com/elixirhealth/catalog/pkg/catalogapi"
	"github.com/elixirhealth/catalog/pkg/server/storage"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"go.uber.org/zap"
)

const (
	pubSchema    = "publication"
	receiptTable = "receipt"

	envKeyCol         = "envelope_key"
	entryKeyCol       = "entry_key"
	authorPubKeyCol   = "author_public_key"
	authorEntityIDCol = "author_entity_id"
	readerPubKeyCol   = "reader_public_key"
	readerEntityIDCol = "reader_entity_id"
	receivedTimeCol   = "received_time"
	receivedDateCol   = "received_date"

	isoDate             = "2006-01-02"
	onConflictDoNothing = "ON CONFLICT DO NOTHING"
)

var (
	psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	fqReceiptTable = pubSchema + "." + receiptTable

	errEmptyDBUrl            = errors.New("empty DB URL")
	errUnexpectedStorageType = errors.New("unexpected storage type")
)

type storer struct {
	params  *storage.Parameters
	db      *sql.DB
	dbCache sq.DBProxyContext
	qr      Querier
	logger  *zap.Logger
}

// New creates a new storage.Storer backed by a Postgres DB at the given dbURL.
func New(dbURL string, params *storage.Parameters, logger *zap.Logger) (storage.Storer, error) {
	if dbURL == "" {
		return nil, errEmptyDBUrl
	}
	if params.Type != bstorage.Postgres {
		return nil, errUnexpectedStorageType
	}
	db, err := sql.Open("postgres", dbURL)
	cerrors.MaybePanic(err)
	return &storer{
		params:  params,
		db:      db,
		dbCache: sq.NewStmtCacher(db),
		qr:      NewQuerier(),
		logger:  logger,
	}, nil
}

func (s *storer) Put(pr *api.PublicationReceipt) error {
	if err := api.ValidatePublicationReceipt(pr); err != nil {
		return err
	}
	q := psql.RunWith(s.dbCache).
		Insert(fqReceiptTable).
		SetMap(getPutPubReceiptStmtValues(pr)).
		Suffix(onConflictDoNothing)
	s.logger.Debug("inserting pub receipt", logPutInsert(q)...)
	ctx, cancel := context.WithTimeout(context.Background(), s.params.PutTimeout)
	defer cancel()
	r, err := s.qr.InsertExecContext(ctx, q)
	if err != nil {
		return err
	}
	nRows, err := r.RowsAffected()
	cerrors.MaybePanic(err) // should never happen w/ Postgres
	logEnvKey := zap.String(logEnvelopeKey, id.Hex(pr.EnvelopeKey))
	if nRows == 0 {
		s.logger.Debug("publication receipt already exists", logEnvKey)
		return nil
	}
	s.logger.Debug("stored new publication receipt", logEnvKey)
	return nil
}

func (s *storer) Search(
	fs *storage.SearchFilters, limit uint32,
) ([]*api.PublicationReceipt, error) {
	if err := storage.ValidateSearchFilters(fs); err != nil {
		return nil, err
	}
	if limit > storage.MaxSearchLimit {
		return nil, storage.ErrSearchLimitTooLarge
	}
	cols, _, _ := prepPubReceiptScan()
	q := psql.RunWith(s.dbCache).
		Select(cols...).
		From(fqReceiptTable).
		Where(getSearchEqPreds(fs)).
		Where(getSearchRangePreds(fs)).
		OrderBy(receivedTimeCol + " DESC").
		Limit(uint64(limit))
	s.logger.Debug("search for pub receipts", logSearchSelect(q)...)
	ctx, cancel := context.WithTimeout(context.Background(), s.params.SearchTimeout)
	defer cancel()
	rows, err := s.qr.SelectQueryContext(ctx, q)
	prs, i := make([]*api.PublicationReceipt, storage.MaxSearchLimit), 0
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		s.logger.Debug("found no search results")
		return prs[:i], nil
	}
	for rows.Next() {
		_, dest, create := prepPubReceiptScan()
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		prs[i] = create()
		i++
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	s.logger.Debug("found search results", zap.Int(nResults, i))
	return prs[:i], nil
}

func (s *storer) Close() error {
	return s.db.Close()
}

func getSearchEqPreds(f *storage.SearchFilters) map[string]interface{} {
	preds := make(map[string]interface{})
	if f.EntryKey != nil {
		preds[entryKeyCol] = f.EntryKey
	}
	if f.AuthorPublicKey != nil {
		preds[authorPubKeyCol] = f.AuthorPublicKey
	}
	if f.AuthorEntityID != "" {
		preds[authorEntityIDCol] = f.AuthorEntityID
	}
	if f.ReaderPublicKey != nil {
		preds[readerPubKeyCol] = f.ReaderPublicKey
	}
	if f.ReaderEntityID != "" {
		preds[readerEntityIDCol] = f.ReaderEntityID
	}
	return preds
}

func getSearchRangePreds(f *storage.SearchFilters) string {
	// date filters below are technically not required; but since date cardinality is much
	// smaller than time, we have an index on it, which should greatly speed up range query
	preds := make([]string, 0, 4)
	if f.Before != 0 {
		beforeTime := api.FromEpochMicros(f.Before)
		timePred := fmt.Sprintf("%s < %d", receivedTimeCol, f.Before)
		preds = append(preds, timePred)

		// +1 day for date bound to safely handle timezones
		beforeDate := beforeTime.Add(24 * time.Hour)
		datePred := fmt.Sprintf("%s <= '%s'", receivedDateCol, beforeDate.Format(isoDate))
		preds = append(preds, datePred)
	}
	if f.After != 0 {
		afterTime := api.FromEpochMicros(f.After)
		timePred := fmt.Sprintf("%s >= %d", receivedTimeCol, f.After)
		preds = append(preds, timePred)

		// -1 day for date bound to safely handle timezones
		afterDate := afterTime.Add(-24 * time.Hour)
		datePred := fmt.Sprintf("%s >= '%s'", receivedDateCol, afterDate.Format(isoDate))
		preds = append(preds, datePred)
	}
	return strings.Join(preds, " AND ")
}

func getPutPubReceiptStmtValues(pr *api.PublicationReceipt) map[string]interface{} {
	receivedTime := api.FromEpochMicros(pr.ReceivedTime)
	return map[string]interface{}{
		envKeyCol:         pr.EnvelopeKey,
		entryKeyCol:       pr.EntryKey,
		authorPubKeyCol:   pr.AuthorPublicKey,
		authorEntityIDCol: pr.AuthorEntityId,
		readerPubKeyCol:   pr.ReaderPublicKey,
		readerEntityIDCol: pr.ReaderEntityId,
		receivedTimeCol:   pr.ReceivedTime,
		receivedDateCol:   receivedTime.Format(isoDate),
	}
}

func prepPubReceiptScan() ([]string, []interface{}, func() *api.PublicationReceipt) {
	pr := &api.PublicationReceipt{}
	cols, dests := SplitColDests(0, []*ColDest{
		{envKeyCol, &pr.EnvelopeKey},
		{entryKeyCol, &pr.EntryKey},
		{authorPubKeyCol, &pr.AuthorPublicKey},
		{authorEntityIDCol, &pr.AuthorEntityId},
		{readerPubKeyCol, &pr.ReaderPublicKey},
		{readerEntityIDCol, &pr.ReaderEntityId},
		// no received date, which is just used to make searches fast
		{receivedTimeCol, &pr.ReceivedTime},
	})
	return cols, dests, func() *api.PublicationReceipt {
		pr.EnvelopeKey = *dests[0].(*[]byte)
		pr.EntryKey = *dests[1].(*[]byte)
		pr.AuthorPublicKey = *dests[2].(*[]byte)
		pr.AuthorEntityId = *dests[3].(*string)
		pr.ReaderPublicKey = *dests[4].(*[]byte)
		pr.ReaderEntityId = *dests[5].(*string)
		pr.ReceivedTime = *dests[6].(*int64)
		return pr
	}
}

// ColDest is a mapping from a column name to a sql.Scan destination type.
type ColDest struct {
	col  string
	dest interface{}
}

// SplitColDests returns a list of column names and their corresponding destination types plus a
// given number of extra destination capacity.
func SplitColDests(nExtraDest int, cds []*ColDest) ([]string, []interface{}) {
	dests := make([]interface{}, len(cds), len(cds)+nExtraDest)
	cols := make([]string, len(cds))
	for i, colDest := range cds {
		cols[i] = colDest.col
		dests[i] = colDest.dest
	}
	return cols, dests
}

// QueryRows is a container for the result of a Select query.
type QueryRows interface {
	Scan(dest ...interface{}) error
	Next() bool
	Close() error
	Err() error
}

// TODO (drausin) move below to service base

// Querier is an interface wrapper around Squirrel query builders and their results for improved
// mocking.
type Querier interface {
	SelectQueryContext(ctx context.Context, b sq.SelectBuilder) (QueryRows, error)
	SelectQueryRowContext(ctx context.Context, b sq.SelectBuilder) sq.RowScanner
	InsertExecContext(ctx context.Context, b sq.InsertBuilder) (sql.Result, error)
	UpdateExecContext(ctx context.Context, b sq.UpdateBuilder) (sql.Result, error)
}

type querierImpl struct{}

// NewQuerier returns a new Querier.
func NewQuerier() Querier {
	return &querierImpl{}
}

func (q *querierImpl) SelectQueryContext(
	ctx context.Context, b sq.SelectBuilder,
) (QueryRows, error) {
	return b.QueryContext(ctx)
}

func (q *querierImpl) SelectQueryRowContext(
	ctx context.Context, b sq.SelectBuilder,
) sq.RowScanner {
	return b.QueryRowContext(ctx)
}

func (q *querierImpl) InsertExecContext(
	ctx context.Context, b sq.InsertBuilder,
) (sql.Result, error) {
	return b.ExecContext(ctx)
}

func (q *querierImpl) UpdateExecContext(
	ctx context.Context, b sq.UpdateBuilder,
) (sql.Result, error) {
	return b.ExecContext(ctx)
}
