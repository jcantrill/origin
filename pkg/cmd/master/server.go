package master

import(
	kubeclient 	"github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	osclient 	"github.com/openshift/origin/pkg/client"
)

type ServerConfig struct {
  config
}

type MasterServer struct{
	sig chan int
	ServerConfig
}

func NewServerConfig(uri string, bindAddr string, nodeHosts []string) *ServerConfig{
	//add checks for empty values?
	return &ServerConfig{
		config: config{
			ListenAddr:  uri,
			bindAddr: 	bindAddr,
			nodeHosts: 	nodeHosts,
		},
	}
}

func (s *MasterServer) OsClient() *osclient.Client {
	return s.getOsClient()
}

func (s *MasterServer) KubeClient() *kubeclient.Client {
	return s.getKubeClient()
}

func (cfg *ServerConfig) StartAllInOne() *MasterServer {
	done := make(chan int, 1)
	go func(){
		cfg.startAllInOne(done)
	}()
	return &MasterServer{
		sig: done,
		ServerConfig: *cfg,
	}
}

func (s *MasterServer) Stop(){
	s.sig <- 1
}
