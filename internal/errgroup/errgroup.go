package errgroup

import "sync/atomic"

type ErrGroup struct {
	errCh chan error
	count atomic.Int64
}

func NewErrGroup() *ErrGroup {
	return &ErrGroup{
		errCh: make(chan error, 1),
	}
}

func (g *ErrGroup) Go(f func() error) {
	g.count.Add(1)

	go func() {
		defer func() {
			if g.count.Add(-1) == 0 {
				close(g.errCh)
			}
		}()

		if err := f(); err != nil {
			g.errCh <- err
		}
	}()
}

func (g *ErrGroup) Wait() error {
	return <-g.errCh
}
