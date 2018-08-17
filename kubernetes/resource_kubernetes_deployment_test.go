package kubernetes

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	v1beta "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func TestAccKubernetesDeployment_basic(t *testing.T) {
	var conf1 v1beta.Deployment
	var conf2 v1beta.Deployment

	deploymentName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	imageName1 := "nginx:1.7.9"
	imageName2 := "nginx:1.11"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesDeploymentConfigBasic(deploymentName, imageName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesDeploymentExists("kubernetes_deployment.test", &conf1),
					resource.TestCheckResourceAttr("kubernetes_deployment.test", "metadata.0.annotations.%", "0"),
					resource.TestCheckResourceAttr("kubernetes_deployment.test", "metadata.0.labels.%", "1"),
					resource.TestCheckResourceAttr("kubernetes_deployment.test", "metadata.0.labels.app", "deployment_label"),
					resource.TestCheckResourceAttr("kubernetes_deployment.test", "metadata.0.name", deploymentName),
					resource.TestCheckResourceAttrSet("kubernetes_deployment.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("kubernetes_deployment.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("kubernetes_deployment.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("kubernetes_deployment.test", "metadata.0.uid"),
					resource.TestCheckResourceAttr("kubernetes_deployment.test", "spec.0.template.0.spec.0.container.0.image", imageName1),
				),
			},
			{
				Config: testAccKubernetesDeploymentConfigBasic(deploymentName, imageName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesDeploymentExists("kubernetes_deployment.test", &conf2),
					resource.TestCheckResourceAttr("kubernetes_deployment.test", "spec.0.template.0.spec.0.container.0.image", imageName2),
					testAccCheckKubernetesDeploymentForceNew(&conf1, &conf2, false),
				),
			},
		},
	})
}

func testAccCheckKubernetesDeploymentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*kubernetes.Clientset)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "kubernetes_deployment" {
			continue
		}

		namespace, name, err := idParts(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := conn.ExtensionsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{})
		if err == nil {
			if resp.Namespace == namespace && resp.Name == name {
				return fmt.Errorf("Deployment still exists: %s: %#v", rs.Primary.ID, resp.Status.Conditions)
			}
		}
	}

	return nil
}

func testAccCheckKubernetesDeploymentExists(n string, obj *v1beta.Deployment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*kubernetes.Clientset)

		namespace, name, err := idParts(rs.Primary.ID)
		if err != nil {
			return err
		}

		out, err := conn.ExtensionsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		*obj = *out
		return nil
	}
}

func testAccCheckKubernetesDeploymentForceNew(old, new *v1beta.Deployment, wantNew bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if wantNew {
			if old.ObjectMeta.UID == new.ObjectMeta.UID {
				return fmt.Errorf("Expecting new resource for pod %s", old.ObjectMeta.UID)
			}
		} else {
			if old.ObjectMeta.UID != new.ObjectMeta.UID {
				return fmt.Errorf("Expecting pod UIDs to be the same: expected %s got %s", old.ObjectMeta.UID, new.ObjectMeta.UID)
			}
		}
		return nil
	}
}

func testAccKubernetesDeploymentConfigBasic(deploymentName, imageName string) string {
	return fmt.Sprintf(`
resource "kubernetes_deployment" "test" {
  metadata {
    name = "%s"

    labels {
      app = "deployment_label"
    }
  }

  spec{
    template {
      spec {
        container {
          image = "%s"
          name  = "containername"
        }
      }
    }
  }
}`, deploymentName, imageName)
}
