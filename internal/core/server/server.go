package server

import (
	"errors"
	"io"
	"log"
	"net"
	"os"
	"syscall"

	"github.com/quockhanhcao/redish/internal/core/command"
	"github.com/quockhanhcao/redish/internal/core/config"
	"github.com/quockhanhcao/redish/internal/core/executor"
	"github.com/quockhanhcao/redish/internal/core/io_multiplexing"
	"github.com/quockhanhcao/redish/internal/core/resp_parser"
)

func RunIoMultiplexingServer() {
	listener, err := net.Listen(config.PROTOCOL, config.PORT)
	if err != nil {
		log.Println("failed to bind to port 3000")
		os.Exit(1)
	}
	defer listener.Close()
	log.Println("server is listening on port 3000")

	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		log.Println("not a TCP connection")
		return
	}
	listenerFile, err := tcpListener.File()
	if err != nil {
		log.Println("failed to get file descriptor: ", err.Error())
		return
	}
	defer listenerFile.Close()
	listenerFD := int(listenerFile.Fd())

	ioMultiplexer, err := io_multiplexing.CreateIOMultiplexer()
	if err != nil {
		log.Fatal("failed to create I/O multiplexer: ", err.Error())
		return
	}
	defer ioMultiplexer.Close()

	err = ioMultiplexer.Monitor(io_multiplexing.Event{
		Fd: listenerFD,
		Op:      io_multiplexing.OperationRead})
	if err != nil {
		log.Println("failed to monitor listener: ", err.Error())
		return
	}

	var events = make([]io_multiplexing.Event, config.MAX_CONNECTIONS)
	for {
		events, err = ioMultiplexer.Wait()

		if err != nil {
			log.Print("error waiting for events: ", err.Error())
			continue
		}
		for _, event := range events {
			if event.Fd == listenerFD {
				log.Printf("new client is trying to connect")
				// set up new connection
				connFd, _, err := syscall.Accept(listenerFD)
				if err != nil {
					log.Println("err", err)
					continue
				}
				log.Printf("set up a new connection")
				// ask epoll to monitor this connection
				if err = ioMultiplexer.Monitor(io_multiplexing.Event{
					Fd: connFd,
					Op:      io_multiplexing.OperationRead,
				}); err != nil {
					log.Fatal(err)
				}
			} else {
				// parse the data here
				cmd, err := readCommand(event.Fd)
				if err != nil {
					if errors.Is(err, io.EOF) || errors.Is(err, syscall.ECONNRESET) {
						log.Print("client disconnected, closing fd ", event.Fd)
						syscall.Close(event.Fd)
						continue
					}
					continue
				}
				// execute the command here
				executor.ExecuteCommand(cmd, event.Fd)
			}
		}
	}
}

func readCommand(fd int) (*command.Command, error) {
	// redis commands are small
	// use small buffer
	var buffer = make([]byte, 512)
	readBytes, err := syscall.Read(fd, buffer)
	if err != nil {
		log.Print("error reading from fd: ", err.Error())
		return nil, err
	}
	if readBytes == 0 {
		// return nil, io.EOF
		return nil, io.EOF
	}
	return resp_parser.ParseCmd(buffer)
}
