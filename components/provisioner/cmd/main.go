package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kyma-project/control-plane/components/provisioner/internal/apperrors"

	"github.com/kyma-project/control-plane/components/provisioner/internal/metrics"

	"github.com/kyma-project/control-plane/components/provisioner/internal/util/k8s"

	provisioningStages "github.com/kyma-project/control-plane/components/provisioner/internal/operations/stages/provisioning"

	retry "github.com/avast/retry-go"

	"github.com/kyma-project/control-plane/components/provisioner/internal/model"
	"github.com/kyma-project/control-plane/components/provisioner/internal/operations/queue"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/control-plane/components/provisioner/internal/healthz"

	"github.com/kyma-project/control-plane/components/provisioner/internal/api/middlewares"
	"github.com/kyma-project/control-plane/components/provisioner/internal/runtime"

	installationSDK "github.com/kyma-incubator/hydroform/install/installation"
	"github.com/kyma-project/control-plane/components/provisioner/internal/api"
	"github.com/kyma-project/control-plane/components/provisioner/internal/installation"

	"github.com/kyma-project/control-plane/components/provisioner/internal/persistence/database"
	"github.com/kyma-project/control-plane/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/kyma-project/control-plane/components/provisioner/internal/uuid"

	"github.com/kyma-project/control-plane/components/provisioner/internal/gardener"

	"github.com/kyma-project/control-plane/components/provisioner/internal/installation/release"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-project/control-plane/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/kubernetes"
)

const connStringFormat string = "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s"

type config struct {
	Address                      string `envconfig:"default=127.0.0.1:3000"`
	APIEndpoint                  string `envconfig:"default=/graphql"`
	PlaygroundAPIEndpoint        string `envconfig:"default=/graphql"`
	DirectorURL                  string `envconfig:"default=http://compass-director.compass-system.svc.cluster.local:3000/graphql"`
	SkipDirectorCertVerification bool   `envconfig:"default=false"`
	OauthCredentialsNamespace    string `envconfig:"default=kcp-system"`
	OauthCredentialsSecretName   string `envconfig:"default=kcp-provisioner-credentials"`

	Database struct {
		User     string `envconfig:"default=postgres"`
		Password string `envconfig:"default=password"`
		Host     string `envconfig:"default=localhost"`
		Port     string `envconfig:"default=5432"`
		Name     string `envconfig:"default=provisioner"`
		SSLMode  string `envconfig:"default=disable"`
	}

	ProvisioningTimeout   queue.ProvisioningTimeouts
	DeprovisioningTimeout queue.DeprovisioningTimeouts
	HibernationTimeout    queue.HibernationTimeouts

	OperatorRoleBinding provisioningStages.OperatorRoleBinding

	Gardener struct {
		Project                                    string `envconfig:"default=gardenerProject"`
		KubeconfigPath                             string `envconfig:"default=./dev/kubeconfig.yaml"`
		AuditLogsPolicyConfigMap                   string `envconfig:"optional"`
		AuditLogsTenantConfigPath                  string `envconfig:"optional"`
		MaintenanceWindowConfigPath                string `envconfig:"optional"`
		ClusterCleanupResourceSelector             string `envconfig:"default=https://service-manager."`
		DefaultEnableKubernetesVersionAutoUpdate   bool   `envconfig:"default=false"`
		DefaultEnableMachineImageVersionAutoUpdate bool   `envconfig:"default=false"`
		ForceAllowPrivilegedContainers             bool   `envconfig:"default=false"`
	}

	LatestDownloadedReleases int  `envconfig:"default=5"`
	DownloadPreReleases      bool `envconfig:"default=true"`

	EnqueueInProgressOperations bool `envconfig:"default=true"`

	MetricsAddress string `envconfig:"default=127.0.0.1:9000"`

	LogLevel string `envconfig:"default=info"`
}

