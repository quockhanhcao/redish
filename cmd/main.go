package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/quockhanhcao/redish/internal/core/config"
	"github.com/quockhanhcao/redish/internal/core/server"
)

func main() {
	dir := flag.String("dir", ".", "Directory path")
	filename := flag.String("dbfilename", "dump.rdb", "RBD file name")
	flag.Parse()

	info, err := os.Stat(*dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// create the directory
		} else {
			fmt.Fprintf(os.Stderr, "Error accessing directory '%s': %v\n", *dir, err)
			os.Exit(1)
		}
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: '%s' is a file, not a directory\n", *dir)
		os.Exit(1)
	}

	cfg := config.PersistenceConfiguration{
		Directory:  *dir,
		DBFileName: *filename,
	}
	server.RunIoMultiplexingServer(cfg)
}
