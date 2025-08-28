package iomultiplexing

const (
	OperationRead  = 1
	OperationWrite = 2
)

type Operation uint32

type Event struct {
	FileDescriptor int
	Operation      Operation
}

type IOMultiplexer interface {
	Monitor(event Event) error
	Wait() ([]Event, error)
	Close() error
}
