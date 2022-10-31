package sshserver

import (
	"fmt"
	"io"
	"net"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

// directStreamLocalOpenSSHMsg represents the message
// sent during the opening of "direct-streamlocal@openssh.com" channels.
//
// See openssh-portable/PROTOCOL, section 2.4.
// connection: Unix domain socket forwarding
// https://github.com/openssh/openssh-portable/blob/master/PROTOCOL#L235
type directStreamLocalOpenSSHMsg struct {
	SocketPath string
	Reserved0  string
	Reserved1  uint32
}

// handleDirectStreamLocalOpenSSH is used to
// forward local conn to a remote unix socket.
// Corresponds to the "direct-streamlocal@openssh.com" channel type.
// Used by the Eleven CLI to reach the GRPC server unix socket.
func handleDirectStreamLocalOpenSSH(
	srv *ssh.Server,
	conn *gossh.ServerConn,
	newChan gossh.NewChannel,
	ctx ssh.Context,
) {

	msg := directStreamLocalOpenSSHMsg{}
	err := gossh.Unmarshal(newChan.ExtraData(), &msg)

	if err != nil {
		newChan.Reject(
			gossh.ConnectionFailed,
			fmt.Sprintf(
				"error parsing direct stream local OpenSSH data: %s",
				err.Error(),
			),
		)
		return
	}

	socketConn, err := net.Dial("unix", msg.SocketPath)

	if err != nil {
		newChan.Reject(gossh.ConnectionFailed, err.Error())
		return
	}

	requestChan, reqs, err := newChan.Accept()

	if err != nil {
		socketConn.Close()
		return
	}

	go gossh.DiscardRequests(reqs)

	go func() {
		defer requestChan.Close()
		defer socketConn.Close()
		io.Copy(requestChan, socketConn)
	}()

	go func() {
		defer requestChan.Close()
		defer socketConn.Close()
		io.Copy(socketConn, requestChan)
	}()
}
