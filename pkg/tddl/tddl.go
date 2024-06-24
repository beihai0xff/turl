// Package tddl provides the tddl sequence number generator
package tddl

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/plugin/optimisticlock"

	"github.com/beihai0xff/turl/configs"
	"github.com/beihai0xff/turl/pkg/workqueue"
)

const (
	// retryInterval is the retry start interval when update tddl failed
	retryInterval = 10 * time.Millisecond
)

var (
	// ErrStepTooSmall is the error of step too small
	ErrStepTooSmall      = errors.New("step must be greater than 0")
	_               TDDL = (*tddlSequence)(nil)
)

// TDDL is the interface of tddl
type TDDL interface {
	// Next returns the next sequence number
	Next(ctx context.Context) (uint64, error)
	// Close closes the tddl
	Close()
	// Renew()
}

// Sequence is the table of sequence
type Sequence struct {
	gorm.Model
	Name     string `gorm:"type:VARCHAR(500);not null;uniqueIndex" json:"name"`
	Sequence uint64 `gorm:"type:bigint;not null" json:"sequence"`
	Version  optimisticlock.Version
}

// TableName returns the table name of the Sequence model
func (Sequence) TableName() string {
	return "sequences"
}

type tddlSequence struct {
	clientID string
	conn     *gorm.DB

	// rowID is the row primary key of the sequence
	rowID uint

	step uint64
	max  uint64
	curr atomic.Uint64

	wg sync.WaitGroup

	stop chan struct{}
	// TODO: use a buffer channel to avoid blocking
	queue chan uint64

	rateLimiter workqueue.RateLimiter[any]
}

// New returns a new tddl implementation
func New(conn *gorm.DB, c *configs.TDDLConfig) (TDDL, error) {
	return newSequence(conn, c)
}

// newSequence creates a new tddlSequence instance
func newSequence(conn *gorm.DB, c *configs.TDDLConfig) (*tddlSequence, error) {
	if c.Step < 1 {
		return nil, ErrStepTooSmall
	}

	s := tddlSequence{
		clientID:    uuid.NewString(),
		conn:        conn,
		step:        c.Step,
		wg:          sync.WaitGroup{},
		stop:        make(chan struct{}),
		queue:       make(chan uint64),
		rateLimiter: workqueue.NewItemExponentialFailureRateLimiter[any](retryInterval, time.Minute),
	}

	if err := s.getRowID(c.SeqName, c.StartNum); err != nil {
		return nil, err
	}

	// filling the curr and max
	s.renew()

	go s.worker()
	s.wg.Add(1)

	return &s, nil
}

func (s *tddlSequence) getRowID(seqName string, startNum uint64) error {
	var seq Sequence

	res := s.conn.Where("name = ?", seqName).Take(&seq)
	if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return res.Error
	}

	if res.RowsAffected == 1 {
		s.rowID = seq.Model.ID
		return nil
	}

	id, err := s.createRecord(seqName, startNum)
	if err != nil {
		return err
	}

	s.rowID = id

	return nil
}

func (s *tddlSequence) createRecord(seqName string, startNum uint64) (uint, error) {
	seq := Sequence{Name: seqName, Sequence: startNum}

	res := s.conn.Create(&seq)
	if res.Error != nil {
		return 0, res.Error
	}

	return seq.Model.ID, nil
}

// func (s *tddlSequence) Renew() {
// 	if !s.mu.TryLock() { // already in renewing
// 		s.wg.Wait()
// 		return
// 	}
// 	defer s.mu.Unlock()
//
// 	go s.renew()
// 	s.wg.Add(1)
// 	s.wg.Wait()
// }

// renew function renews the sequence number
// should be called in a single goroutine
func (s *tddlSequence) renew() {
	defer s.rateLimiter.Forget(s.clientID) // forget the retry times

	var seq = Sequence{}

	for {
		select {
		case <-s.stop: // receive stop signal
			return
		default:
		}

		seq = Sequence{}
		res := s.conn.Where("id = ?", s.rowID).Take(&seq) // update the sequence with cas

		if res.Error == nil {
			res = s.conn.Model(&seq).Update("sequence", seq.Sequence+s.step)
			if res.Error == nil && res.RowsAffected == 1 {
				break
			}

			slog.Debug("cas sequence failed")
		} else {
			slog.Warn("get sequence failed", slog.String("error", res.Error.Error()))
		}

		time.Sleep(s.rateLimiter.When(s.clientID))
	}

	s.curr.Store(seq.Sequence - s.step)
	s.max = seq.Sequence
	slog.Debug("renew tddl sequence success", slog.Group("sequence",
		slog.String("clientID", s.clientID),
		slog.String("name", seq.Name),
		slog.Uint64("maxSequence", seq.Sequence),
		slog.Uint64("currSequence", s.curr.Load()),
		slog.Int("retryTimes", s.rateLimiter.Retries(s.clientID)),
	))
}

// Next returns the next sequence number
func (s *tddlSequence) Next(ctx context.Context) (uint64, error) {
	// if ctx is already done, return immediately
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case next := <-s.queue:
			return next, nil
		}
	}
}

func (s *tddlSequence) worker() {
	defer s.wg.Done()

	next := s.curr.Load()

	for {
		select {
		case s.queue <- next:
			next = s.curr.Add(1)
			for next >= s.max { // the serial number has been exhausted
				s.renew()
				next = s.curr.Load()
			}
		case <-s.stop:
			return
		}
	}
}

// Close closes the tddl
func (s *tddlSequence) Close() {
	close(s.stop)
	s.wg.Wait()
}
