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
	coreV1 "k8s.io/api/core/v1"
)

type serviceTest struct {
	suite.Suite
	chartPath string
	release   string
	namespace string
	templates []string
}

func TestServiceTemplate(t *testing.T) {
	t.Parallel() // marking the test to be run in Parallel

	chartPath, err := filepath.Abs("../../app")
	fmt.Println(chartPath)
	require.NoError(t, err)

	suite.Run(t, &serviceTest{
		chartPath: chartPath,
		release:   "sample-app-test",
		namespace: "app-" + strings.ToLower(random.UniqueId()), //This will create a unique namespace for each TestServiceTemplate Run
		templates: []string{"templates/service.yaml"},
	})
}

func (s *serviceTest) TestService() {

	options := &helm.Options{
		SetValues: map[string]string{
			"service.type": "NodePort",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", s.namespace),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.release, s.templates)
	var service coreV1.Service
	helm.UnmarshalK8SYaml(s.T(), output, &service)

	s.Require().Equal("NodePort", string(service.Spec.Type))
}
