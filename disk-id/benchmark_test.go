package main

import "testing"

func BenchmarkGetNewID(b *testing.B) {
	c := NewCounter("counter.txt")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := c.GetNewID()
			if err != nil {
				b.Fatalf("Error generating ID: %v", err)
			}
		}
	})
}
