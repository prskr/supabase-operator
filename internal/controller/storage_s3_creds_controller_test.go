package controller_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
	"code.icb4dc0.de/prskr/supabase-operator/internal/controller"
)

func TestStorageS3CredsController(t *testing.T) {
	storage := new(supabasev1alpha1.Storage)

	MustReadObject(t, filepath.Join("testdata", "storage_basic.yaml"), storage)

	t.Log(string(MustMarshal(t, storage)))

	clientSet := fake.NewClientBuilder().
		WithObjects(storage).
		WithScheme(testingScheme).
		Build()

	log.SetLogger(zap.New(zap.UseDevMode(true)))

	controllerReconciler := &controller.StorageS3CredentialsReconciler{
		Client: clientSet,
		Scheme: testingScheme,
	}

	_, err := controllerReconciler.Reconcile(t.Context(), reconcile.Request{
		NamespacedName: client.ObjectKeyFromObject(storage),
	})

	assert.NoError(t, err)
}

func MustMarshal(t testing.TB, obj runtime.Object) []byte {
	t.Helper()

	raw, err := json.Marshal(obj)
	if err != nil {
		t.Errorf("failed to marshal object: %v", err)
		t.Fail()
	}
	return raw
}

func MustReadObject(t testing.TB, relativePath string, obj runtime.Object) {
	t.Helper()

	gkvs, _, err := testingScheme.ObjectKinds(obj)
	if err != nil {
		t.Errorf("failed to determine group version kind: %v", err)
		t.Fatal(err)
	}

	decoder := scheme.Codecs.UniversalDecoder(testingScheme.PreferredVersionAllGroups()...)

	raw, err := os.ReadFile(relativePath)
	if err != nil {
		t.Errorf("failed to read file %s: %v", relativePath, err)
		t.Fail()
	}

	_, _, err = decoder.Decode(raw, &gkvs[0], obj)
	if err != nil {
		t.Errorf("failed to unmarshal file %s: %v", relativePath, err)
		t.Fail()
	}
}
