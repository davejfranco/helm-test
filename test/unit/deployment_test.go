package test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	appsV1 "k8s.io/api/apps/v1"
)

type deploymentTest struct {
	suite.Suite
	chartPath string
	release   string
	namespace string
	templates []string
}

func TestDeploymentTemplate(t *testing.T) {
	t.Parallel() // marking the test to be run in Parallel

	chartPath, err := filepath.Abs("../../app")
	fmt.Println(chartPath)
	require.NoError(t, err)

	suite.Run(t, &deploymentTest{
		chartPath: chartPath,
		release:   "sample-app-test",
		namespace: "app-" + strings.ToLower(random.UniqueId()), //This will create a unique namespace for each TestServiceTemplate Run
		templates: []string{"templates/deployment.yaml"},
	})
}

func (s *deploymentTest) TestDeployment() {

	options := &helm.Options{
		SetValues: map[string]string{
			"replicaCount": "1",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", s.namespace),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.release, s.templates)
	var deployment appsV1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	s.Require().Equal(int32(1), *deployment.Spec.Replicas)
}