func (c *config) String() string {
	return fmt.Sprintf("Address: %s, APIEndpoint: %s, DirectorURL: %s, "+
		"SkipDirectorCertVerification: %v, OauthCredentialsNamespace: %s, OauthCredentialsSecretName: %s, "+
		"DatabaseUser: %s, DatabaseHost: %s, DatabasePort: %s, "+
		"DatabaseName: %s, DatabaseSSLMode: %s, "+
		"ProvisioningTimeoutClusterCreation: %s "+
		"ProvisioningTimeoutInstallation: %s, ProvisioningTimeoutUpgrade: %s, "+
		"ProvisioningTimeoutAgentConfiguration: %s, ProvisioningTimeoutAgentConnection: %s, "+
		"DeprovisioningTimeoutClusterDeletion: %s, DeprovisioningTimeoutWaitingForClusterDeletion: %s "+
		"OperatorRoleBindingL2SubjectName: %s, OperatorRoleBindingL3SubjectName: %s, OperatorRoleBindingCreatingForAdmin: %t"+
		"GardenerProject: %s, GardenerKubeconfigPath: %s, GardenerAuditLogsPolicyConfigMap: %s, AuditLogsTenantConfigPath: %s, "+
		"ForceAllowPrivilegedContainers: %t, "+
		"LatestDownloadedReleases: %d, DownloadPreReleases: %v, "+
		"EnqueueInProgressOperations: %v"+
		"LogLevel: %s",
		c.Address, c.APIEndpoint, c.DirectorURL,
		c.SkipDirectorCertVerification, c.OauthCredentialsNamespace, c.OauthCredentialsSecretName,
		c.Database.User, c.Database.Host, c.Database.Port,
		c.Database.Name, c.Database.SSLMode,
		c.ProvisioningTimeout.ClusterCreation.String(),
		c.ProvisioningTimeout.Installation.String(), c.ProvisioningTimeout.Upgrade.String(),
		c.ProvisioningTimeout.AgentConfiguration.String(), c.ProvisioningTimeout.AgentConnection.String(),
		c.DeprovisioningTimeout.ClusterDeletion.String(), c.DeprovisioningTimeout.WaitingForClusterDeletion.String(),
		c.OperatorRoleBinding.L2SubjectName, c.OperatorRoleBinding.L3SubjectName, c.OperatorRoleBinding.CreatingForAdmin,
		c.Gardener.Project, c.Gardener.KubeconfigPath, c.Gardener.AuditLogsPolicyConfigMap, c.Gardener.AuditLogsTenantConfigPath,
		c.Gardener.ForceAllowPrivilegedContainers,
		c.LatestDownloadedReleases, c.DownloadPreReleases,
		c.EnqueueInProgressOperations,
		c.LogLevel)
}

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Failed to load application config")

	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Warnf("Invalid log level: '%s', defaulting to 'info'", cfg.LogLevel)
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)

	log.Infof("Starting Provisioner")
	log.Infof("Config: %s", cfg.String())

	connString := fmt.Sprintf(connStringFormat, cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode)

	gardenerNamespace := fmt.Sprintf("garden-%s", cfg.Gardener.Project)

	gardenerClusterConfig, err := newGardenerClusterConfig(cfg)
	exitOnError(err, "Failed to initialize Gardener cluster client")

	gardenerClientSet, err := gardener.NewClient(gardenerClusterConfig)
	exitOnError(err, "Failed to create Gardener cluster clientset")

	k8sCoreClientSet, err := kubernetes.NewForConfig(gardenerClusterConfig)
	exitOnError(err, "Failed to create Kubernetes clientset")

	secretsInterface := k8sCoreClientSet.CoreV1().Secrets(gardenerNamespace)

	shootClient := gardenerClientSet.Shoots(gardenerNamespace)

	connection, err := database.InitializeDatabaseConnection(connString, databaseConnectionRetries)
	exitOnError(err, "Failed to initialize persistence")

	installationHandlerConstructor := func(c *rest.Config, o ...installationSDK.InstallationOption) (installationSDK.Installer, error) {
		return installationSDK.NewKymaInstaller(c, o...)
	}

	dbsFactory := dbsession.NewFactory(connection)
	installationService := installation.NewInstallationService(cfg.ProvisioningTimeout.Installation, installationHandlerConstructor, cfg.Gardener.ClusterCleanupResourceSelector)

	directorClient, err := newDirectorClient(cfg)
	exitOnError(err, "Failed to initialize Director client")

	k8sClientProvider := k8s.NewK8sClientProvider()

	runtimeConfigurator := runtime.NewRuntimeConfigurator(k8sClientProvider, directorClient)

	provisioningQueue := queue.CreateProvisioningQueue(
		cfg.ProvisioningTimeout,
		dbsFactory,
		installationService,
		runtimeConfigurator,
		provisioningStages.NewCompassConnectionClient,
		directorClient,
		shootClient,
		secretsInterface,
		cfg.OperatorRoleBinding,
		k8sClientProvider)

	upgradeQueue := queue.CreateUpgradeQueue(cfg.ProvisioningTimeout, dbsFactory, directorClient, installationService)

	deprovisioningQueue := queue.CreateDeprovisioningQueue(cfg.DeprovisioningTimeout, dbsFactory, installationService, directorClient, shootClient, 5*time.Minute)

	shootUpgradeQueue := queue.CreateShootUpgradeQueue(cfg.ProvisioningTimeout, dbsFactory, directorClient, shootClient, cfg.OperatorRoleBinding, k8sClientProvider)

	hibernationQueue := queue.CreateHibernationQueue(cfg.HibernationTimeout, dbsFactory, directorClient, shootClient)

	provisioner := gardener.NewProvisioner(gardenerNamespace, shootClient, dbsFactory, cfg.Gardener.AuditLogsPolicyConfigMap, cfg.Gardener.MaintenanceWindowConfigPath)
	shootController, err := newShootController(gardenerNamespace, gardenerClusterConfig, dbsFactory, cfg.Gardener.AuditLogsTenantConfigPath)
	exitOnError(err, "Failed to create Shoot controller.")
	go func() {
		err := shootController.StartShootController()
		exitOnError(err, "Failed to start Shoot Controller")
	}()

	httpClient := newHTTPClient(false)
	fileDownloader := release.NewFileDownloader(httpClient)

	releaseRepository := release.NewReleaseRepository(connection, uuid.NewUUIDGenerator())
	gcsDownloader := release.NewGCSDownloader(fileDownloader)

	releaseProvider := release.NewReleaseProvider(releaseRepository, gcsDownloader)

	provisioningSVC := newProvisioningService(
		cfg.Gardener.Project,
		provisioner,
		dbsFactory,
		releaseProvider,
		directorClient,
		provisioningQueue,
		deprovisioningQueue,
		upgradeQueue,
		shootUpgradeQueue,
		hibernationQueue,
		cfg.Gardener.DefaultEnableKubernetesVersionAutoUpdate,
		cfg.Gardener.DefaultEnableMachineImageVersionAutoUpdate,
		cfg.Gardener.ForceAllowPrivilegedContainers)

	validator := api.NewValidator(dbsFactory.NewReadSession())
	resolver := api.NewResolver(provisioningSVC, validator)
	logger := log.WithField("Component", "Artifact Downloader")
	downloader := release.NewArtifactsDownloader(releaseRepository, cfg.LatestDownloadedReleases, cfg.DownloadPreReleases, httpClient, fileDownloader, logger)

	// Run release downloader
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go downloader.FetchPeriodically(ctx, release.ShortInterval, release.LongInterval)

	provisioningQueue.Run(ctx.Done())

	deprovisioningQueue.Run(ctx.Done())

	upgradeQueue.Run(ctx.Done())

	shootUpgradeQueue.Run(ctx.Done())

	hibernationQueue.Run(ctx.Done())

	gqlCfg := gqlschema.Config{
		Resolvers: resolver,
	}
	executableSchema := gqlschema.NewExecutableSchema(gqlCfg)

	presenter := apperrors.NewPresenter(log.StandardLogger())

	log.Infof("Registering endpoint on %s...", cfg.APIEndpoint)
	router := mux.NewRouter()
	router.Use(middlewares.ExtractTenant)

	router.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	router.HandleFunc(cfg.APIEndpoint, handler.GraphQL(executableSchema, handler.ErrorPresenter(presenter.Do)))
	router.HandleFunc("/healthz", healthz.NewHTTPHandler(log.StandardLogger()))

	// Metrics
	err = metrics.Register(dbsFactory.NewReadSession())
	exitOnError(err, "Failed to register metrics collectors")

	// Expose metrics on different port as it cannot be secured with mTLS
	metricsRouter := mux.NewRouter()
	metricsRouter.Handle("/metrics", promhttp.Handler())

	metricsServer := &http.Server{
		Handler: metricsRouter,
		Addr:    cfg.MetricsAddress,
	}

	log.Infof("API listening on %s...", cfg.Address)
	log.Infof("Metrics API listening on %s...", cfg.MetricsAddress)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := http.ListenAndServe(cfg.Address, router); err != nil {
			log.Errorf("Error starting server: %s", err.Error())
		}
	}()

	go func() {
		if err := metricsServer.ListenAndServe(); err != nil {
			log.Errorf("Error starting metrics server: %s", err.Error())
		}
	}()

	if cfg.EnqueueInProgressOperations {
		err = enqueueOperationsInProgress(dbsFactory, provisioningQueue, deprovisioningQueue, upgradeQueue, shootUpgradeQueue, hibernationQueue)
		exitOnError(err, "Failed to enqueue in progress operations")
	}

	wg.Wait()
}

