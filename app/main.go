package main

import (
	"errors"
	"io"
	"log"
	"net"
	// "os"

	"github.com/quockhanhcao/redish/internal/core/server"
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

func (p *Pool) Start() {
	for i := 0; i < len(p.workers); i++ {
		p.workers[i] = NewWorker(i+1, p.jobQueue)
		p.workers[i].Start()
	}
}

func (p *Pool) AddJob(job Job) {
	p.jobQueue <- job
}

func NewWorker(id int, jobQueue chan Job) *Worker {
	return &Worker{
		id:       id,
		jobQueue: jobQueue,
	}
}

func (w *Worker) Start() {
	go func() {
		for job := range w.jobQueue {
			log.Printf("Processing job from %s by worker %d\n", job.conn.RemoteAddr(), w.id)
			handleConnection(job.conn)
		}
	}()
}

func main() {
	// listener, err := net.Listen("tcp", "0.0.0.0:3000")
	// if err != nil {
	// 	log.Println("Failed to bind to port 3000")
	// 	os.Exit(1)
	// }
	// defer listener.Close()
	// log.Println("Server is listening on port 3000")
	// threadPool := NewPool(2)
	// threadPool.Start()
	// for {
	// 	conn, err := listener.Accept()
	// 	if err != nil {
	// 		log.Println("Error accepting connection: ", err.Error())
	// 		os.Exit(1)
	// 	}
	// 	threadPool.AddJob(Job{conn})
	// }
	server.StartServer()
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("Connection closed by client")
				return
			}
			log.Println("Error reading from connection: ", err.Error())
			return
		}
		// log.Println("Received data: ", string(buf[:n]))
		conn.Write([]byte("+PONG\r\n"))
	}
}
