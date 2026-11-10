package segment_test

import (
	"slices"
	"testing"

	"campaigncomposer/internal/seed"
	"campaigncomposer/internal/segment"
)

func TestDefault_CitiesMatchSeed(t *testing.T) {
	f, ok := segment.Default().Field("city")
	if !ok {
		t.Fatal("no city field")
	}
	want := append(seed.TRCities(), seed.USCities()...)
	if !slices.Equal(f.Enum, want) {
		t.Fatalf("city enum = %v, want %v", f.Enum, want)
	}
}