func enqueueOperationsInProgress(dbFactory dbsession.Factory, provisioningQueue, deprovisioningQueue, upgradeQueue, shootUpgradeQueue, hibernationQueue queue.OperationQueue) error {
	readSession := dbFactory.NewReadSession()

	var inProgressOps []model.Operation
	var err error

	// Due to Schema Migrator running post upgrade the pod will be in crash loop back off and Helm deployment will not finish
	// therefor we need to wait for schema to be initialized in case of blank installation
	err = retry.Do(func() error {
		inProgressOps, err = readSession.ListInProgressOperations()
		if err != nil {
			log.Warnf("failed to list in progress operation")
			return err
		}
		return nil
	}, retry.Attempts(30), retry.DelayType(retry.FixedDelay), retry.Delay(5*time.Second))
	if err != nil {
		return fmt.Errorf("error enqueuing in progress operations: %s", err.Error())
	}

	for _, op := range inProgressOps {
		if op.Type == model.Provision {
			provisioningQueue.Add(op.ID)
			continue
		}

		if op.Type == model.Deprovision {
			deprovisioningQueue.Add(op.ID)
		}

		if op.Type == model.Upgrade {
			upgradeQueue.Add(op.ID)
		}

		if op.Type == model.UpgradeShoot {
			shootUpgradeQueue.Add(op.ID)
		}

		if op.Type == model.Hibernate {
			hibernationQueue.Add(op.ID)
		}
	}

	return nil
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}
