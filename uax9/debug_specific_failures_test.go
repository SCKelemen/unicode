package uax9

import (
	"testing"
)

func TestDebugSpecificFailures(t *testing.T) {
	cases := []struct {
		name     string
		classes  []BidiClass
		para     int
		expected []int
	}{
		{
			name:     "EN PDF ET para=1",
			classes:  []BidiClass{ClassEN, ClassPDF, ClassET},
			para:     1,
			expected: []int{2, -1, 2},
		},
		{
			name:     "EN BN ET para=1",
			classes:  []BidiClass{ClassEN, ClassBN, ClassET},
			para:     1,
			expected: []int{2, -1, 2},
		},
		{
			name:     "ET PDF EN para=1",
			classes:  []BidiClass{ClassET, ClassPDF, ClassEN},
			para:     1,
			expected: []int{2, -1, 2},
		},
		{
			name:     "FSI EN R para=0",
			classes:  []BidiClass{ClassFSI, ClassEN, ClassR},
			para:     0,
			expected: []int{0, 2, 1},
		},
		{
			name:     "AN ES AN para=1",
			classes:  []BidiClass{ClassAN, ClassES, ClassAN},
			para:     1,
			expected: []int{2, 1, 2},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			classesCopy := make([]BidiClass, len(tc.classes))
			copy(classesCopy, tc.classes)

			actual := computeLevels(classesCopy, tc.para)

			t.Logf("Input:    %v", classesToString(tc.classes))
			t.Logf("Expected: %v", tc.expected)
			t.Logf("Actual:   %v", actual)

			match := len(actual) == len(tc.expected)
			if match {
				for i := range tc.expected {
					if tc.expected[i] != -1 && tc.expected[i] != actual[i] {
						match = false
						break
					}
				}
			}

			if !match {
				t.Errorf("Levels mismatch")
			}
		})
	}
}
