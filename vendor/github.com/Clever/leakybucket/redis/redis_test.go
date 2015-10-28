package redis

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Clever/leakybucket"
	"github.com/Clever/leakybucket/test"
)

func getLocalStorage() *Storage {
	storage, err := New("tcp", os.Getenv("REDIS_URL"))
	if err != nil {
		panic(err)
	}
	return storage
}

func flushDb() {
	storage := getLocalStorage()
	conn := storage.pool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHDB")
	if err != nil {
		panic(err)
	}
}

func TestInvalidHost(t *testing.T) {
	_, err := New("tcp", "localhost:6378")
	if err == nil {
		t.Fatalf("expected error connecting to invalid host")
	}
}

func TestCreate(t *testing.T) {
	flushDb()
	test.CreateTest(getLocalStorage())(t)
}

func TestAdd(t *testing.T) {
	flushDb()
	test.AddTest(getLocalStorage())(t)
}

func TestThreadSafeAdd(t *testing.T) {
	// Redis Add is not thread safe. If you run this, the test should fail because it never received
	// ErrorFull. It's not thread safe because we don't atomically check the state of the bucket and
	// increment.
	t.Skip()
	flushDb()
	test.ThreadSafeAddTest(getLocalStorage())(t)
}

func TestReset(t *testing.T) {
	flushDb()
	test.AddResetTest(getLocalStorage())(t)
}

func TestFindOrCreate(t *testing.T) {
	flushDb()
	test.FindOrCreateTest(getLocalStorage())(t)
}

func TestBucketInstanceConsistencyTest(t *testing.T) {
	flushDb()
	test.BucketInstanceConsistencyTest(getLocalStorage())(t)
}

// One implementation of redis leaky bucket had a bug where very fast access could result in us
// creating buckets without a TTL on them. This test was reliably able to reproduce this bug.
func TestFastAccess(t *testing.T) {
	flushDb()
	s := getLocalStorage()
	bucket, err := s.Create("testbucket", 10, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	hold := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-hold
			if _, err := bucket.Add(1); err != nil && err != leakybucket.ErrorFull {
				t.Fatal(err)
			}
		}()
	}
	close(hold) // Let all concurrent requests start
	wg.Wait()   // Wait for all concurrent requests to finish

	conn := s.pool.Get()
	defer conn.Close()

	if exists, err := conn.Do("GET", "testbucket"); err != nil {
		t.Fatal(err)
	} else if exists == nil {
		return
	}
	ttl, err := conn.Do("PTTL", "testbucket")
	if err != nil {
		t.Fatal(err)
	}
	if ttl.(int64) == -1 {
		t.Fatal("no ttl set on bucket")
	}
}
