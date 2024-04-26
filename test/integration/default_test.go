package test

import (
	"fmt"
	"time"
	"strings"
	"testing"
  "crypto/tls"
  "path/filepath"

	"github.com/gruntwork-io/terratest/modules/helm"
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type integrationTest struct {
	suite.Suite
	chartPath string
	release   string
	namespace string
	templates []string
}


// func TestIsValidChart(t *testing.T) {
func TestDeploymentTemplate(t *testing.T) {
	t.Parallel() // marking the test to be run in Parallel

	chartPath, err := filepath.Abs("../../app")
	require.NoError(t, err)
 
  // tmplDir, err := os.ReadDir(filepath.Join(chartPath, "templates/"))
  // if err != nil {
  //   require.NoError(t, err)
  // }
  //
  // tmplFiles := make([]string, len(tmplDir))
  // for _, file := range tmplDir {
  //   if strings.Contains(file.Name(), ".yaml") {
  //     tmplFiles = append(tmplFiles, filepath.Join("templates/", file.Name()))
  //   }
  // }
  // 
	suite.Run(t, &integrationTest{
		chartPath: chartPath,
		release:   "sample-app-test",
		namespace: "app-" + strings.ToLower(random.UniqueId()), //This will create a unique namespace for each TestServiceTemplate Run
		templates: []string{"templates"},
	})
}


func (i *integrationTest) TestIntegrationDefault() {
	// t.Parallel() // marking the test to be run in Parallel

	//chartPath, err := filepath.Abs("../../app")
	//require.NoError(t, err)
  

	// randomId := strings.ToLower(random.UniqueId())
	// namespace := "app-" + randomId
	// release := "app-test-" + randomId

	k8sOptions := k8s.NewKubectlOptions("", "", i.namespace)
	k8s.CreateNamespace(i.T(), k8sOptions, i.namespace)

	defer k8s.DeleteNamespace(i.T(), k8sOptions, i.namespace)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.repository": "nginx",
			"image.tag":        "1.25.4-alpine-slim",
		},
		KubectlOptions: k8sOptions,
	}
   
	helm.Install(i.T(), options, i.chartPath, i.release)
	//defer helm.Delete(t, options, release, true)

	// Wait for the service adn deployment to be ready
	k8s.WaitUntilServiceAvailable(i.T(), k8sOptions, i.release, 10, 1*time.Second)
	k8s.WaitUntilDeploymentAvailable(i.T(), k8sOptions, i.release, 10, 1*time.Second)

	// Create a tunnel to the service, so we can make requests to it
	tunnel := k8s.NewTunnel(k8sOptions, k8s.ResourceTypeService, i.release, 8080, 80)
	defer tunnel.Close()
	tunnel.ForwardPort(i.T())
	fmt.Println(tunnel.Endpoint())

	tlsConfig := tls.Config{}

	http_helper.HttpGetWithRetryWithCustomValidation(
		i.T(),
		fmt.Sprintf("http://%s", tunnel.Endpoint()),
		&tlsConfig,
		5,
		2*time.Second,
		func(status int, body string) bool {
			return status == 200
		})
}


