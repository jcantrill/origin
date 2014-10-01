package integration

import (
	"testing"
//	"encoding/json"
//	kubeclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	//the required functionality moved to an 'api' package?
//	kubeopts "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
//	osclient "github.com/openshift/origin/pkg/client"
//	buildapi "github.com/openshift/origin/pkg/build/api"
			 "github.com/openshift/origin/pkg/cmd/master"
)

func TestEndToEnd(t *testing.T){
	//create server
	serverConfig := master.NewServerConfig("127.0.0.1:8080", "127.0.0.1", []string{"127.0.0.1"})
	server := serverConfig.StartAllInOne()
	defer server.Stop()

	//create openshift client
	osclient := server.OsClient()

	//create kube client
	kubeclient :=server.KubeClient()
	
	/*
	//create config via openshift command
	opts := kubeopts.KubeConfig{
		Config: "buildcfg/buildcfg.json"
	}
	sourceConfig := new(buildapi.BuildConfig)
	if err := json.Unmarshal(
	buildConfig, err := client.CreateBuildConfig()
	//validate output matches expectation
	if err != nil{
		t.Fail("Create of buildConfig failed with error:", err)
	}
*/


  //trigger build
  //validate response

  //monitor response (with timeout) for build complete

}

