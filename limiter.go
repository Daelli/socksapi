package main

import "sync"

type LIMITER struct {
	pool chan struct{}
	wg   sync.WaitGroup
}

func NewLimiter() *LIMITER {
	return &LIMITER{
		pool: make(chan struct{}, 100),
	}
}

func (l *LIMITER) Run(job func()) {
	l.pool <- struct{}{}

	l.wg.Go(func() {
		defer func() { <-l.pool }()

		job()
	})
}

func (l *LIMITER) Wait() {
	l.wg.Wait()
}
