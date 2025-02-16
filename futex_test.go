package atomicwaitnotify

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func BenchmarkFutex(b *testing.B) {
	for i := 1; i < 10; i++ {
		delays := make([]int64, i)

		b.Run(fmt.Sprintf("P=1 C=%d", len(delays)), func(b *testing.B) {
			Counter := uint32(0)
			SendTP := int64(0)

			wg := sync.WaitGroup{}

			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					time.Sleep(1 * time.Millisecond)
					Counter++
					atomic.StoreInt64(&SendTP, time.Now().UnixNano())
					NotifyAllUint32(&Counter)

					if Counter >= uint32(b.N) {
						return
					}
				}
			}()

			wg.Add(len(delays))
			for consumerID := range delays {
				go func() {
					defer wg.Done()
					for i := range uint32(b.N) {
						WaitUint32(&Counter, i)
						now := time.Now()
						delay := now.UnixNano() - SendTP
						delays[consumerID] += delay

						if Counter >= uint32(b.N) {
							return
						}
					}
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

func TestFutexWaitAndWake(t *testing.T) {
	var futexVal uint32 = 0

	// Start a goroutine that waits on the futex
	wg := 1
	go func() {
		defer func() {
			wg -= 1
		}()

		// Goroutine waits for futex to change
		err := WaitUint32(&futexVal, 0)
		if err != nil {
			t.Errorf("FutexWait failed: %v", err)
		}
	}()

	// Give the goroutine time to start and block
	time.Sleep(100 * time.Millisecond)

	// Ensure goroutine is still waiting by checking the waitgroup count
	if wg != 1 {
		t.Error("Goroutine should still be waiting on futex")
	}

	// Now, wake up the waiting goroutine by changing the futex value and waking
	futexVal = 1
	if err := NotifyOneUint32(&futexVal); err != nil {
		serr := uintptr(err.(syscall.Errno))
		t.Errorf("FutexWake failed: '%v' '%v'", err, serr)
	}

	// Wait for the goroutine to finish
	for {
		if wg == 0 {
			break
		}
		runtime.Gosched()
	}

	// Test passes if goroutine finishes successfully
}

func TestFutexTimeout(t *testing.T) {
	var futexVal uint32 = 0
	var wg sync.WaitGroup

	// Set a timeout of 200ms
	timeout := 200 * time.Millisecond

	// Start a goroutine that waits on the futex with a timeout
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Wait with a timeout
		err := WaitTimeoutUint32(&futexVal, 0, timeout)
		if err != syscall.ETIMEDOUT {
			t.Errorf("FutexWait with timeout failed, expected ETIMEDOUT but got: %v", err)
		}
	}()

	// Wait for the goroutine to finish
	wg.Wait()
}

func TestMultipleWaiters(t *testing.T) {
	var futexVal uint32 = 0
	var wg sync.WaitGroup
	numWaiters := 3

	// Start multiple goroutines that wait on the futex
	for i := 0; i < numWaiters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := WaitUint32(&futexVal, 0)
			if err != nil {
				t.Errorf("FutexWait failed: %v", err)
			}
		}()
	}

	// Give the goroutines time to start and block
	time.Sleep(100 * time.Millisecond)

	// Now wake all waiting goroutines
	futexVal = 1
	if err := NotifyAllUint32(&futexVal); err != nil {
		t.Errorf("FutexWake failed: %v", err)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Test passes if all goroutines finish successfully
}
