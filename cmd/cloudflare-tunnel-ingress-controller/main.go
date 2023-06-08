package main

import (
	"context"
	cloudflarecontroller "github.com/STRRL/cloudflare-tunnel-ingress-controller/pkg/cloudflare-controller"
	"github.com/STRRL/cloudflare-tunnel-ingress-controller/pkg/controller"
	"github.com/cloudflare/cloudflare-go"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"
	"log"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	crlog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type rootCmdFlags struct {
	logger logr.Logger
	// for annotation on Ingress
	ingressClass string
	// for IngressClass.spec.controller
	controllerClass      string
	logLevel             int
	cloudflareAPIToken   string
	cloudflareAccountId  string
	cloudflareTunnelId   string
	cloudflareTunnelName string
}

func main() {
	var rootLogger = stdr.NewWithOptions(log.New(os.Stderr, "", log.LstdFlags), stdr.Options{LogCaller: stdr.All})

	options := rootCmdFlags{
		logger:          rootLogger.WithName("main"),
		ingressClass:    "cloudflare-tunnel",
		controllerClass: "strrl.dev/cloudflare-tunnel-ingress-controller",
		logLevel:        0,
	}

	crlog.SetLogger(rootLogger.WithName("controller-runtime"))

	rootCommand := cobra.Command{
		Use: "tunnel-controller",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			stdr.SetVerbosity(options.logLevel)
			logger := options.logger
			logger.Info("logging verbosity", "level", options.logLevel)

			if options.cloudflareTunnelName != "" && options.cloudflareTunnelId != "" {
				logger.Info("flag cloudflare-tunnel-id and cloudflare-tunnel-name are exclusive, please specify only one",
					"cloudflare-tunnel-name", options.cloudflareTunnelName,
					"cloudflare-tunnel-id", options.cloudflareTunnelId,
				)
				os.Exit(1)
			}

			logger.V(3).Info("build cloudflare client with API Token", "api-token", options.cloudflareAPIToken)
			cloudflareClient, err := cloudflare.NewWithAPIToken(options.cloudflareAPIToken)
			if err != nil {
				logger.Error(err, "create cloudflare client")
				os.Exit(1)
			}

			var tunnelClient *cloudflarecontroller.TunnelClient
			if options.cloudflareAccountId == "" {
				logger.V(3).Info("bootstrap tunnel client with tunnel id", "account-id", options.cloudflareAccountId, "tunnel-id", options.cloudflareTunnelId)
				tunnelClient = cloudflarecontroller.NewTunnelClient(logger.WithName("tunnel-client"), cloudflareClient, options.cloudflareAccountId, options.cloudflareTunnelId)
			} else {
				logger.V(3).Info("bootstrap tunnel client with tunnel name", "account-id", options.cloudflareAccountId, "tunnel-name", options.cloudflareTunnelName)
				var err error
				tunnelClient, err = cloudflarecontroller.BootstrapTunnelClientWithTunnelName(ctx, logger.WithName("tunnel-client"), cloudflareClient, options.cloudflareAccountId, options.cloudflareTunnelName)
				if err != nil {
					logger.Error(err, "bootstrap tunnel client with tunnel name")
					os.Exit(1)
				}
			}

			cfg, err := config.GetConfig()
			if err != nil {
				logger.Error(err, "unable to get kubeconfig")
				os.Exit(1)
			}
			mgr, err := manager.New(cfg, manager.Options{})
			if err != nil {
				logger.Error(err, "unable to set up manager")
				os.Exit(1)
			}

			logger.Info("cloudflare-tunnel-ingress-controller start serving")
			err = controller.RegisterIngressController(logger, mgr,
				controller.IngressControllerOptions{
					IngressClassName:    options.ingressClass,
					ControllerClassName: options.controllerClass,
					CFTunnelClient:      tunnelClient,
				})

			if err != nil {
				return err
			}
			// controller-runtime manager would graceful shutdown with signal by itself, no need to provide context
			return mgr.Start(context.Background())
		},
	}

	rootCommand.PersistentFlags().StringVar(&options.ingressClass, "ingress-class", options.ingressClass, "ingress class name")
	rootCommand.PersistentFlags().StringVar(&options.controllerClass, "controller-class", options.controllerClass, "controller class name")
	rootCommand.PersistentFlags().IntVarP(&options.logLevel, "log-level", "v", options.logLevel, "numeric log level")
	rootCommand.PersistentFlags().StringVar(&options.cloudflareAPIToken, "cloudflare-api-token", options.cloudflareAPIToken, "cloudflare api token")
	rootCommand.PersistentFlags().StringVar(&options.cloudflareAccountId, "cloudflare-account-id", options.cloudflareAccountId, "cloudflare account id")
	rootCommand.PersistentFlags().StringVar(&options.cloudflareTunnelId, "cloudflare-tunnel-id", options.cloudflareTunnelId, "cloudflare tunnel id, exclusive with cloudflare-tunnel-name")
	rootCommand.PersistentFlags().StringVar(&options.cloudflareTunnelName, "cloudflare-tunnel-name", options.cloudflareTunnelName, "cloudflare tunnel name, exclusive with cloudflare-tunnel-id")

	err := rootCommand.Execute()
	if err != nil {
		panic(err)
	}
}