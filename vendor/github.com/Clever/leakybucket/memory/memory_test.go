package memory

import (
	"testing"

	"github.com/Clever/leakybucket/test"
)

func TestCreate(t *testing.T) {
	test.CreateTest(New())(t)
}

func TestAdd(t *testing.T) {
	test.AddTest(New())(t)
}

func TestAddResetTest(t *testing.T) {
	test.AddResetTest(New())(t)
}

func TestThreadSafeAdd(t *testing.T) {
	test.ThreadSafeAddTest(New())(t)
}

func TestReset(t *testing.T) {
	test.AddResetTest(New())(t)
}

func TestFindOrCreate(t *testing.T) {
	test.FindOrCreateTest(New())(t)
}

func TestBucketInstanceConsistencyTest(t *testing.T) {
	test.BucketInstanceConsistencyTest(New())(t)
}
