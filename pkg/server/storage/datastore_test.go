package storage

import (
	"context"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"cloud.google.com/go/datastore"
	cerrors "github.com/drausin/libri/libri/common/errors"
	"github.com/drausin/libri/libri/common/id"
	"github.com/drausin/libri/libri/common/logging"
	libriapi "github.com/drausin/libri/libri/librarian/api"
	api "github.com/elxirhealth/catalog/pkg/catalogapi"
	"github.com/elxirhealth/service-base/pkg/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

const (
	datastoreEmulatorHostEnv = "DATASTORE_EMULATOR_HOST"
	datastoreEmulatorAddr    = "localhost:2002"
	dummyDatastoreProject    = "dummy-datastore-test"
)

// TestDatastoreStorer_PutSearch is an integration test that spins up a DataStore emulator to
// ensure that the DataStore Search queries behave as expected.
func TestDatastoreStorer_PutSearch(t *testing.T) {
	dataDir, err := ioutil.TempDir("", "catalog-datastore-test")
	assert.Nil(t, err)
	defer cerrors.MaybePanic(os.RemoveAll(dataDir))
	datastoreProc := startDatastoreEmulator(dataDir, datastoreEmulatorAddr)
	defer stopDatastoreEmulator(datastoreProc)

	lg := logging.NewDevInfoLogger()
	d, err := NewDatastore(dummyDatastoreProject, NewDefaultParameters(), lg)
	assert.Nil(t, err)

	now := time.Now().Unix() * 1E6
	envKey1 := append(make([]byte, id.Length-1), 1)
	envKey2 := append(make([]byte, id.Length-1), 2)
	envKey3 := append(make([]byte, id.Length-1), 3)
	envKey4 := append(make([]byte, id.Length-1), 4)
	entryKey1 := append(make([]byte, id.Length-1), 5)
	entryKey2 := append(make([]byte, id.Length-1), 6)
	authorPub1 := append(make([]byte, libriapi.ECPubKeyLength-1), 7)
	authorPub2 := append(make([]byte, libriapi.ECPubKeyLength-1), 8)
	readerPub1 := append(make([]byte, libriapi.ECPubKeyLength-1), 9)
	readerPub2 := append(make([]byte, libriapi.ECPubKeyLength-1), 10)
	readerPub3 := append(make([]byte, libriapi.ECPubKeyLength-1), 11)

	prs := []*api.PublicationReceipt{
		{
			EnvelopeKey:     envKey1,
			EntryKey:        entryKey1,
			AuthorPublicKey: authorPub1,
			ReaderPublicKey: readerPub1,
			ReceivedTime:    now - 5,
		},
		{
			EnvelopeKey:     envKey2,
			EntryKey:        entryKey1,
			AuthorPublicKey: authorPub1,
			ReaderPublicKey: readerPub2,
			ReceivedTime:    now - 4,
		},
		{
			EnvelopeKey:     envKey3,
			EntryKey:        entryKey1,
			AuthorPublicKey: authorPub1,
			ReaderPublicKey: readerPub3,
			ReceivedTime:    now - 3,
		},
		{
			EnvelopeKey:     envKey4,
			EntryKey:        entryKey2,
			AuthorPublicKey: authorPub2,
			ReaderPublicKey: readerPub1,
			ReceivedTime:    now - 2,
		},
	}
	for _, pr := range prs {
		err = d.Put(pr)
		assert.Nil(t, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	count, err := d.(*datastoreStorer).client.Count(ctx, datastore.NewQuery(publicationReceiptKind))
	cancel()
	assert.Nil(t, err)
	assert.Equal(t, len(prs), count)

	// check entry key filter
	limit := MaxSearchLimit
	f := &SearchFilters{
		EntryKey: entryKey1,
	}
	results, err := d.Search(f, limit)
	assert.Nil(t, err)
	assert.Len(t, results, 3)
	assert.Equal(t, envKey3, results[0].EnvelopeKey)
	assert.Equal(t, envKey2, results[1].EnvelopeKey)
	assert.Equal(t, envKey1, results[2].EnvelopeKey)

	// check author filter
	f = &SearchFilters{
		AuthorPublicKey: authorPub1,
	}
	results, err = d.Search(f, limit)
	assert.Nil(t, err)
	assert.Len(t, results, 3)
	assert.Equal(t, envKey3, results[0].EnvelopeKey)
	assert.Equal(t, envKey2, results[1].EnvelopeKey)
	assert.Equal(t, envKey1, results[2].EnvelopeKey)

	// check reader filter
	f = &SearchFilters{
		ReaderPublicKey: readerPub1,
	}
	results, err = d.Search(f, limit)
	assert.Nil(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, envKey4, results[0].EnvelopeKey)
	assert.Equal(t, envKey1, results[1].EnvelopeKey)

	// check entry + author filter
	f = &SearchFilters{
		EntryKey:        entryKey1,
		ReaderPublicKey: readerPub1,
	}
	results, err = d.Search(f, limit)
	assert.Nil(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, envKey1, results[0].EnvelopeKey)

	// check before filter
	f = &SearchFilters{
		EntryKey: entryKey1,
		Before:   now - 3,
	}
	results, err = d.Search(f, limit)
	assert.Nil(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, envKey2, results[0].EnvelopeKey)
	assert.Equal(t, envKey1, results[1].EnvelopeKey)

	// check limit
	f = &SearchFilters{
		EntryKey: entryKey1,
	}
	results, err = d.Search(f, 1)
	assert.Nil(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, envKey3, results[0].EnvelopeKey)
}

func TestDatastoreStorer_Put_ok(t *testing.T) {
	lg, params := zap.NewNop(), NewDefaultParameters()
	rng := rand.New(rand.NewSource(0))
	pr := &api.PublicationReceipt{
		EnvelopeKey:     util.RandBytes(rng, id.Length),
		EntryKey:        util.RandBytes(rng, id.Length),
		AuthorPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
		ReaderPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
		ReceivedTime:    api.ToEpochMicros(time.Now()),
	}

	// mock PR already existing
	d := &datastoreStorer{
		client: &fixedDatastoreClient{
			value: pr,
		},
		logger: lg,
		params: params,
	}
	err := d.Put(pr)
	assert.Nil(t, err)

	// mock PR not already existing
	d = &datastoreStorer{
		client: &fixedDatastoreClient{},
		logger: lg,
		params: params,
	}
	err = d.Put(pr)
	assert.Nil(t, err)
}

func TestDatastoreStorer_Put_err(t *testing.T) {
	lg, params := zap.NewNop(), NewDefaultParameters()
	rng := rand.New(rand.NewSource(0))

	d := &datastoreStorer{
		client: &fixedDatastoreClient{},
		logger: lg,
		params: params,
	}
	pr := &api.PublicationReceipt{}
	err := d.Put(pr)
	assert.NotNil(t, err)

	d = &datastoreStorer{
		client: &fixedDatastoreClient{
			getErr: errors.New("some Get error"),
		},
		logger: lg,
		params: params,
	}
	pr = &api.PublicationReceipt{
		EnvelopeKey:     util.RandBytes(rng, id.Length),
		EntryKey:        util.RandBytes(rng, id.Length),
		AuthorPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
		ReaderPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
		ReceivedTime:    api.ToEpochMicros(time.Now()),
	}
	err = d.Put(pr)
	assert.NotNil(t, err)

	d = &datastoreStorer{
		client: &fixedDatastoreClient{
			getErr: datastore.ErrNoSuchEntity,
			putErr: errors.New("some Put error"),
		},
	}
	err = d.Put(pr)
	assert.NotNil(t, err)
}

func TestDatastoreStorer_Search_err(t *testing.T) {
	lg, params := zap.NewNop(), NewDefaultParameters()
	rng := rand.New(rand.NewSource(0))

	// invalid search filter
	d := &datastoreStorer{
		client: &fixedDatastoreClient{},
		logger: lg,
		params: params,
	}
	fs := &SearchFilters{}
	result, err := d.Search(fs, MaxSearchLimit)
	assert.Nil(t, result)
	assert.NotNil(t, err)

	// too large search limit
	fs = &SearchFilters{
		EntryKey: util.RandBytes(rng, id.Length),
	}
	result, err = d.Search(fs, MaxSearchLimit*2)
	assert.Nil(t, result)
	assert.Equal(t, ErrSearchLimitTooLarge, err)

	// Next error
	d = &datastoreStorer{
		client: &fixedDatastoreClient{},
		iter: &fixedDatastoreIterator{
			err: errors.New("some Next error"),
		},
		logger: lg,
		params: params,
	}
	result, err = d.Search(fs, MaxSearchLimit)
	assert.Nil(t, result)
	assert.NotNil(t, err)

	// decodeStoredPubReceipt error
	d = &datastoreStorer{
		client: &fixedDatastoreClient{},
		iter: &fixedDatastoreIterator{
			keys: []*datastore.Key{
				datastore.NameKey(publicationReceiptKind, "some env key", nil),
			},
			values: []*PublicationReceipt{
				{}, // invalid
			},
		},
		logger: lg,
		params: params,
	}
	result, err = d.Search(fs, MaxSearchLimit)
	assert.Nil(t, result)
	assert.NotNil(t, err)
}

func TestDecodeStoredPubReceipt_err(t *testing.T) {
	prs := map[string]*PublicationReceipt{
		"bad env key": {
			EnvelopeKey: datastore.NameKey(publicationReceiptKind, "$$$$", nil),
		},
		"bad entry key": {
			EnvelopeKey: datastore.NameKey(publicationReceiptKind, "abcd", nil),
			EntryKey:    "$$$$",
		},
		"bad author pub key": {
			EnvelopeKey:     datastore.NameKey(publicationReceiptKind, "abcd", nil),
			EntryKey:        "abcd",
			AuthorPublicKey: "$$$$",
		},
		"bad reader pub key": {
			EnvelopeKey:     datastore.NameKey(publicationReceiptKind, "abcd", nil),
			EntryKey:        "abcd",
			AuthorPublicKey: "abcd",
			ReaderPublicKey: "$$$$",
		},
	}
	for info, pr := range prs {
		decoded, err := decodeStoredPubReceipt(pr)
		assert.Nil(t, decoded, info)
		assert.NotNil(t, err, info)
	}
}

func startDatastoreEmulator(dataDir string, addr string) *os.Process {
	cmd := exec.Command("gcloud", "beta", "emulators", "datastore", "start",
		"--no-store-on-disk",
		"--host-port", addr,
		"--project", dummyDatastoreProject,
		"--data-dir", dataDir,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	cerrors.MaybePanic(err)
	os.Setenv(datastoreEmulatorHostEnv, addr)
	return cmd.Process
}

func stopDatastoreEmulator(process *os.Process) {
	pgid, err := syscall.Getpgid(process.Pid)
	cerrors.MaybePanic(err)
	err = syscall.Kill(-pgid, syscall.SIGKILL)
	cerrors.MaybePanic(err)
	log.Printf("stopped child processes under pid %d\n", pgid)
}

type fixedDatastoreClient struct {
	value      interface{}
	getErr     error
	putErr     error
	deleteErr  error
	countValue int
	countErr   error
	runResult  *datastore.Iterator
}

func (f *fixedDatastoreClient) Count(ctx context.Context, q *datastore.Query) (int, error) {
	return f.countValue, f.countErr
}

func (f *fixedDatastoreClient) Put(key *datastore.Key, value interface{}) (*datastore.Key, error) {
	if f.putErr != nil {
		return nil, f.putErr
	}
	f.value = value
	return key, nil
}

func (f *fixedDatastoreClient) Get(key *datastore.Key, dest interface{}) error {
	if f.getErr != nil {
		return f.getErr
	}
	if f.value == nil {
		return datastore.ErrNoSuchEntity
	} else if key.Kind == publicationReceiptKind {
		dest.(*PublicationReceipt).EnvelopeKey = key
		dest.(*PublicationReceipt).EntryKey =
			hex.EncodeToString(f.value.(*api.PublicationReceipt).EntryKey)
		dest.(*PublicationReceipt).AuthorPublicKey =
			hex.EncodeToString(f.value.(*api.PublicationReceipt).AuthorPublicKey)
		dest.(*PublicationReceipt).ReaderPublicKey =
			hex.EncodeToString(f.value.(*api.PublicationReceipt).ReaderPublicKey)
	}
	return nil
}

func (f *fixedDatastoreClient) Delete(keys []*datastore.Key) error {
	f.value = nil
	return f.deleteErr
}

func (f *fixedDatastoreClient) Run(ctx context.Context, q *datastore.Query) *datastore.Iterator {
	return f.runResult
}

type fixedDatastoreIterator struct {
	err    error
	keys   []*datastore.Key
	values []*PublicationReceipt
	offset int
}

func (f *fixedDatastoreIterator) Init(iter *datastore.Iterator) {}

func (f *fixedDatastoreIterator) Next(dest interface{}) (*datastore.Key, error) {
	if f.err != nil {
		return nil, f.err
	}
	defer func() { f.offset++ }()
	if f.offset == len(f.values) {
		return nil, iterator.Done
	}
	v := f.values[f.offset]
	dest.(*PublicationReceipt).EnvelopeKey = f.keys[f.offset]
	dest.(*PublicationReceipt).EntryKey = v.EntryKey
	dest.(*PublicationReceipt).AuthorPublicKey = v.AuthorPublicKey
	dest.(*PublicationReceipt).ReaderPublicKey = v.ReaderPublicKey
	return f.keys[f.offset], nil
}