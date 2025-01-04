package api

import (
	"iter"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectList[T metav1.Object] interface {
	client.ObjectList
	Iter() iter.Seq[T]
}
