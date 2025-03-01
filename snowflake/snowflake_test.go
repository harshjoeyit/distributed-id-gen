package main

import (
	"testing"
)

func BenchmarkSnowflakeID(b *testing.B) {
	s := NewSnowflake()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := s.GenerateNewID()
			if err != nil {
				b.Fatalf("Error generating ID: %v", err)
			}
		}
	})
}
