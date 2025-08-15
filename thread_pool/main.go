package threadpool

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

// thread pool
// first we need a Pool
// a pool consists of a thread safe queue of Job (can use built-in channel)
// an array of Worker
// each Job is actually a network connection to the server (in this case tcp server)
// let's give each worker an ID, and a queue to pull Job out to execute

type Pool struct {
	jobQueue chan Job
	workers  []*Worker
}

type Job struct {
	conn net.Conn
}

type Worker struct {
	id       int
	jobQueue chan Job
}

func NewPool(numWorkers int) *Pool {
	return &Pool{
		jobQueue: make(chan Job),
		workers:  make([]*Worker, numWorkers),
	}
}

func (p *Pool) AddJob(job Job) {
	p.jobQueue <- job
}

func (w *Worker) Start() {
	go func() {
		for job := range w.jobQueue {
			fmt.Printf("Processing job from %s by worker %d", job.conn.RemoteAddr(), w.id)
			handleConnection(job.conn)
		}
	}()
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Server is listening on port 6379")
	threadPool := NewPool(2)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		threadPool.AddJob(Job{conn})
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
