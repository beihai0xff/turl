// Package tddl provides the tddl sequence number generator
package tddl

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/optimisticlock"
)

// updateTDDLFailedSleepTime is the sleep time when update tddl failed
const updateTDDLFailedSleepTime = 10 * time.Millisecond

var (
	// ErrStepTooSmall is the error of step too small
	ErrStepTooSmall      = errors.New("step must be greater than 0")
	_               TDDL = (*tddlSequence)(nil)
)

// TDDL is the interface of tddl
type TDDL interface {
	// Next returns the next sequence number
	Next(ctx context.Context) uint64
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

type tddlSequence struct {
	conn *gorm.DB

	// rowID is the row primary key of the sequence
	rowID uint

	curr atomic.Uint64
	step uint64
	max  uint64

	wg sync.WaitGroup

	sendq, stop chan struct{}
	recvq       chan uint64
}

// New returns a new tddl implementation
func New(conn *gorm.DB, c *Config) (TDDL, error) {
	return newSequence(conn, c)
}

// newSequence creates a new tddlSequence instance
func newSequence(conn *gorm.DB, c *Config) (*tddlSequence, error) {
	if c.Step < 1 {
		return nil, ErrStepTooSmall
	}

	s := tddlSequence{
		conn:  conn,
		step:  c.Step,
		wg:    sync.WaitGroup{},
		sendq: make(chan struct{}),
		stop:  make(chan struct{}),
		recvq: make(chan uint64),
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

func (s *tddlSequence) renew() {
	seq := Sequence{}

	for {
		res := s.conn.Where("id = ?", s.rowID).Take(&seq)

		if res.Error != nil { // update the sequence with cas
			slog.Warn("failed to get sequence", slog.String("error", res.Error.Error()))
			time.Sleep(updateTDDLFailedSleepTime)

			continue
		}

		res = s.conn.Model(&seq).Update("sequence", seq.Sequence+s.step)
		if res.Error == nil && res.RowsAffected == 1 {
			break
		}
	}

	s.curr.Store(seq.Sequence - s.step - 1)
	s.max = seq.Sequence
	slog.Info("renew tddl sequence success", slog.Group("sequence",
		slog.String("name", seq.Name),
		slog.Uint64("maxSequence", seq.Sequence),
		slog.Uint64("currSequence", s.curr.Load()),
	))
}

// Next returns the next sequence number
func (s *tddlSequence) Next(ctx context.Context) uint64 {
	s.sendq <- struct{}{}

	for {
		select {
		case <-ctx.Done():
			return 0
		case next := <-s.recvq:
			return next
		}
	}
}

func (s *tddlSequence) worker() {
	defer s.wg.Done()

	for {
		select {
		case <-s.sendq:
			next := s.curr.Add(1)
			for next >= s.max { // the serial number has been exhausted
				s.renew()
				next = s.curr.Add(1)
			}

			s.recvq <- next
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
