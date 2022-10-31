package env

import (
	"testing"

	"github.com/eleven-sh/agent/proto"
)

func TestWorkspaceCountRepoOwners(t *testing.T) {
	testCases := []struct {
		test          string
		repositories  []*proto.EnvRepository
		expectedCount int
	}{
		{
			test:          "with empty repositories",
			repositories:  []*proto.EnvRepository{},
			expectedCount: 0,
		},

		{
			test: "with one repository owner",
			repositories: []*proto.EnvRepository{
				{
					Name:  "api",
					Owner: "jeremylevy",
				},

				{
					Name:  "website",
					Owner: "jeremylevy",
				},
			},
			expectedCount: 1,
		},

		{
			test: "with multiple repository owners",
			repositories: []*proto.EnvRepository{
				{
					Name:  "api",
					Owner: "jeremylevy",
				},

				{
					Name:  "website",
					Owner: "jeremylevy",
				},

				{
					Name:  "test",
					Owner: "recode-sh",
				},

				{
					Name:  "api",
					Owner: "scaffold-sh",
				},
			},
			expectedCount: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			count := countRepoOwners(tc.repositories)

			if tc.expectedCount != count {
				t.Fatalf(
					"expected count to equal '%d', got '%d'",
					tc.expectedCount,
					count,
				)
			}
		})
	}
}

func TestGetRepoDirPathInWorkspace(t *testing.T) {
	testCases := []struct {
		test            string
		repository      *proto.EnvRepository
		repoOwnersCount int
		expectedPath    string
	}{
		{
			test: "without multiple repository owners",
			repository: &proto.EnvRepository{
				Name:  "test",
				Owner: "jeremylevy",
			},
			repoOwnersCount: 1,
			expectedPath:    "/home/eleven/workspace/test",
		},

		{
			test: "with multiple repository owners",
			repository: &proto.EnvRepository{
				Name:  "api",
				Owner: "jeremylevy",
			},
			repoOwnersCount: 3,
			expectedPath:    "/home/eleven/workspace/jeremylevy-api",
		},

		{
			test: "with bad characters in repository names and one repository owner",
			repository: &proto.EnvRepository{
				Name:  "ap_i(",
				Owner: "jere(m_ylevy",
			},
			repoOwnersCount: 1,
			expectedPath:    "/home/eleven/workspace/ap-i",
		},

		{
			test: "with bad characters in repository names and multiple repository owners",
			repository: &proto.EnvRepository{
				Name:  "ap_i(",
				Owner: "jere(m_ylevy",
			},
			repoOwnersCount: 3,
			expectedPath:    "/home/eleven/workspace/jere-m-ylevy-ap-i",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			path := getRepoDirPathInWorkspace(
				tc.repository,
				tc.repoOwnersCount,
			)

			if tc.expectedPath != path {
				t.Fatalf(
					"expected path to equal '%s', got '%s'",
					tc.expectedPath,
					path,
				)
			}
		})
	}
}
