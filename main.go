/*

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
	"flag"
	"fmt"
	"github.com/ory/keto-maester/keto"
	"net/http"
	"net/url"
	"os"
	"time"

	ketov1alpha1 "github.com/ory/keto-maester/api/v1alpha1"
	"github.com/ory/keto-maester/controllers"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {

	apiv1.AddToScheme(scheme)
	ketov1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var (
		metricsAddr, ketoURL, forwardedProto, syncPeriod string
		ketoPort                                         int
		enableLeaderElection                             bool
	)

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&ketoURL, "keto-url", "", "The address of ORY Hydra")
	flag.IntVar(&ketoPort, "keto-port", 4456, "Port ORY Keto is listening on")
	flag.StringVar(&forwardedProto, "forwarded-proto", "", "If set, this adds the value as the X-Forwarded-Proto header in requests to the ORY Keto admin server")
	flag.StringVar(&syncPeriod, "sync-period", "10h", "Determines the minimum frequency at which watched resources are reconciled")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	syncPeriodParsed, err := time.ParseDuration(syncPeriod)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		SyncPeriod:         &syncPeriodParsed,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if ketoURL == "" {
		setupLog.Error(fmt.Errorf("keto URL can't be empty"), "unable to create controller", "controller", "Policy")
		os.Exit(1)
	}

	u, err := url.Parse(fmt.Sprintf("%s:%d", ketoURL, ketoPort))
	if err != nil {
		setupLog.Error(fmt.Errorf("keto URL must be valid url"), "unable to create controller", "controller", "Policy")
		os.Exit(1)
	}

	ketoClient := &keto.Client{
		KetoURL:        *u,
		HTTPClient:     &http.Client{},
		ForwardedProto: forwardedProto,
	}

	err = (&controllers.KetoPolicyReconciler{Reconciler: &controllers.Reconciler{
		Client:     mgr.GetClient(),
		Log:        ctrl.Log.WithName("controllers").WithName("Policy"),
		KetoClient: ketoClient,
	}}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Policy")
		os.Exit(1)
	}

	err = (&controllers.KetoRoleReconciler{Reconciler: &controllers.Reconciler{
		Client:     mgr.GetClient(),
		Log:        ctrl.Log.WithName("controllers").WithName("Role"),
		KetoClient: ketoClient,
	}}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Role")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
