package main

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	infrav1alpha3 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/controllers"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/manifests"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/packer"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/record"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

var rootFlags = struct {
	MetricsAddr          string
	EnableLeaderElection bool
	Verbose              bool
	ManifestsConfigPath  string
	PackerConfigPath     string
	WebhookPort          int
}{}

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = infrav1alpha3.AddToScheme(scheme)
	_ = clusterv1.AddToScheme(scheme)
	_ = bootstrapv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme

	rootCmd.PersistentFlags().BoolVarP(&rootFlags.Verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().StringVar(&rootFlags.MetricsAddr, "metrics-addr", ":8484", "The address the metrics endpoint binds to.")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.EnableLeaderElection, "enable-leader-election", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	rootCmd.PersistentFlags().StringVarP(&rootFlags.ManifestsConfigPath, "manifests-config-path", "m", "", "Path to the manifests config. Disable manifest deployment if not set")
	rootCmd.PersistentFlags().StringVarP(&rootFlags.PackerConfigPath, "packer-config-path", "p", "", "Path to the packer config. Disable image building if not set")
	rootCmd.PersistentFlags().IntVar(&rootFlags.WebhookPort, "webhook-port", 0, "Webhook Server port, disabled by default. When enabled, the manager will only work as webhook server, no reconcilers are installed.")
}

var rootCmd = &cobra.Command{
	Use:  "cluster-api-provider-hcloud",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ctrl.SetLogger(zap.New(func(o *zap.Options) {
			o.Development = rootFlags.Verbose
		}))

		mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:             scheme,
			MetricsBindAddress: rootFlags.MetricsAddr,
			LeaderElection:     rootFlags.EnableLeaderElection,
			LeaderElectionID:   "capi-hcloud-leader-election",
			Port:               rootFlags.WebhookPort,
		})
		if err != nil {
			setupLog.Error(err, "unable to start manager")
			os.Exit(1)
		}

		// Initialize event recorder.
		record.InitFromRecorder(mgr.GetEventRecorderFor("hcloud-controller"))

		if rootFlags.WebhookPort == 0 {
			// run in controller mode

			// Initialise manifests generator
			manifestsMgr := manifests.New(ctrl.Log.WithName("module").WithName("manifests"), rootFlags.ManifestsConfigPath)
			if err := manifestsMgr.Initialize(); err != nil {
				setupLog.Error(err, "unable to initialise manifests manager")
				os.Exit(1)
			}
			//Packer generator, initialization happens in cluster controller
			packerMgr := packer.New(ctrl.Log.WithName("module").WithName("packer"))

			if err = (&controllers.HcloudClusterReconciler{
				Client:    mgr.GetClient(),
				Log:       ctrl.Log.WithName("controllers").WithName("HcloudCluster"),
				Scheme:    mgr.GetScheme(),
				Packer:    packerMgr,
				Manifests: manifestsMgr,
			}).SetupWithManager(mgr, controller.Options{}); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", "HcloudCluster")
				os.Exit(1)
			}
			if err = (&controllers.HcloudMachineReconciler{
				Client:    mgr.GetClient(),
				Log:       ctrl.Log.WithName("controllers").WithName("HcloudMachine"),
				Scheme:    mgr.GetScheme(),
				Packer:    packerMgr,
				Manifests: manifestsMgr,
			}).SetupWithManager(mgr, controller.Options{}); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", "HcloudMachine")
				os.Exit(1)
			}
			if err = (&controllers.HcloudVolumeReconciler{
				Client:    mgr.GetClient(),
				Log:       ctrl.Log.WithName("controllers").WithName("HcloudVolume"),
				Scheme:    mgr.GetScheme(),
				Packer:    packerMgr,
				Manifests: manifestsMgr,
			}).SetupWithManager(mgr, controller.Options{}); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", "HcloudVolume")
				os.Exit(1)
			}
			// +kubebuilder:scaffold:builder
		} else {
			// run in webhook mode

			type webhookSetuper interface {
				SetupWebhookWithManager(manager.Manager) error
			}

			for _, t := range []webhookSetuper{
				&infrav1alpha3.HcloudCluster{},
				&infrav1alpha3.HcloudClusterList{},
				&infrav1alpha3.HcloudMachine{},
				&infrav1alpha3.HcloudMachineList{},
			} {
				if err = t.SetupWebhookWithManager(mgr); err != nil {
					setupLog.Error(err, "unable to create webhook", "webhook", "HcloudCluster")
					os.Exit(1)
				}
			}
		}

		setupLog.Info("starting manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			setupLog.Error(err, "problem running manager")
			os.Exit(1)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		setupLog.Error(err, "problem executing rootCmd")
		os.Exit(1)
	}
}
