package main

import (
	"context"
	"github.com/ghodss/yaml"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/lithammer/dedent"
	rbac "k8s.io/api/rbac/v1"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"text/template"
)

var (
	globalrolebinding = template.Must(template.New("globalrolebinding").Parse(
		dedent.Dedent(`apiVersion: iam.kubesphere.io/v1alpha2
kind: GlobalRoleBinding
metadata:
  labels:
    iam.kubesphere.io/user-ref: {{ .Name }}
    kubefed.io/managed: "false"
  name: {{ .Name }}-platform-admin
roleRef:
  apiGroup: iam.kubesphere.io
  kind: GlobalRole
  name: {{ .Role }}
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: {{ .Name }}

    `)))
)

func generateGlobalRoleBinding(name, role string) (string, error) {
	return util.Render(globalrolebinding, util.Data{
		"Name": name,
		"Role": role,
	})
}

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	clusterrolebingdings, err := clientset.RbacV1().ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})

	for _, clusterrolebingding := range clusterrolebingdings.Items {
		if clusterrolebingding.RoleRef.Name == "cluster-admin" {
			toGlobalRoleBinding(&clusterrolebingding, clientset, "platform-admin")
		}
		if clusterrolebingding.RoleRef.Name == "cluster-regular" {
			toGlobalRoleBinding(&clusterrolebingding, clientset, "platform-regular")
		}
		if clusterrolebingding.RoleRef.Name == "workspaces-admin" {
			toGlobalRoleBinding(&clusterrolebingding, clientset, "workspaces-admin")
		}
	}
}

func toGlobalRoleBinding(binding *rbac.ClusterRoleBinding, clientset *kubernetes.Clientset, role string) {
	for _, subject := range binding.Subjects {
		if subject.APIGroup == "rbac.authorization.k8s.io" && subject.Kind == "User" {
			name := subject.Name
			objStr, err := generateGlobalRoleBinding(name, role)
			if err != nil {
				log.Fatal(err)
			}
			j2, err1 := yaml.YAMLToJSON([]byte(objStr))
			if err1 != nil {
				log.Fatal(err)
			}
			err2 := clientset.RESTClient().Post().
				AbsPath("/apis/iam.kubesphere.io/v1alpha2/globalrolebindings").
				Body(j2).
				Do(context.TODO()).Error()
			if err2 != nil {
				if kubeErr.IsAlreadyExists(err2) {
					log.Printf("%s is already exists.\n", name)
				} else {
					log.Fatal(err2)
				}
			}
			if err := clientset.RbacV1().ClusterRoleBindings().Delete(context.TODO(), binding.Name, metav1.DeleteOptions{}); err != nil {
				log.Fatal(err)
			}
		}
	}
}
