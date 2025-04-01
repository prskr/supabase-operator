package controller_test

import (
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
)

var testingScheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(testingScheme))
	utilruntime.Must(supabasev1alpha1.AddToScheme(testingScheme))
}
