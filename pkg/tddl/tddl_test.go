package tddl

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/beiai0xff/turl/pkg/db"
	"github.com/beiai0xff/turl/test"
)

const (
	testSeqName = "test-tddl"
)

func TestMain(m *testing.M) {
	db, err := db.NewDB(test.DSN)
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&Sequence{})
	if err != nil {
		panic(err)
	}

	exitCode := m.Run()
	db.Exec("DROP TABLE sequences")
	os.Exit(exitCode)
}

func newMockDB(t *testing.T) *gorm.DB {
	db, err := db.NewDB(test.DSN)
	require.NoError(t, err)

	return db
}

func TestNewSequence_Interface(t *testing.T) {
	gormDB := newMockDB(t)
	t.Cleanup(func() {
		gormDB.Exec("DELETE FROM sequences")
	})

	_, err := New(gormDB, &Config{
		Step:     10,
		SeqName:  testSeqName,
		StartNum: 10000,
	})
	require.NoError(t, err)
}

func TestNewSequence(t *testing.T) {
	gormDB := newMockDB(t)
	gormDB.Exec("DROP TABLE sequences")
	gormDB.AutoMigrate(&Sequence{})

	t.Cleanup(func() {
		gormDB.Exec("DELETE FROM sequences")
	})

	s, err := newSequence(gormDB, &Config{
		Step:     100,
		SeqName:  testSeqName,
		StartNum: 10000,
	})
	assert.NoError(t, err)
	t.Cleanup(s.Close)

	assert.Equal(t, uint64(100), s.step)
	assert.Equal(t, uint(1), s.rowID)
	assert.Equal(t, uint64(10100), s.max)
	assert.Equal(t, uint64(9999), s.curr.Load())
	assert.Equal(t, uint64(10000), s.Next(context.Background()))
}
func Test_tddlSequence_createRecord(t *testing.T) {
	gormDB := newMockDB(t)
	t.Cleanup(func() {
		gormDB.Exec("DELETE FROM sequences")
	})

	s := tddlSequence{conn: gormDB}
	_, err := s.createRecord(testSeqName, 10000)
	require.NoError(t, err)

	// create again should return error
	_, err = s.createRecord(testSeqName, 10000)
	require.Error(t, err)
}

func Test_tddlSequence_getRowID(t *testing.T) {
	gormDB := newMockDB(t)
	t.Cleanup(func() {
		gormDB.Exec("DELETE FROM sequences")
	})

	s := tddlSequence{conn: gormDB}
	require.NoError(t, s.getRowID(testSeqName, 10000))
	pre := s.rowID

	// get again should return nil, and the rowID should be the same
	require.NoError(t, s.getRowID(testSeqName, 10000))
	require.Equal(t, pre, s.rowID)
}

func Test_tddlSequence_Next(t *testing.T) {
	gormDB := newMockDB(t)
	t.Cleanup(func() {
		gormDB.Exec("DELETE FROM sequences")
	})

	s, err := newSequence(gormDB, &Config{
		Step:     100,
		SeqName:  testSeqName,
		StartNum: 10000,
	})
	assert.NoError(t, err)
	t.Cleanup(s.Close)

	wg, testDataLength := sync.WaitGroup{}, 10000
	ch := make(chan uint64, testDataLength)
	start := time.Now()
	for range testDataLength {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch <- s.Next(context.Background())
		}()
	}

	wg.Wait()
	fmt.Println("time:", time.Since(start))

	close(ch)
	arr := make([]uint64, 0, testDataLength)
	for v := range ch {
		arr = append(arr, v)
	}

	assert.Equal(t, testDataLength, len(arr))

	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})

	for i, x := range arr {
		assert.Equal(t, i+10000, int(x))
	}

	assert.Equal(t, testDataLength+10000, int(s.Next(context.Background())))
}

func Test_tddlSequence_multi_clients(t *testing.T) {
	gormDB := newMockDB(t)
	t.Cleanup(func() {
		gormDB.Exec("DELETE FROM sequences")
	})

	s1, err := newSequence(gormDB, &Config{
		Step:     100,
		SeqName:  testSeqName,
		StartNum: 10000,
	})
	assert.NoError(t, err)
	t.Cleanup(s1.Close)

	s2, err := newSequence(gormDB, &Config{
		Step:     100,
		SeqName:  testSeqName,
		StartNum: 10000,
	})
	assert.NoError(t, err)
	t.Cleanup(s2.Close)

	wg, testDataLength := sync.WaitGroup{}, 10000
	ch := make(chan uint64, testDataLength)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range testDataLength / 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ch <- s1.Next(context.Background())
			}()
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()
		for range testDataLength / 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ch <- s2.Next(context.Background())
			}()
		}
	}()

	wg.Wait()
	close(ch)

	// check sequence is valid
	arr := make([]uint64, 0, testDataLength)
	for v := range ch {
		arr = append(arr, v)
	}

	assert.Equal(t, testDataLength, len(arr))

	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})

	for i, x := range arr {
		assert.Equal(t, i+10000, int(x))
	}

	assert.Equal(t, testDataLength+10000, int(s1.Next(context.Background())))
}
