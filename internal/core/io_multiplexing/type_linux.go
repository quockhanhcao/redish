package iomultiplexing

import "syscall"

func (e *Event) toEpollEvent() syscall.EpollEvent {
	op := syscall.EPOLLIN
	if e.Operation == OperationWrite {
		op = syscall.EPOLLOUT
	}
	return syscall.EpollEvent{
		Events: uint32(op),
		Fd:     int32(e.FileDescriptor),
	}
}

func toGenericEvent(event syscall.EpollEvent) Event {
	var operation = OperationRead
	if event.Events == syscall.EPOLLOUT {
		operation = OperationWrite
	}
	return Event{
		FileDescriptor: int(event.Fd),
		Operation:      operation,
	}
}
