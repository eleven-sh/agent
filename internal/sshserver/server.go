package sshserver

import (
	"log"
	"os/user"

	"github.com/gliderlabs/ssh"
)

func NewServer(
	hostKeyFilePath string,
	authorizedUsers []AuthorizedUser,
	listenAddr string,
) (*ssh.Server, error) {

	auth := newAuthenticator(hostKeyFilePath, authorizedUsers)

	hostKeySigner, err := auth.buildHostKeySigner()

	if err != nil {
		return nil, err
	}

	forwardHandler := &ssh.ForwardedTCPHandler{}

	return &ssh.Server{
		Addr: listenAddr,

		HostSigners: []ssh.Signer{hostKeySigner},

		PublicKeyHandler: func(ctx ssh.Context, publicKey ssh.PublicKey) bool {
			publicKeyAuthorized, err := auth.doesPublicKeyAuthorizedForUser(
				ctx.User(),
				publicKey,
			)

			if err != nil {
				log.Printf(
					"[SSH server] Error in public key handler: %v",
					err,
				)

				return false
			}

			return publicKeyAuthorized
		},

		Handler: func(sshSession ssh.Session) {
			user, err := user.Lookup(sshSession.User())

			if err != nil {
				log.Printf(
					"[SSH server] Error during user lookup: %v",
					err,
				)

				sshSession.Close()
				return
			}

			sessionCmdBuilder := newSessionCmdBuilder(user)

			sessionHandler := newSessionHandler(sessionCmdBuilder)
			sessionHandler.handle(sshSession)
		},

		LocalPortForwardingCallback: ssh.LocalPortForwardingCallback(func(ctx ssh.Context, dhost string, dport uint32) bool {
			return true
		}),

		ReversePortForwardingCallback: ssh.ReversePortForwardingCallback(func(ctx ssh.Context, host string, port uint32) bool {
			return true
		}),

		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":        forwardHandler.HandleSSHRequest,
			"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
		},

		ChannelHandlers: map[string]ssh.ChannelHandler{
			"direct-tcpip":                   ssh.DirectTCPIPHandler,
			"session":                        ssh.DefaultSessionHandler,
			"direct-streamlocal@openssh.com": handleDirectStreamLocalOpenSSH,
		},
	}, nil
}
