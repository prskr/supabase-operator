package supabase

import (
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
	corev1 "k8s.io/api/core/v1"
)

func fixedEnvOf(key, value string) fixedEnv {
	return fixedEnvFunc(func() corev1.EnvVar {
		return corev1.EnvVar{
			Name:  key,
			Value: value,
		}
	})
}

type fixedEnvFunc func() corev1.EnvVar

func (f fixedEnvFunc) Var() corev1.EnvVar {
	return f()
}

type fixedEnv interface {
	Var() corev1.EnvVar
}

type stringEnv string

func (e stringEnv) Var(value string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  string(e),
		Value: value,
	}
}

type stringSliceEnv struct {
	key       string
	separator string
}

func (e stringSliceEnv) Var(value []string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  e.key,
		Value: strings.Join(value, e.separator),
	}
}

type intEnv[T constraints.Integer] string

func (e intEnv[T]) Var(value T) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  string(e),
		Value: strconv.FormatInt(int64(value), 10),
	}
}

type boolEnv string

func (e boolEnv) Var(value bool) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  string(e),
		Value: strconv.FormatBool(value),
	}
}

type secretEnv string

func (e secretEnv) Var(sel *corev1.SecretKeySelector) corev1.EnvVar {
	return corev1.EnvVar{
		Name: string(e),
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: sel,
		},
	}
}
