package server

import (
	"log"
	"net"
	"os"

	"github.com/quockhanhcao/redish/internal/core/config"
	iomultiplexing "github.com/quockhanhcao/redish/internal/core/io_multiplexing"
)

func StartServer() {
	listener, err := net.Listen(config.PROTOCOL, config.PORT)
	if err != nil {
		log.Println("Failed to bind to port 3000")
		os.Exit(1)
	}
	defer listener.Close()
	log.Println("Server is listening on port 3000")

	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		log.Println("Not a TCP connection")
		return
	}
	listenerFile, err := tcpListener.File()
	if err != nil {
		log.Println("Failed to get file descriptor: ", err.Error())
		return
	}
	defer listenerFile.Close()
	listenerFD := int(listenerFile.Fd())

	ioMultiplexer, err := iomultiplexing.CreateIOMultiplexer()
	if err != nil {
		log.Println("Failed to create I/O multiplexer: ", err.Error())
		return
	}
	defer ioMultiplexer.Close()

	err = ioMultiplexer.Monitor(iomultiplexing.Event{
		FileDescriptor: listenerFD,
		Operation:      iomultiplexing.OperationRead})
	if err != nil {
		log.Println("Failed to monitor listener: ", err.Error())
		return
	}
}
