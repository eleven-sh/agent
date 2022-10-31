package config

const (
	DefaultSSHServerListenPort = "22"
	SSHServerListenPort        = "2200"
	SSHServerListenAddr        = ":" + SSHServerListenPort
	SSHServerHostKeyFilePath   = ElevenUserHomeDirPath + "/.ssh/eleven-ssh-server-host-key"

	ElevenUserAuthorizedSSHKeysFilePath = ElevenUserHomeDirPath + "/.ssh/authorized_keys"
	GitHubPublicSSHKeyFilePath          = ElevenUserHomeDirPath + "/.ssh/" + ElevenUserName + "-github.pub"
)
