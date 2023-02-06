package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/enchant97/file-sync-protocol/prototypes/proto-3/core"
)

func processClientPathInput(filePaths []string) []string {
	if len(filePaths) == 1 {
		if info, _ := os.Stat(filePaths[0]); info.IsDir() {
			paths := make([]string, 0)
			filepath.Walk(filePaths[0], func(path string, info fs.FileInfo, err error) error {
				if !info.IsDir() {
					paths = append(paths, path)
				}
				return nil
			})
			return paths
		}
	}
	return filePaths
}

func main() {
	args := os.Args

	mtu := uint32(1500)

	if value, isSet := os.LookupEnv("NET_MTU"); isSet {
		value, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			log.Fatalln("NET_MTU invalid")
		}
		mtu = uint32(value)
	}

	mtu -= core.Ipv4ReservedBytes

	log.Printf("receive MTU = '%d'\n", mtu)

	if args[1] == "server" {
		log.Println("starting server...")
		server(args[2], mtu)
	}
	if args[1] == "client" {
		chunks_per_block := uint(50)
		if value, isSet := os.LookupEnv("CHUNKS_PER_BLOCK"); isSet {
			value, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				log.Fatalln("CHUNKS_PER_BLOCK invalid")
			}
			chunks_per_block = uint(value)
		}
		log.Printf("chunks per block = '%d'\n", chunks_per_block)

		log.Println("starting client...")
		client(args[2], mtu, chunks_per_block, processClientPathInput(args[3:]))
	}
	log.Println("done")
}
