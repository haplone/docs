
# 使用cobra来处理系统启动的命令行交互

* main.go中import 	"github.com/cockroachdb/cockroach/pkg/cli"
* 在cli/cli.go中使用init方法对md进行初始化

```go
// cli/cli.go
func init() {
	// ...

	cockroachCmd.AddCommand(
		StartCmd,
		initCmd,
		certCmd,
		quitCmd,

		sqlShellCmd,
		userCmd,
		zoneCmd,
		nodeCmd,
		dumpCmd,

		// Miscellaneous commands.
		genCmd,
		versionCmd,
		debugCmd,
	)
}
```
* 在cli/start.go中有StartCmd的定义

```go
// cli/start.go
var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "start a node",
	Long: `
Start a CockroachDB node, which will export data from one or more
storage devices, specified via --store flags.

If no cluster exists yet and this is the first node, no additional
flags are required. If the cluster already exists, and this node is
uninitialized, specify the --join flag to point to any healthy node
(or list of nodes) already part of the cluster.
`,
	Example: `  cockroach start --insecure --store=attrs=ssd,path=/mnt/ssd1 [--join=host:port,[host:port]]`,
	RunE:    MaybeShoutError(MaybeDecorateGRPCError(runStart)),
}
```
* 然后系统才开始通过runStart接手进行系统初始化

# 启动server

* 一堆的参数检查设置、log准备
* new 一个server
* 启动server

```go
// cli/start.go
func runStart(cmd *cobra.Command, args []string) error {
	// ...

	// 使用分布式追踪协议opentracing
	tracer := serverCfg.Settings.Tracer
	sp := tracer.StartSpan("server start")
	ctx := opentracing.ContextWithSpan(context.Background(), sp)

	// Set up the logging and profiling output.
	stopper, err := setupAndInitializeLoggingAndProfiling(ctx)
  // ...
	// grpc 出场
	grpcutil.SetSeverity(log.Severity_WARNING)

	// ...
	log.Info(ctx, "starting cockroach node")

	// ...
	var s *server.Server
	errChan := make(chan error, 1)
	go func() {
		// ...
		if err := func() error {
			// Instantiate the server.
			var err error
			s, err = server.NewServer(serverCfg, stopper)
			// ...

			// Attempt to start the server.
			if err := s.Start(ctx); err != nil {
				// ...
			}
			// ...
			// 一堆的善后处理
	}()
	// ... 应用关闭的处理

	return returnErr
}
```

# new server


```go
// server/server.go
// NewServer creates a Server from a server.Config.
func NewServer(cfg Config, stopper *stop.Stopper) (*Server, error) {
	// ...
	s := &Server{
		// ...
	}
	// ...
	s.grpc = rpc.NewServerWithInterceptor(s.rpcContext, s.Intercept())

	s.gossip = gossip.New(
		// ...
	)

	// ...
	s.distSender = kv.NewDistSender(distSenderCfg, s.gossip)
	// ...
	s.db = client.NewDBWithContext(s.tcsFactory, s.clock, dbCtx)
	// ...
	s.nodeLiveness = storage.NewNodeLiveness(
		// ...
	)
  // ...
	s.storePool = storage.NewStorePool(
		// ...
	)

  //
	s.raftTransport = storage.NewRaftTransport(
		s.cfg.AmbientCtx, st, storage.GossipAddressResolver(s.gossip), s.grpc, s.rpcContext,
	)
	// ...
	s.leaseMgr = sql.NewLeaseManager(
		// ...
	)

	// ...
	rootSQLMemoryMonitor.Start(context.Background(), nil, mon.MakeStandaloneBudget(s.cfg.SQLMemoryPoolSize))

	// Set up the DistSQL temp engine.
	tempEngine, err := engine.NewTempEngine(s.cfg.TempStorageConfig)

	s.stopper.AddCloser(tempEngine)
	// Remove temporary directory linked to tempEngine after closing
	// tempEngine.
	s.stopper.AddCloser(stop.CloserFn(func() {
    // ...
		err = os.RemoveAll(s.cfg.TempStorageConfig.Path)
    // ...
	}))

	// Set up admin memory metrics for use by admin SQL executors.
	s.adminMemMetrics = sql.MakeMemMetrics("admin", cfg.HistogramWindowInterval())
	s.registry.AddMetricStruct(s.adminMemMetrics)

  //
	s.tsDB = ts.NewDB(s.db, s.cfg.Settings)
  //
	s.tsServer = ts.MakeServer(s.cfg.AmbientCtx, s.tsDB, nodeCountFn, s.cfg.TimeSeriesServerConfig, s.stopper)

	// The InternalExecutor will be further initialized later, as we create more
	// of the server's components. There's a circular dependency - many things
	// need an InternalExecutor, but the InternalExecutor needs an ExecutorConfig,
	// which in turn needs many things. That's why everybody that needs an
	// InternalExecutor takes pointers to this one instance.
	sqlExecutor := sql.InternalExecutor{}

	// ...
	s.node = NewNode(
		storeCfg, s.recorder, s.registry, s.stopper,
		txnMetrics, nil /* execCfg */, &s.rpcContext.ClusterID)

	s.sessionRegistry = sql.MakeSessionRegistry()
	s.jobRegistry = jobs.MakeRegistry(
		s.cfg.AmbientCtx, s.clock, s.db, &sqlExecutor, &s.nodeIDContainer, st, func(opName, user string) (interface{}, func()) {
			// This is a hack to get around a Go package dependency cycle. See comment
			// in sql/jobs/registry.go on planHookMaker.
			return sql.NewInternalPlanner(opName, nil, user, &sql.MemoryMetrics{}, &execCfg)
		})

	// Set up the DistSQL server.
	distSQLCfg := distsqlrun.ServerConfig{
		// ...
	}

	s.distSQLServer = distsqlrun.NewServer(ctx, distSQLCfg)
	distsqlrun.RegisterDistSQLServer(s.grpc, s.distSQLServer)

	s.admin = newAdminServer(s, &sqlExecutor)
	s.status = newStatusServer(
		// ...
	)
	s.authentication = newAuthenticationServer(s, &sqlExecutor)
	for _, gw := range []grpcGatewayServer{s.admin, s.status, s.authentication, &s.tsServer} {
		gw.RegisterService(s.grpc)
	}
  //
	s.initServer = newInitServer(s)
	s.initServer.semaphore.acquire()

	serverpb.RegisterInitServer(s.grpc, s.initServer)

	//
	virtualSchemas, err := sql.NewVirtualSchemaHolder(ctx, st)
	if err != nil {
		log.Fatal(ctx, err)
	}

	// Set up Executor
	s.sqlExecutor = sql.NewExecutor(execCfg, s.stopper)
	if s.cfg.UseLegacyConnHandling {
		s.registry.AddMetricStruct(s.sqlExecutor)
	}

	s.pgServer = pgwire.MakeServer(
		// ...
	)
	// ...

	sqlExecutor.ExecCfg = &execCfg
	s.execCfg = &execCfg

	s.leaseMgr.SetExecCfg(&execCfg)
	s.leaseMgr.RefreshLeases(s.stopper, s.db, s.gossip)

	s.node.InitLogger(&execCfg)

	return s, nil
}

```
