package main

import (
	"testing"
)

func BenchmarkSnowflakeID(b *testing.B) {
	t := NewTicket()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := t.NewID()
			if err != nil {
				b.Fatalf("Error generating ID: %v", err)
			}
		}
	})
}
