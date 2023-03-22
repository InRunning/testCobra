package pcap

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var K8sClient client.Client

type S3Config struct {
	AccessKeyId     string
	SecretAccessKey string
	Endpoint        string
}

func init() {
	k8sClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: runtime.NewScheme()})
	if err != nil {
		log.Panic(err)
	}
	K8sClient = k8sClient
}

func GetS3Config() (*S3Config, error) {
	// get CRD object by name
	// crdName := "smf.axyom.casa-systems.io/v1alpha1.SMF"
	ctx := context.Background()

	crd := &unstructured.Unstructured{}
	// create API resource reference
	crd.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "smf.axyom.casa-systems.io",
		Version: "v1alpha1",
		Kind:    "SMF",
	})
	err := K8sClient.Get(ctx, client.ObjectKey{
		Name: "smf",
	}, crd)
	if err != nil {
		return nil, err
	}
	s3Storage, _, err := unstructured.NestedMap(crd.Object, "spec", "tracing", "pcapConfig", "s3Storage")
	if err != nil {
		return nil, err
	}
	fmt.Println("S3 storage: ", s3Storage)
}
