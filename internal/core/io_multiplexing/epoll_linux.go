//go:build linux

package io_multiplexing

import (
	"syscall"

	"github.com/quockhanhcao/redish/internal/core/config"
)

type Epoll struct {
	fd            int
	epollEvents   []syscall.EpollEvent
	genericEvents []Event
}

// Close implements IOMultiplexer.
func (e *Epoll) Close() error {
	return syscall.Close(e.fd)
}

// Monitor implements IOMultiplexer.
func (e *Epoll) Monitor(event Event) error {
	epollEvent := event.toEpollEvent()
	return syscall.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, event.Fd, &epollEvent)
}

// Wait implements IOMultiplexer.
func (e *Epoll) Wait() ([]Event, error) {
	n, err := syscall.EpollWait(e.fd, e.epollEvents, -1)
	if err != nil {
		return nil, err
	}
	for i := range n {
		e.genericEvents[i] = toGenericEvent(e.epollEvents[i])
	}
	return e.genericEvents[:n], nil
}

func CreateIOMultiplexer() (*Epoll, error) {
	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &Epoll{
		fd:            epollFD,
		epollEvents:   make([]syscall.EpollEvent, config.MAX_CONNECTIONS),
		genericEvents: make([]Event, config.MAX_CONNECTIONS),
	}, nil
}
