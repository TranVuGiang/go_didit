package godidit_test

import (
	"testing"

	godidit "github.com/TranVuGiang/go_didit"
)

func TestToAlpha3(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want string
	}{
		{"VN", "VNM"},
		{"US", "USA"},
		{"vn", "VNM"}, // case-insensitive
		{"ID", "IDN"},
		{"SG", "SGP"},
		{"KH", "KHM"},
		{"TH", "THA"},
		{"GB", "GBR"},
		{"BR", "BRA"},
		{"UNKNOWN", "UNKNOWN"}, // passthrough when not found
		{"",        ""},        // empty passthrough
	}

	for _, tc := range cases {
		got := godidit.ToAlpha3(tc.in)
		if got != tc.want {
			t.Errorf("ToAlpha3(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestToGenderCode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want string
	}{
		{"male", "M"},
		{"Male", "M"},
		{"MALE", "M"},
		{"m", "M"},
		{"female", "F"},
		{"Female", "F"},
		{"f", "F"},
		{"other", "X"},
		{"non-binary", "X"},
		{"x", "X"},
		{"unknown", "X"},
		{"CUSTOM", "CUSTOM"}, // passthrough when not recognised
		{"", ""},
	}

	for _, tc := range cases {
		got := godidit.ToGenderCode(tc.in)
		if got != tc.want {
			t.Errorf("ToGenderCode(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestToDocTypeCode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want string
	}{
		{"PASSPORT", "P"},
		{"passport", "P"},
		{"P", "P"},
		{"ID_CARD", "ID"},
		{"ID CARD", "ID"},
		{"NATIONAL_ID", "ID"},
		{"ID", "ID"},
		{"DRIVER_LICENSE", "DL"},
		{"DRIVING_LICENSE", "DL"},
		{"DL", "DL"},
		{"RESIDENCE_PERMIT", "RP"},
		{"RP", "RP"},
		{"OTHER", "OTHER"}, // passthrough when not recognised
		{"", ""},
	}

	for _, tc := range cases {
		got := godidit.ToDocTypeCode(tc.in)
		if got != tc.want {
			t.Errorf("ToDocTypeCode(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
