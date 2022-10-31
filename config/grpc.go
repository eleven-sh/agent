package config

const (
	GRPCServerAddrProtocol = "unix"
	GRPCServerAddr         = ElevenAgentConfigDirPath + "/grpc-server.sock"
	GRPCServerURI          = "unix://" + GRPCServerAddr
)
