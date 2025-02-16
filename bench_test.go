package atomicwaitnotify

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/edsrzf/mmap-go"
)

func runConsumer(id string, addr *uint32, tp *int64, max uint32) (delay int64) {
	for i := range max {
		err := WaitUint32(addr, i)
		if err != nil {
			log.Printf("[%s] wait failed: %v\n", id, err)
		}
		now := time.Now()
		delay += now.UnixNano() - atomic.LoadInt64(tp)
	}
	return
}

func BenchmarkThread(b *testing.B) {
	for i := 1; i < 16; i++ {
		delays := make([]int64, i)
		producerDelay := time.Duration(0)

		b.Run(fmt.Sprintf("P=1 C=%d", len(delays)), func(b *testing.B) {
			Counter := uint32(0)
			SendTP := int64(0)

			wg := sync.WaitGroup{}

			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() {
					b.ReportMetric(float64(producerDelay/time.Duration(b.N)), "prodns/msg")
				}()
				for {
					time.Sleep(1 * time.Millisecond)

					st := time.Now()
					Counter++
					atomic.StoreInt64(&SendTP, time.Now().UnixNano())
					NotifyAllUint32(&Counter)
					en := time.Now()

					producerDelay += en.Sub(st)

					if Counter >= uint32(b.N) {
						return
					}
				}
			}()

			wg.Add(len(delays))
			for consumerID := range delays {
				go func() {
					defer wg.Done()
					delays[consumerID] = runConsumer(fmt.Sprintf("%d", consumerID), &Counter, &SendTP, uint32(b.N))
				}()
			}

			wg.Wait()

			var sum float64 = 0
			for _, delay := range delays {
				sum += float64(delay)
			}
			sum = sum / float64(len(delays)) / float64(b.N)

			b.ReportMetric(float64(sum), "ns/msg")
		})
	}
}

func BenchmarkProcess(b *testing.B) {
	// Create a temporary file for shared memory.
	tmpFile, err := os.CreateTemp(os.TempDir(), "futex_shm")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if err := tmpFile.Truncate(4096); err != nil {
		b.Fatalf("truncate failed: %v", err)
	}

	// Map shared memory from temporary file.
	f, err := os.OpenFile(tmpFile.Name(), os.O_RDWR, 0600)
	if err != nil {
		b.Fatalf("failed to open shm file: %v", err)
	}
	defer f.Close()

	data, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		b.Fatalf("mmap failed: %v", err)
	}
	defer data.Unmap()

	addr := (*uint32)(unsafe.Pointer(&data[0]))
	tp := (*int64)(unsafe.Pointer(&data[8]))

	for i := 1; i < 10; i++ {
		delays := make([]int64, i)

		b.Run(fmt.Sprintf("P=1 C=%d", len(delays)), func(b *testing.B) {
			atomic.StoreUint32(addr, 0)

			cmd := exec.Command("go", "run", "benchmark_producer/main.go",
				"-shm", tmpFile.Name(),
				"-max", strconv.Itoa(b.N))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Start(); err != nil {
				b.Fatalf("failed to start producer: %v", err)
			}

			wg := sync.WaitGroup{}
			wg.Add(len(delays))
			for consumerID := range delays {
				go func() {
					defer wg.Done()
					delays[consumerID] = runConsumer(fmt.Sprintf("%d", consumerID), addr, tp, uint32(b.N))
				}()
			}
			wg.Wait()

			if err := cmd.Wait(); err != nil {
				b.Fatalf("producer process error: %v", err)
			}

			var sum float64 = 0
			for _, delay := range delays {
				sum += float64(delay)
			}
			sum = sum / float64(len(delays)) / float64(b.N)

			b.ReportMetric(float64(sum), "ns/msg")
		})

	}
}
