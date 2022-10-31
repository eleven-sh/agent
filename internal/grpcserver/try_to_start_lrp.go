package grpcserver

import (
	"sync"
	"time"

	"github.com/eleven-sh/agent/internal/state"
	"github.com/eleven-sh/agent/proto"
)

func (*agentServer) TryToStartLongRunningProcess(
	req *proto.TryToStartLongRunningProcessRequest,
	stream proto.Agent_TryToStartLongRunningProcessServer,
) error {

	heartbeatChan := make(chan error, 1)
	var heartbeatChanLock sync.Mutex

	go func() {
		defer heartbeatChanLock.Unlock()

		pollSleepDuration := 1 * time.Second

		for {
			heartbeatChanLock.Lock()

			select {
			case <-heartbeatChan:
				return
			default:
				err := stream.Send(&proto.TryToStartLongRunningProcessReply{
					Heartbeat: "beat",
				})

				if err != nil {
					heartbeatChan <- err
					return
				}
			}

			heartbeatChanLock.Unlock()

			time.Sleep(pollSleepDuration)
		}
	}()

	exitOutput, exitErrMsg, err := state.StartProcessAndWaitForSleep(
		req.Cwd,
		req.Cmd,
		heartbeatChan,
	)

	heartbeatChanLock.Lock()
	close(heartbeatChan)
	heartbeatChanLock.Unlock()

	if err != nil {
		return err
	}

	return stream.Send(&proto.TryToStartLongRunningProcessReply{
		ErrorOutput:  exitOutput,
		ErrorMessage: exitErrMsg,
	})
}
