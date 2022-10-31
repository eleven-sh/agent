package sshserver

import (
	"os"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

type AuthorizedUser struct {
	UserName               string
	AuthorizedKeysFilePath string
}

type authenticator struct {
	hostKeyFilePath string
	authorizedUsers []AuthorizedUser
}

func newAuthenticator(
	hostKeyFilePath string,
	authorizedUsers []AuthorizedUser,
) *authenticator {

	return &authenticator{
		hostKeyFilePath: hostKeyFilePath,
		authorizedUsers: authorizedUsers,
	}
}

func (a *authenticator) buildHostKeySigner() (ssh.Signer, error) {
	hostKey, err := os.ReadFile(a.hostKeyFilePath)

	if err != nil {
		return nil, err
	}

	return gossh.ParsePrivateKey(hostKey)
}

func (a *authenticator) doesPublicKeyAuthorizedForUser(
	username string,
	publicKey ssh.PublicKey,
) (bool, error) {

	authorizedKeys, err := a.lookupAuthorizedKeysForUser(
		username,
	)

	if err != nil {
		return false, err
	}

	if authorizedKeys == nil {
		return false, nil
	}

	for _, authorizedKey := range authorizedKeys {
		if ssh.KeysEqual(publicKey, authorizedKey) {
			return true, nil
		}
	}

	return false, nil
}

func (a *authenticator) lookupAuthorizedKeysForUser(
	username string,
) ([]ssh.PublicKey, error) {

	for _, authorizedUser := range a.authorizedUsers {
		if authorizedUser.UserName != username {
			continue
		}

		authorizedKeysBytes, err := os.ReadFile(
			authorizedUser.AuthorizedKeysFilePath,
		)

		if err != nil {
			return nil, err
		}

		authorizedKeys, err := parseAuthorizedKeys(
			authorizedKeysBytes,
		)

		if err != nil {
			return nil, err
		}

		return authorizedKeys, nil
	}

	return nil, nil
}

func parseAuthorizedKeys(
	authorizedKeysBytes []byte,
) ([]ssh.PublicKey, error) {

	authorizedKeys := []ssh.PublicKey{}

	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)

		if err != nil {
			return nil, err
		}

		authorizedKeys = append(authorizedKeys, pubKey)
		authorizedKeysBytes = rest
	}

	return authorizedKeys, nil
}
