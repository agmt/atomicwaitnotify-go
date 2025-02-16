package atomicwaitnotify

import (
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"
)

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
