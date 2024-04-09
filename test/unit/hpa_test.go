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
	hpav2 "k8s.io/api/autoscaling/v2"
)

type hpaTest struct {
	suite.Suite
	chartPath string
	release   string
	namespace string
	templates []string
}

func TestHpaTemplate(t *testing.T) {
	t.Parallel() // marking the test to be run in Parallel

	chartPath, err := filepath.Abs("../../app")
	fmt.Println(chartPath)
	require.NoError(t, err)

	suite.Run(t, &hpaTest{
		chartPath: chartPath,
		release:   "sample-app-test",
		namespace: "app-" + strings.ToLower(random.UniqueId()), //This will create a unique namespace for each TestServiceTemplate Run
		templates: []string{"templates/hpa.yaml"},
	})
}

func (s *hpaTest) TestDefaultHpa() {

	options := &helm.Options{
		SetValues: map[string]string{
			"autoscaling.enabled": "true",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", s.namespace),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.release, s.templates)
	var hpa hpav2.HorizontalPodAutoscaler
	helm.UnmarshalK8SYaml(s.T(), output, &hpa)

	//expectedMinReplicas := int32(1)
	//s.Require().Equal(expectedMinReplicas, *hpa.Spec.MinReplicas)
	var expectedValues = map[string]int32{
		"minReplicas":                    int32(1),
		"maxReplicas":                    int32(100),
		"targetCPUUtilizationPercentage": int32(80),
	}

	s.Require().Equal(expectedValues["minReplicas"], *hpa.Spec.MinReplicas)
	s.Require().Equal(expectedValues["maxReplicas"], hpa.Spec.MaxReplicas)
	s.Require().Equal(expectedValues["targetCPUUtilizationPercentage"], *hpa.Spec.Metrics[0].Resource.Target.AverageUtilization)

}

func (s *hpaTest) TestMemoryUtilizationHpa() {

	options := &helm.Options{
		SetValues: map[string]string{
			"autoscaling.enabled":                           "true",
			"autoscaling.targetMemoryUtilizationPercentage": "80",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", s.namespace),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.release, s.templates)
	var hpa hpav2.HorizontalPodAutoscaler
	helm.UnmarshalK8SYaml(s.T(), output, &hpa)

	var expectedMemory int32 = 80
	s.Require().Equal(expectedMemory, *hpa.Spec.Metrics[0].Resource.Target.AverageUtilization)

}

// Test to check if the HPA is disabled
func (s *hpaTest) TestHpaDisabled() {
	options := &helm.Options{
		SetValues: map[string]string{
			"autoscaling.enabled": "false",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", s.namespace),
	}

	_, err := helm.RenderRemoteTemplateE(s.T(), options, s.chartPath, s.release, s.templates)
	// We expect an error because the HPA is disabled
	s.Require().Error(err)
}
