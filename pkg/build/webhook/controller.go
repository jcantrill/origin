package webhook

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/openshift/origin/pkg/build/api"
	"github.com/openshift/origin/pkg/client"
)

// Webhook verification is dependent on the sending side, it can be
// eg. github, bitbucket or else, so there must be a separate Plugin
// instance for each webhook provider.
type Plugin interface {
	// Method extracts build information returning it with eventual error.
	Extract(buildCfg *api.BuildConfig, path string, req *http.Request) (*api.Build, error)
}

// controller used for processing webhook requests.
type controller struct {
	osClient client.Interface
	plugins  map[string]Plugin
}

// urlVars holds parsed URL parts.
type urlVars struct {
	buildId string
	secret  string
	plugin  string
	path    string
}

// NewController creates new webhook controller and feed it with provided plugins.
func NewController(osClient client.Interface, plugins map[string]Plugin) http.Handler {
	return &controller{osClient: osClient, plugins: plugins}
}

// ServeHTTP main REST service method.
func (c *controller) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	uv, err := parseUrl(req.URL.Path)
	if err != nil {
		notFound(w, err.Error())
		return
	}

	buildCfg, err := c.osClient.GetBuildConfig(uv.buildId)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	if uv.secret != buildCfg.Secret {
		badRequest(w, "")
		return
	}

	plugin, ok := c.plugins[uv.plugin]
	if !ok {
		notFound(w, "Plugin ", uv.plugin, " not found!")
		return
	}
	build, err := plugin.Extract(buildCfg, uv.path, req)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	if build == nil {
		build = &api.Build{
			Input: buildCfg.DesiredInput,
		}
	}

	if _, err := c.osClient.CreateBuild(build); err != nil {
		badRequest(w, err.Error())
	}
}

func parseUrl(url string) (uv urlVars, err error) {
	parts := splitPath(url)
	if len(parts) < 3 {
		err = fmt.Errorf("Unexpected URL %s!", url)
		return
	}
	uv = urlVars{parts[0], parts[1], parts[2], ""}
	if len(parts) > 3 {
		uv.path = strings.Join(parts[3:], "/")
	}
	return
}

func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

func notFound(w http.ResponseWriter, args ...string) {
	http.Error(w, strings.Join(args, ""), http.StatusNotFound)
}

func badRequest(w http.ResponseWriter, args ...string) {
	http.Error(w, strings.Join(args, ""), http.StatusBadRequest)
}
