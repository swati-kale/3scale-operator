package e2e

import (
	goctx "context"
	"testing"
	"time"

	"github.com/3scale/3scale-operator/pkg/apis"
	appsgroup "github.com/3scale/3scale-operator/pkg/apis/apps"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	e2eutil "github.com/3scale/3scale-operator/test/e2e/e2eutil"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	frameworke2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	clientappsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
)

func TestApiManagerController(t *testing.T) {
	var err error

	apimanagerList := &appsv1alpha1.APIManagerList{
		TypeMeta: metav1.TypeMeta{
			Kind:       appsgroup.APIManagerKind,
			APIVersion: appsv1alpha1.SchemeGroupVersion.String(),
		},
	}

	err = framework.AddToFrameworkScheme(apis.AddToScheme, apimanagerList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	// Run subtests
	t.Run("apimanager-group", func(t *testing.T) {
		t.Run("StandardDeploy", productizedUnconstrainedDeploymentSubtest)
	})
}

func newAPIManagerCluster(t *testing.T) (*framework.Framework, *framework.TestCtx) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("initialized cluster resources")

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	f := framework.Global
	t.Log("waiting until operator Deployment is ready...")
	err = frameworke2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "3scale-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("operator Deployment is ready")

	return f, ctx
}

func productizedUnconstrainedDeploymentSubtest(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	t.Log("Example text")
	time.Sleep(5 * time.Second)
	t.Fatal("Finished test")

	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("initialized cluster resources")

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	f := framework.Global
	t.Log("waiting until operator Deployment is ready...")

	err = frameworke2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "3scale-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("operator Deployment is ready")

	apimanager := &appsv1alpha1.APIManager{
		Spec: appsv1alpha1.APIManagerSpec{
			AmpRelease:     "2.4",
			WildcardDomain: "test1.127.0.0.1.nip.io",
			Productized:    false,
			Evaluation:     true,
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "example-apimanager",
			Namespace: namespace,
		},
	}

	err = f.Client.Create(goctx.TODO(), apimanager, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	osAppsV1Client, err := clientappsv1.NewForConfig(f.KubeConfig)
	if err != nil {
		t.Fatal(err)
	}

	err = waitForAllApiManagerStandardDeploymentConfigs(t, f.KubeClient, osAppsV1Client, namespace, "3scale-operator", retryInterval, time.Minute*15)
	if err != nil {
		t.Fatal(err)
	}
}

func waitForAllApiManagerStandardDeploymentConfigs(t *testing.T, kubeclient kubernetes.Interface, osAppsV1Client clientappsv1.AppsV1Interface, namespace, name string, retryInterval, timeout time.Duration) error {
	deploymentConfigNames := []string{ // TODO gather this from constants/somewhere centralized
		"apicast-production",
		"apicast-staging",
		"apicast-wildcard-router",
		"backend-cron",
		"backend-listener",
		"backend-redis",
		"backend-worker",
		"system-app",
		"system-memcache",
		"system-mysql",
		"system-redis",
		"system-sidekiq",
		"system-sphinx",
		"zync",
		"zync-database",
	}

	for _, dcName := range deploymentConfigNames {
		err := e2eutil.WaitForDeploymentConfig(t, kubeclient, osAppsV1Client, namespace, dcName, retryInterval, time.Minute*15)
		if err != nil {
			return err
		}
	}

	return nil
}