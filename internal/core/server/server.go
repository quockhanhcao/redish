package server

import (
	"errors"
	"io"
	"log"
	"net"
	"os"
	"syscall"

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

	var events = make([]iomultiplexing.Event, config.MAX_CONNECTIONS)
	for {
		events, err = ioMultiplexer.Wait()

		if err != nil {
			log.Print("error waiting for events: ", err.Error())
			continue
		}
		for _, event := range events {
			if event.FileDescriptor == listenerFD {
				conn, err := tcpListener.Accept()
				if err != nil {
					log.Print("error accepting connection: ", err.Error())
					continue
				}
				connectionFD, err := conn.(*net.TCPConn).File()
				if err != nil {
					log.Print("error getting connection file descriptor: ", err.Error())
					conn.Close()
					continue
				}
				err = ioMultiplexer.Monitor(iomultiplexing.Event{
					FileDescriptor: int(connectionFD.Fd()),
					Operation:      iomultiplexing.OperationRead})
				if err != nil {
					log.Print("error monitoring connection: ", err.Error())
					conn.Close()
					continue
				}
			} else {
				// parse the data here
				command, err := readCommand(event.FileDescriptor)
				if err != nil {
					if errors.Is(err, io.EOF) || errors.Is(err, syscall.ECONNRESET) {
						log.Print("connection closed by client")
					} else {
						log.Print("error reading command: ", err.Error())
					}
				}

				// execute the command here
			}
		}
	}
}

func readCommand(fd int) (command, error) {
	// redis commands are small
	// use small buffer
	var buffer = make([]byte, 512)
	readBytes, err := syscall.Read(fd, buffer)
	if err != nil {
		log.Print("error reading from fd: ", err.Error())
		return nil, err
	}
	if readBytes == 0  {
		return nil, io.EOF
	}
	return parseCommand(buffer[:readBytes])
}
