package main

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	bootstrapv1 "sigs.k8s.io/cluster-api-bootstrap-provider-kubeadm/api/v1alpha2"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	infrastructurev1alpha3 "github.com/simonswine/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/simonswine/cluster-api-provider-hcloud/controllers"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/manifests"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/packer"
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
}{}

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = infrastructurev1alpha3.AddToScheme(scheme)
	_ = clusterv1.AddToScheme(scheme)
	_ = bootstrapv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme

	rootCmd.PersistentFlags().BoolVarP(&rootFlags.Verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().StringVar(&rootFlags.MetricsAddr, "metrics-addr", ":8484", "The address the metrics endpoint binds to.")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.EnableLeaderElection, "enable-leader-election", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	rootCmd.PersistentFlags().StringVarP(&rootFlags.ManifestsConfigPath, "manifests-config-path", "m", "", "Path to the manifests config. Disable manifest deployment if not set")
	rootCmd.PersistentFlags().StringVarP(&rootFlags.PackerConfigPath, "packer-config-path", "p", "", "Path to the packer config. Disable image building if not set")
}

var rootCmd = &cobra.Command{
	Use:  "cluster-api-provider-hcloud",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ctrl.SetLogger(zap.New(func(o *zap.Options) {
			o.Development = rootFlags.Verbose
		}))

		// Initialise manifests generator
		manifestsMgr := manifests.New(ctrl.Log.WithName("module").WithName("manifests"), rootFlags.ManifestsConfigPath)
		if err := manifestsMgr.Initialize(); err != nil {
			setupLog.Error(err, "unable to initialise manifests manager")
			os.Exit(1)
		}

		// Initialise packer generator
		packerMgr := packer.New(ctrl.Log.WithName("module").WithName("packer"), rootFlags.PackerConfigPath)
		if err := packerMgr.Initialize(); err != nil {
			setupLog.Error(err, "unable to initialise packer manager")
			os.Exit(1)
		}

		mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:             scheme,
			MetricsBindAddress: rootFlags.MetricsAddr,
			LeaderElection:     rootFlags.EnableLeaderElection,
			Port:               9443,
		})
		if err != nil {
			setupLog.Error(err, "unable to start manager")
			os.Exit(1)
		}

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
