package fifo

import (
	"reflect"
	"testing"
)

func TestFIFO16_Contiguous(t *testing.T) {
	t.Run("empty queue", func(t *testing.T) {
		q := NewFIFO16(5)
		s := q.Contiguous()
		if s != nil {
			t.Errorf("Expected nil slice for empty queue, got %v", s)
		}
	})

	t.Run("not full and not wrapped", func(t *testing.T) {
		q := NewFIFO16(5)
		q.Enqueue(1)
		q.Enqueue(2)
		q.Enqueue(3)

		s := q.Contiguous()
		expected := []int16{1, 2, 3}
		if !reflect.DeepEqual(s, expected) {
			t.Errorf("Expected %v, got %v", expected, s)
		}
		// Ensure state is not changed unnecessarily
		if q.head != 0 || q.tail != 3 {
			t.Errorf(
				"Expected head=0, tail=3, got head=%d, tail=%d",
				q.head,
				q.tail,
			)
		}
	})

	t.Run("full and wrapped", func(t *testing.T) {
		q := NewFIFO16(5) // cap=5
		for i := int16(1); i <= 7; i++ {
			q.Enqueue(i) // Enqueue 1,2,3,4,5 (full), 6 (drops 1), 7 (drops 2)
		}
		// Queue should contain [3, 4, 5, 6, 7]
		// Internal state: head=2, tail=2, buf=[6,7,3,4,5]
		exectedBefore := []int16{6, 7, 3, 4, 5}
		if !reflect.DeepEqual(q.buf, exectedBefore) {
			t.Errorf(
				"Expected internal buffer to be %v before contiguous, got %v",
				exectedBefore,
				q.buf,
			)
		}

		s := q.Contiguous()
		expected := []int16{3, 4, 5, 6, 7}
		if !reflect.DeepEqual(s, expected) {
			t.Fatalf("Expected %v, got %v", expected, s)
		}

		// Check internal state after Contiguous()
		if q.head != 0 || q.tail != 0 {
			t.Errorf(
				"Expected head=0, tail=0 after contiguous, got head=%d, tail=%d",
				q.head,
				q.tail,
			)
		}
		if !reflect.DeepEqual(q.buf, expected) {
			t.Errorf(
				"Expected internal buffer to be %v, got %v",
				expected,
				q.buf,
			)
		}
	})

	t.Run("subsequent operations after contiguous", func(t *testing.T) {
		q := NewFIFO16(5)
		for i := int16(1); i <= 7; i++ {
			q.Enqueue(i)
		}

		// Make it contiguous: state becomes head=0, tail=0, buf=[3,4,5,6,7]
		q.Contiguous()

		// Dequeue
		v, ok := q.Dequeue()
		if !ok || v != 3 {
			t.Fatalf("Expected to dequeue 3, got %d", v)
		}

		// Check state after dequeue
		// head=1, tail=0, count=4, elements are [4,5,6,7], underlying buf=[3,4,5,6,7]
		if q.head != 1 || q.tail != 0 || q.count != 4 {
			t.Fatalf(
				"State incorrect after dequeue: head=%d, tail=%d, count=%d",
				q.head,
				q.tail,
				q.count,
			)
		}

		// Enqueue
		q.Enqueue(8)
		// underlying buf should now be [8,4,5,6,7], head=1, tail=1, count=5
		expectedBuf := []int16{8, 4, 5, 6, 7}
		if !reflect.DeepEqual(q.buf, expectedBuf) {
			t.Errorf("Expected buffer %v after enqueue, got %v", expectedBuf, q.buf)
		}
		finalSlice := q.Contiguous()
		expected := []int16{4, 5, 6, 7, 8}
		if !reflect.DeepEqual(finalSlice, expected) {
			t.Errorf("Expected final slice %v, got %v", expected, finalSlice)
		}
	})
}
