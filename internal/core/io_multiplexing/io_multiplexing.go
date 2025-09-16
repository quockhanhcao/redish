package io_multiplexing

const (
	OperationRead  = 1
	OperationWrite = 2
)

type Operation uint32

type Event struct {
	Fd int
	Op Operation
}

type IOMultiplexer interface {
	Monitor(event Event) error
	Wait() ([]Event, error)
	Close() error
}
