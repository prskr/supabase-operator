/*
Copyright 2025 Peter Kurfer.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"github.com/alecthomas/kong"
	"go.uber.org/zap/zapcore"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	supabasev1alpha1 "github.com/prskr/supabase-operator/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(supabasev1alpha1.AddToScheme(scheme))
	utilruntime.Must(monitoringv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

type app struct {
	Manager      manager      `cmd:"" name:"manager" help:"Run the Kubernetes operator"`
	ControlPlane controlPlane `cmd:"" name:"control-plane" help:"Run the Envoy control plane"`

	Logging struct {
		Development     bool          `name:"development" default:"false"`
		Level           zapcore.Level `name:"level" default:"info"`
		StacktraceLevel zapcore.Level `name:"stacktrace-level" default:"warn"`
	} `embed:"" prefix:"logging."`
}

//nolint:unparam // signature required by kong
func (a app) AfterApply(kongctx *kong.Context) error {
	opts := zap.Options{
		Development:     a.Logging.Development,
		Level:           a.Logging.Level,
		StacktraceLevel: a.Logging.StacktraceLevel,
		TimeEncoder:     zapcore.ISO8601TimeEncoder,
	}

	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)
	kongctx.Bind(logger)

	logger.Info("Completed logger setup")

	return nil
}

func main() {
	var app app

	kongCtx := kong.Parse(
		&app,
		kong.Name("supabase-operator"),
		kong.BindTo(ctrl.SetupSignalHandler(), (*context.Context)(nil)),
	)

	if err := kongCtx.Run(); err != nil {
		setupLog.Error(err, "failed to run app")
		os.Exit(1)
	}
}
