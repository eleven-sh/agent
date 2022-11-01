package grpcserver

import (
	"reflect"
	"testing"
)

func TestSortRuntimes(t *testing.T) {
	testCases := []struct {
		test             string
		runtimes         map[string]string
		expectedRuntimes []runtime
	}{
		{
			test:             "with no runtimes",
			runtimes:         map[string]string{},
			expectedRuntimes: []runtime{},
		},

		{
			test: "with ruby runtime",
			runtimes: map[string]string{
				"docker": "latest",
				"ruby":   "3.1.2",
				"python": "2.4.0",
				"rust":   "1.0.1",
			},
			expectedRuntimes: []runtime{
				{
					name:    "docker",
					version: "latest",
				},

				{
					name:    "python",
					version: "2.4.0",
				},

				{
					name:    "rust",
					version: "1.0.1",
				},

				{
					name:    "ruby",
					version: "3.1.2",
				},
			},
		},

		{
			test: "without ruby runtime",
			runtimes: map[string]string{
				"docker": "latest",
				"python": "2.4.0",
				"node":   "1.0.1",
			},
			expectedRuntimes: []runtime{
				{
					name:    "docker",
					version: "latest",
				},

				{
					name:    "node",
					version: "1.0.1",
				},

				{
					name:    "python",
					version: "2.4.0",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			sortedRuntimes := sortRuntimes(tc.runtimes)

			if !reflect.DeepEqual(sortedRuntimes, tc.expectedRuntimes) {
				t.Fatalf(
					"expected runtimes to equal '%+v', got '%+v'",
					tc.expectedRuntimes,
					sortedRuntimes,
				)
			}
		})
	}
}
