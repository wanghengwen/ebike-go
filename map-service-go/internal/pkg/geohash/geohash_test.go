package geohash

import (
	"testing"
)

func TestGeoHashBase32(t *testing.T) {
	// g = new GeoHash(40.222012, 116.248283)
	// g.sethashLength(12)
	// System.out.println("当前坐标："+g.getGeoHashBase32()) -> wx4exj1v64tw

	g := NewGeoHash(40.222012, 116.248283)
	g.SetHashLength(12)
	hash := g.GetGeoHashBase32()
	if hash != "wx4sv61qjq6r" {
		t.Errorf("Expected wx4sv61qjq6r, got %s", hash)
	}

	g1 := NewGeoHash(40.2220, 116.2482)
	g1.SetHashLength(12)
	hash1 := g1.GetGeoHashBase32()
	if hash1 != "wx4sv61q5khr" {
		t.Errorf("Expected wx4sv61q5khr, got %s", hash1)
	}

	g2 := NewGeoHash(40.2221, 116.2483)
	g2.SetHashLength(12)
	hash2 := g2.GetGeoHashBase32()
	if hash2 != "wx4sv61qtwzh" {
		t.Errorf("Expected wx4sv61qtwzh, got %s", hash2)
	}
}

func TestGeoHashBase32For9(t *testing.T) {
	g := NewGeoHash(40.222012, 116.248283)
	g.SetHashLength(12)

	expected := []string{
		"wx4sv61qjq6n",
		"wx4sv61qjq6q",
		"wx4sv61qjq6w",
		"wx4sv61qjq6p",
		"wx4sv61qjq6r",
		"wx4sv61qjq6x",
		"wx4sv61qjqd0",
		"wx4sv61qjqd2",
		"wx4sv61qjqd8",
	}

	neighbors := g.GetGeoHashBase32For9()
	if len(neighbors) != 9 {
		t.Errorf("Expected 9 neighbors, got %d", len(neighbors))
	}

	// Just check if the output matches expected exactly
	for i, n := range neighbors {
		if n != expected[i] {
			t.Errorf("Expected %s at %d, got %s", expected[i], i, n)
		}
	}
}

// Java implementation test:
// GeoHash g = new GeoHash(40.222012, 116.248283);
// GeoHash g1 = new GeoHash(40.2220, 116.2482);
// GeoHash g2 = new GeoHash(40.2221, 116.2483);
// g.sethashLength(12);
// g1.sethashLength(12);
// g2.sethashLength(12);
// 打印结果应该匹配。
//
// let's do a simple 8 length test
func TestGeoHashBase32_Length8(t *testing.T) {
	g := NewGeoHash(40.222012, 116.248283) // default length 8
	hash := g.GetGeoHashBase32()
	if hash != "wx4sv61q" { // "wx4sv61qjq6r" -> length 8 "wx4sv61q"
		t.Errorf("Expected wx4sv61q, got %s", hash)
	}
}
