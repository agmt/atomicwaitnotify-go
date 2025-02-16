package main

import (
	"flag"
	"log"
	"os"
	"sync/atomic"
	"time"
	"unsafe"

	"atomicwaitnotify"

	"github.com/edsrzf/mmap-go"
)

func runProducer(addr *uint32, tp *int64, max uint32) {
	for i := range max {
		time.Sleep(1 * time.Millisecond)
		i++
		atomic.StoreInt64(tp, time.Now().UnixNano())
		atomic.StoreUint32(addr, i)
		err := atomicwaitnotify.NotifyAllUint32(addr)
		if err != nil {
			log.Printf("notify failed: %v\n", err)
		}
	}
}

func main() {
	shmPath := flag.String("shm", "", "path to shared memory file")
	max := flag.Uint("max", 1000, "number of updates to perform")
	flag.Parse()
	if *shmPath == "" {
		log.Fatal("missing -shm file path")
	}

	f, err := os.OpenFile(*shmPath, os.O_RDWR, 0600)
	if err != nil {
		log.Fatalf("failed to open shm file: %v", err)
	}
	defer f.Close()

	if err != nil {
		log.Fatalf("truncate failed: %v", err)
	}

	data, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		log.Fatalf("mmap failed: %v", err)
	}
	defer data.Unmap()

	runProducer((*uint32)(unsafe.Pointer(&data[0])), (*int64)(unsafe.Pointer(&data[8])), uint32(*max))
}
