package system

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDoesFileExist(t *testing.T) {
	testCases := []struct {
		test             string
		path             string
		expectedResponse bool
	}{
		{
			test:             "with existing file",
			path:             "./testdata/existing_file",
			expectedResponse: true,
		},

		{
			test:             "with non-existing file",
			path:             "./testdata/non_existing_file",
			expectedResponse: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			exists, err := DoesFileExist(tc.path)

			if err != nil {
				t.Fatalf("expected no error, got '%+v'", err)
			}

			if exists != tc.expectedResponse {
				t.Fatalf(
					"expected file exists to equal '%v', got '%v'",
					tc.expectedResponse,
					exists,
				)
			}
		})
	}
}

func TestRemoveDirContent(t *testing.T) {
	testCases := []struct {
		test                   string
		dirPath                string
		firstLvlFilesCtAtStart int
	}{
		{
			test:                   "with empty dir",
			dirPath:                "./testdata/empty_dir",
			firstLvlFilesCtAtStart: 0,
		},

		{
			test:                   "with non-empty dir",
			dirPath:                "./testdata/non_empty_dir",
			firstLvlFilesCtAtStart: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			defer func() {
				err := os.MkdirAll(
					"./testdata/non_empty_dir/test/placeholder_dir",
					os.FileMode(0700),
				)

				if err != nil {
					t.Fatalf("expected no error, got '%+v'", err)
				}

				err = os.MkdirAll(
					"./testdata/non_empty_dir/b",
					os.FileMode(0700),
				)

				if err != nil {
					t.Fatalf("expected no error, got '%+v'", err)
				}

				err = os.WriteFile(
					"./testdata/non_empty_dir/placeholder",
					[]byte(""),
					os.FileMode(0600),
				)

				if err != nil {
					t.Fatalf("expected no error, got '%+v'", err)
				}

				err = os.WriteFile(
					"./testdata/non_empty_dir/b/placeholder",
					[]byte(""),
					os.FileMode(0600),
				)

				if err != nil {
					t.Fatalf("expected no error, got '%+v'", err)
				}
			}()

			fileAndDirPaths, err := filepath.Glob(
				filepath.Join(tc.dirPath, "*"),
			)

			if err != nil {
				t.Fatalf("expected no error, got '%+v'", err)
			}

			if tc.firstLvlFilesCtAtStart != len(fileAndDirPaths) {
				t.Fatalf(
					"expected first level files count at start to equal '%d', got '%d'",
					tc.firstLvlFilesCtAtStart,
					len(fileAndDirPaths),
				)
			}

			err = RemoveDirContent(tc.dirPath)

			if err != nil {
				t.Fatalf("expected no error, got '%+v'", err)
			}

			fileAndDirPaths, err = filepath.Glob(
				filepath.Join(tc.dirPath, "*"),
			)

			if err != nil {
				t.Fatalf("expected no error, got '%+v'", err)
			}

			if fileAndDirPaths != nil {
				t.Fatalf(
					"expected empty dir, got '%+v'",
					fileAndDirPaths,
				)
			}
		})
	}
}
