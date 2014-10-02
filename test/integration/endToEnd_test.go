package integration

import (
	"testing"
	"time"
//	"encoding/json"
//	kubeclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	//the required functionality moved to an 'api' package?
//	kubeopts "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
//	osclient "github.com/openshift/origin/pkg/client"
	buildapi "github.com/openshift/origin/pkg/build/api"
			 "github.com/openshift/origin/pkg/cmd/master"
)

func TestEndToEnd(t *testing.T){
	//create server
	serverConfig := master.NewServerConfig("127.0.0.1:8080", "127.0.0.1", []string{"127.0.0.1"})
	serverConfig.Timeout = 10 * time.Second
	server, err := serverConfig.StartAllInOne()
	if err != nil{
		t.Error("Unable to start the server in the alloted time: %s", err)	
	}
	defer server.Stop()

	//create openshift client
	osclient := server.OsClient()

	//create kube client
	//kubeclient :=server.KubeClient()
	
	/*
	//create config via openshift command
	opts := kubeopts.KubeConfig{
		Config: "buildcfg/buildcfg.json"
	}
	if err := json.Unmarshal(
*/
	sourceConfig := &buildapi.BuildConfig{}
	_, err = osclient.CreateBuildConfig(sourceConfig)
	//validate output matches expectation
	if err != nil{
		t.Error("Create of buildConfig failed with error: %s", err)
	}


  //trigger build
  //validate response

  //monitor response (with timeout) for build complete

}

