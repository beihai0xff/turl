package tddl

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/beihai0xff/turl/configs"
	"github.com/beihai0xff/turl/internal/tests"
	"github.com/beihai0xff/turl/pkg/db/mysql"
)

const (
	testSeqName = "test-tddl"
)

func TestMain(m *testing.M) {
	_ = tests.CreateTable(&Sequence{})

	code := m.Run()
	_ = tests.DropTable(&Sequence{})

	os.Exit(code)
}

func newMockDB(t *testing.T) *gorm.DB {
	db, err := mysql.New(tests.GlobalConfig.MySQL)
	require.NoError(t, err)

	return db
}

func TestNewSequence_Interface(t *testing.T) {
	gormDB := newMockDB(t)
	_, err := New(gormDB, &configs.TDDLConfig{
		Step:     10,
		SeqName:  testSeqName,
		StartNum: 10000,
	})
	require.NoError(t, err)
}

func TestNewSequence(t *testing.T) {
	gormDB := newMockDB(t)
	require.NoError(t, gormDB.Exec("DELETE FROM sequences").Error)

	s, err := newSequence(gormDB, &configs.TDDLConfig{
		Step:     100,
		SeqName:  testSeqName,
		StartNum: 10000,
	})
	require.NoError(t, err)
	t.Cleanup(s.Close)

	require.Equal(t, 100, int(s.step))
	require.Greater(t, int(s.rowID), 1)
	require.Equal(t, uint64(10100), s.max)
	require.Equal(t, uint64(10000), s.curr.Load())
	next, err := s.Next(context.Background())
	require.NoError(t, err)
	require.Equal(t, uint64(10000), next)
}
func Test_tddlSequence_createRecord(t *testing.T) {
	gormDB := newMockDB(t)
	require.NoError(t, gormDB.Exec("DELETE FROM sequences").Error)

	s := tddlSequence{conn: gormDB}
	_, err := s.createRecord(testSeqName, 10000)
	require.NoError(t, err)

	// create again should return error
	_, err = s.createRecord(testSeqName, 10000)
	require.Error(t, err)
}

func Test_tddlSequence_getRowID(t *testing.T) {
	gormDB := newMockDB(t)
	require.NoError(t, gormDB.Exec("DELETE FROM sequences").Error)

	s := tddlSequence{conn: gormDB}
	require.NoError(t, s.getRowID(testSeqName, 10000))
	pre := s.rowID

	// get again should return nil, and the rowID should be the same
	require.NoError(t, s.getRowID(testSeqName, 10000))
	require.Equal(t, pre, s.rowID)
}

func Test_tddlSequence_Next(t *testing.T) {
	gormDB := newMockDB(t)
	require.NoError(t, gormDB.Exec("DELETE FROM sequences").Error)

	s, err := newSequence(gormDB, &configs.TDDLConfig{
		Step:     100,
		SeqName:  testSeqName,
		StartNum: 10000,
	})
	require.NoError(t, err)
	t.Cleanup(s.Close)

	wg, testDataLength := sync.WaitGroup{}, 1000
	ch := make(chan uint64, testDataLength)
	start := time.Now()
	for range testDataLength {
		wg.Add(1)
		go func() {
			defer wg.Done()
			next, _ := s.Next(context.Background())
			ch <- next
		}()
	}

	wg.Wait()
	fmt.Println("time:", time.Since(start))

	close(ch)
	arr := make([]uint64, 0, testDataLength)
	for v := range ch {
		arr = append(arr, v)
	}

	require.Equal(t, testDataLength, len(arr))

	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})

	for i, x := range arr {
		require.Equal(t, i+10000, int(x))
	}

	next, err := s.Next(context.Background())
	require.NoError(t, err)
	require.Equal(t, testDataLength+10000, int(next))
}

func Test_tddlSequence_multi_clients(t *testing.T) {
	gormDB := newMockDB(t)
	require.NoError(t, gormDB.Exec("DELETE FROM sequences").Error)

	s1, err := newSequence(gormDB, &configs.TDDLConfig{
		Step:     100,
		SeqName:  testSeqName,
		StartNum: 10000,
	})
	require.NoError(t, err)
	t.Cleanup(s1.Close)

	s2, err := newSequence(gormDB, &configs.TDDLConfig{
		Step:     100,
		SeqName:  testSeqName,
		StartNum: 10000,
	})
	require.NoError(t, err)
	t.Cleanup(s2.Close)

	wg, testDataLength := sync.WaitGroup{}, 10000
	ch := make(chan uint64, testDataLength)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range testDataLength / 2 {
			next, _ := s1.Next(context.Background())
			ch <- next
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for range testDataLength / 2 {
			next, _ := s2.Next(context.Background())
			ch <- next
		}
	}()

	wg.Wait()
	close(ch)

	// check sequence is valid
	arr := make([]uint64, 0, testDataLength)
	for v := range ch {
		arr = append(arr, v)
	}

	require.Equal(t, testDataLength, len(arr))

	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})

	for i := 1; i < testDataLength; i++ {
		require.True(t, arr[i] > arr[i-1])
	}

	next, err := s1.Next(context.Background())
	require.NoError(t, err)
	require.LessOrEqual(t, int(next), testDataLength+10100)
}

func Test_tddlSequence_Next_timeout(t *testing.T) {
	gormDB := newMockDB(t)
	require.NoError(t, gormDB.Exec("DELETE FROM sequences").Error)

	s1, err := newSequence(gormDB, &configs.TDDLConfig{
		Step:     10,
		SeqName:  testSeqName,
		StartNum: 10000,
	})
	require.NoError(t, err)
	t.Cleanup(s1.Close)

	next, err := s1.Next(context.Background())
	require.NoError(t, err)
	require.Equal(t, 10000, int(next))

	// set the deadline to 7 hours ago
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-7*time.Hour))
	cancel()
	// because the deadline is already expired, so the Next should return immediately
	// but in golang, select multi channels, the order is random, maybe the queue channel is selected first and return next value

	next, err = s1.Next(ctx)
	require.Equal(t, 0, int(next))
	require.ErrorIs(t, err, context.DeadlineExceeded)

	next, err = s1.Next(context.Background())
	require.NoError(t, err)
	require.Equal(t, 10001, int(next))
}

func Test_tddlSequence_renew_failed(t *testing.T) {
	gormDB := newMockDB(t)
	require.NoError(t, gormDB.Exec("DELETE FROM sequences").Error)

	s, err := newSequence(gormDB, &configs.TDDLConfig{
		Step:     1,
		SeqName:  testSeqName,
		StartNum: 10000,
	})
	require.NoError(t, err)

	sqlDB, _ := gormDB.DB()
	sqlDB.Close()

	go func() {
		time.Sleep(time.Second)
		// should retry 7 times
		require.Equal(t, 7, s.rateLimiter.Retries(context.Background(), s.clientID))
		s.Close()
	}()

	s.renew()
}
