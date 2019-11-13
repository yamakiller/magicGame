package global

import "sync"

var (
	oneEnvir sync.Once
	env      *GatewayEnv
)

type GatewayEnv struct {
	ServerID                    int32 `json:"server-id"`
	ClientMax                   int   `json:"client-max"`
	ClientOutQueueMax           int   `json:"client-out-queue-max"`
	ClientKleep                 int   `json:"client-kleep-time"`
	ClientBufferLimit           int   `json:"client-buffer-limit"`
	LinkerForwardErr            int   `json:"linker-forward-err"`
	LinkerForwardErrInterval    int   `json:"linker-forward-err-interval"`
	ResponseClientBufferLimit   int
	ResponseClientCheckInterval int    `json:"response-client-check-time"`
	ResponseTimeout             int    `json:"response-timeout"`
	NetAddr                     string `json:"listen-address"`
	NetProtoVersion             int32  `json:"net-protocol-version"`
}

//EnvInstance
//@method EnvInstance desc: Global environment variable singleton mode interface function
//@return (*Envir) environment variable
func EnvInstance() *GatewayEnv {
	oneEnvir.Do(func() {
		env = &GatewayEnv{}
	})

	return env
}
