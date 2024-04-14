package test

import (
	"crypto/tls"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/stretchr/testify/require"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
)

func TestIntegrationDefault(t *testing.T) {
	t.Parallel() // marking the test to be run in Parallel

	chartPath, err := filepath.Abs("../../app")
	require.NoError(t, err)

	randomId := strings.ToLower(random.UniqueId())
	namespace := "app-" + randomId
	release := "app-test-" + randomId

	k8sOptions := k8s.NewKubectlOptions("", "", namespace)
	k8s.CreateNamespace(t, k8sOptions, namespace)

	defer k8s.DeleteNamespace(t, k8sOptions, namespace)

	options := &helm.Options{
		KubectlOptions: k8sOptions,
		SetValues: map[string]string{
			"image.repository": "nginx",
			"image.tag":        "1.25.4-alpine-slim",
		},
	}

	helm.Install(t, options, chartPath, release)
	//defer helm.Delete(t, options, release, true)

	// Wait for the service adn deployment to be ready
	k8s.WaitUntilServiceAvailable(t, k8sOptions, release, 10, 1*time.Second)
	k8s.WaitUntilDeploymentAvailable(t, k8sOptions, release, 10, 1*time.Second)

	// Create a tunnel to the service, so we can make requests to it
	tunnel := k8s.NewTunnel(k8sOptions, k8s.ResourceTypeService, release, 8080, 80)
	defer tunnel.Close()
	tunnel.ForwardPort(t)
	fmt.Println(tunnel.Endpoint())

	tlsConfig := tls.Config{}

	http_helper.HttpGetWithRetryWithCustomValidation(
		t,
		fmt.Sprintf("http://%s", tunnel.Endpoint()),
		&tlsConfig,
		5,
		2*time.Second,
		func(status int, body string) bool {
			return status == 200
		})
}
