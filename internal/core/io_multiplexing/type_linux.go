package iomultiplexing

import "syscall"

func (e Event) toEpollEvent() syscall.EpollEvent {
	var event uint32 = syscall.EPOLLIN
	if e.Operation == OperationWrite {
		event = syscall.EPOLLOUT
	}
	return syscall.EpollEvent{
		Fd:     int32(e.FileDescriptor),
		Events: uint32(event),
	}
}

func toGenericEvent(event syscall.EpollEvent) Event {
	var operation Operation = OperationRead
	if event.Events == syscall.EPOLLOUT {
		operation = OperationWrite
	}
	return Event{
		FileDescriptor: int(event.Fd),
		Operation:      operation,
	}
}
