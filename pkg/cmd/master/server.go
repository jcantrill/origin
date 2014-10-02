package master

import(
	"net"
	"time"
	kubeclient 	"github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	osclient 	"github.com/openshift/origin/pkg/client"
)

type ServerConfig struct {
  config
  Timeout time.Duration
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

func (cfg *ServerConfig) StartAllInOne() (*MasterServer, error) {
	done := make(chan int, 1)
	go func(){
		cfg.startAllInOne(done)
	}()
	
	
	unblock := make(chan int, 1)
	var err error
	
	go func(){
		start := time.Now()
		for elapsed := 0 * time.Nanosecond; cfg.Timeout > elapsed; elapsed = time.Now().Sub(start){
			dialer := net.Dialer{
				Timeout: cfg.Timeout, 
				DualStack: true,
			}
			var conn net.Conn
			conn, err = dialer.Dial("tcp", cfg.ListenAddr)
			if conn != nil{
				unblock <- 1
	  			defer conn.Close()
				break
			}
		}
	}()
	//block until server is up
	<-unblock
	if err != nil{
		return nil, err
	}

	return &MasterServer{
		sig: done,
		ServerConfig: *cfg,
	}, nil
}

func (s *MasterServer) Stop(){
	s.sig <- 1
}
