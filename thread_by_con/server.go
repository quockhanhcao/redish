package threadbycon

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Server is listening on port 6379")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
            if errors.Is(err, io.EOF) {
                fmt.Println("Connection closed by client")
                return
            }
			fmt.Println("Error reading from connection: ", err.Error())
			return
		}
		fmt.Println("Received data: ", string(buf[:n]))
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n+PONG\r\n"))
	}
}
