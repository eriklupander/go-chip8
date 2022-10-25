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

func Test_Mask(t *testing.T) {
	v1 := uint16(4090)
	v2 := uint16(10)
	v3 := v1 + v2
	if v3 != uint16(4100) {
		t.Errorf("expected v1+v2 to equal 4100")
	}
	v4 := v3 % (0x1000)
	v5 := v3 & 0xFFF
	if v4 != uint16(4) {
		t.Errorf("expected v3 mod 0xFF to equal 4, was %d", v4)
	}
	if v5 != uint16(4) {
		t.Errorf("expected v3 mod 0xFF to equal 4, was %d", v5)
	}
}
