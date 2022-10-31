package sshserver

import (
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestDoesPublicKeyAuthorizedForUser(t *testing.T) {
	testCases := []struct {
		test             string
		authorizedUsers  []AuthorizedUser
		username         string
		publicKey        string
		expectedResponse bool
	}{
		{
			test: "with ed25519 public key authorized for user",
			authorizedUsers: []AuthorizedUser{
				{
					UserName:               "jeremy",
					AuthorizedKeysFilePath: "./testdata/authorized_keys",
				},

				{
					UserName:               "root",
					AuthorizedKeysFilePath: "./testdata/empty_authorized_keys",
				},
			},
			username:         "jeremy",
			publicKey:        "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJqmVkvKmywIYkfXOWWPya3I1zAbWGwOGu9Q870Zh49v jeremylevy@macbook-pro-de-jeremy.home",
			expectedResponse: true,
		},

		{
			test: "with rsa public key authorized for user",
			authorizedUsers: []AuthorizedUser{
				{
					UserName:               "root",
					AuthorizedKeysFilePath: "./testdata/authorized_keys",
				},
			},
			username:         "root",
			publicKey:        "ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAQB/nAmOjTmezNUDKYvEeIRf2YnwM9/uUG1d0BYsc8/tRtx+RGi7N2lUbp728MXGwdnL9od4cItzky/zVdLZE2cycOa18xBK9cOWmcKS0A8FYBxEQWJ/q9YVUgZbFKfYGaGQxsER+A0w/fX8ALuk78ktP31K69LcQgxIsl7rNzxsoOQKJ/CIxOGMMxczYTiEoLvQhapFQMs3FL96didKr/QbrfB1WT6s3838SEaXfgZvLef1YB2xmfhbT9OXFE3FXvh2UPBfN+ffE7iiayQf/2XR+8j4N4bW30DiPtOQLGUrH1y5X/rpNZNlWW2+jGIxqZtgWg7lTy3mXy5x836Sj/6L jje.levy@gmail.com",
			expectedResponse: true,
		},

		{
			test: "with unauthorized user",
			authorizedUsers: []AuthorizedUser{
				{
					UserName:               "jeremy",
					AuthorizedKeysFilePath: "./testdata/authorized_keys",
				},
			},
			username:         "root",
			publicKey:        "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJqmVkvKmywIYkfXOWWPya3I1zAbWGwOGu9Q870Zh49v jeremylevy@macbook-pro-de-jeremy.home",
			expectedResponse: false,
		},

		{
			test: "with public key not authorized for user",
			authorizedUsers: []AuthorizedUser{
				{
					UserName:               "jeremy",
					AuthorizedKeysFilePath: "./testdata/authorized_keys",
				},

				{
					UserName:               "root",
					AuthorizedKeysFilePath: "./testdata/empty_authorized_keys",
				},
			},
			username:         "jeremy",
			publicKey:        "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFDoDZZVSGDdQGaSARMJ/4fTs2CvdUl1tPN47Xkz8YzY jeremylevy@macbook-pro-de-jeremy.home",
			expectedResponse: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			auth := newAuthenticator("", tc.authorizedUsers)

			publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(tc.publicKey))

			if err != nil {
				t.Fatalf("expected no error, got '%+v'", err)
			}

			publicKeyAuthorized, err := auth.doesPublicKeyAuthorizedForUser(
				tc.username,
				publicKey,
			)

			if err != nil {
				t.Fatalf("expected no error, got '%+v'", err)
			}

			if publicKeyAuthorized != tc.expectedResponse {
				t.Fatalf(
					"expected public key authorized to equal '%v', got '%v'",
					tc.expectedResponse,
					publicKeyAuthorized,
				)
			}
		})
	}
}
