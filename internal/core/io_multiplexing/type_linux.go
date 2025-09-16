package io_multiplexing

import "syscall"

func (e Event) toEpollEvent() syscall.EpollEvent {
	var event uint32 = syscall.EPOLLIN
	if e.Op == OperationWrite {
		event = syscall.EPOLLOUT
	}
	return syscall.EpollEvent{
		Fd:     int32(e.Fd),
		Events: event,
	}
}

func toGenericEvent(event syscall.EpollEvent) Event {
	var operation Operation = OperationRead
	if event.Events == syscall.EPOLLOUT {
		operation = OperationWrite
	}
	return Event{
		Fd:        int(event.Fd),
		Op: operation,
	}
}
