package main

import (
	"math"
	"testing"
)

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		in      string
		want    rgb
		wantErr bool
	}{
		{"#ff0000", rgb{1, 0, 0}, false},
		{"#00ff00", rgb{0, 1, 0}, false},
		{"#0000ff", rgb{0, 0, 1}, false},
		{"#f00", rgb{1, 0, 0}, false},
		{"ffffff", rgb{1, 1, 1}, false},
		{"#zzzzzz", rgb{}, true},
		{"#ff00", rgb{}, true},
		{"", rgb{}, true},
	}
	for _, tt := range tests {
		got, err := parseHexColor(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseHexColor(%q) error = nil, want error", tt.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseHexColor(%q) error = %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseHexColor(%q) = %+v, want %+v", tt.in, got, tt.want)
		}
	}
}

func TestClamp01(t *testing.T) {
	tests := []struct{ in, want float64 }{
		{-1, 0}, {0, 0}, {0.5, 0.5}, {1, 1}, {2, 1},
	}
	for _, tt := range tests {
		if got := clamp01(tt.in); got != tt.want {
			t.Errorf("clamp01(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestScaleByte(t *testing.T) {
	tests := []struct {
		in   float64
		want byte
	}{
		{0, 0}, {1, 255}, {0.5, 128}, {-1, 0}, {2, 255},
	}
	for _, tt := range tests {
		if got := scaleByte(tt.in); got != tt.want {
			t.Errorf("scaleByte(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestHSVToRGBCorners(t *testing.T) {
	tests := []struct {
		name string
		h    float64
		want rgb
	}{
		{"red", 0.0 / 6, rgb{1, 0, 0}},
		{"yellow", 1.0 / 6, rgb{1, 1, 0}},
		{"green", 2.0 / 6, rgb{0, 1, 0}},
		{"cyan", 3.0 / 6, rgb{0, 1, 1}},
		{"blue", 4.0 / 6, rgb{0, 0, 1}},
		{"magenta", 5.0 / 6, rgb{1, 0, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hsvToRGB(tt.h, 1, 1)
			if math.Abs(got.R-tt.want.R) > 1e-9 || math.Abs(got.G-tt.want.G) > 1e-9 || math.Abs(got.B-tt.want.B) > 1e-9 {
				t.Errorf("hsvToRGB(%v, 1, 1) = %+v, want %+v", tt.h, got, tt.want)
			}
		})
	}
}

func TestHSVToRGBZeroSaturationIsGray(t *testing.T) {
	got := hsvToRGB(0.42, 0, 0.7)
	want := rgb{0.7, 0.7, 0.7}
	if got != want {
		t.Errorf("hsvToRGB(0.42, 0, 0.7) = %+v, want %+v", got, want)
	}
}
