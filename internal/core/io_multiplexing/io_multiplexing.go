package iomultiplexing

const (
	OperationRead  = 1
	OperationWrite = 2
)

type Event struct {
	FileDescriptor int
	Operation      int
}

type IOMultiplexer interface {
	Monitor(Event) error
	Wait() ([]Event, error)
	Close() error
}
