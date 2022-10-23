package main

import "testing"

func Test_binaryToDecimal3Digit(t *testing.T) {
	parts := binaryDecimalConversion(byte(239))
	if parts[0] != 2 {
		t.Errorf("expected 2, got %d", parts[0])
	}
	if parts[1] != 3 {
		t.Errorf("expected 3, got %d", parts[1])
	}
	if parts[2] != 9 {
		t.Errorf("expected 9, got %d", parts[2])
	}
}

func Test_binaryToDecimal3DigitWith106(t *testing.T) {
	parts := binaryDecimalConversion(byte(106))
	if parts[0] != 1 {
		t.Errorf("expected 1, got %d", parts[0])
	}
	if parts[1] != 0 {
		t.Errorf("expected 0, got %d", parts[1])
	}
	if parts[2] != 6 {
		t.Errorf("expected 6, got %d", parts[2])
	}
}
func Test_binaryToDecimal3DigitWith100(t *testing.T) {
	parts := binaryDecimalConversion(byte(100))
	if parts[0] != 1 {
		t.Errorf("expected 1, got %d", parts[0])
	}
	if parts[1] != 0 {
		t.Errorf("expected 0, got %d", parts[1])
	}
	if parts[2] != 0 {
		t.Errorf("expected 0, got %d", parts[2])
	}
}

func Test_binaryToDecimal2Digit(t *testing.T) {
	parts := binaryDecimalConversion(byte(39))

	if parts[0] != 3 {
		t.Errorf("expected 3, got %d", parts[0])
	}
	if parts[1] != 9 {
		t.Errorf("expected 9, got %d", parts[1])
	}
}

func Test_binaryToDecimal1Digit(t *testing.T) {
	parts := binaryDecimalConversion(byte(3))

	if parts[0] != 3 {
		t.Errorf("expected 3, got %d", parts[0])
	}
}
