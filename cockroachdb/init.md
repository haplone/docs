
# 使用cobra来处理系统启动的命令行交互

* main.go中import 	"github.com/cockroachdb/cockroach/pkg/cli"
* 在cli/cli.go中使用init方法对cmd进行初始化

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
	// 使用分布式追踪协议opentracing
	tracer := serverCfg.Settings.Tracer
	sp := tracer.StartSpan("server start")
	ctx := opentracing.ContextWithSpan(context.Background(), sp)

	// Set up the logging and profiling output.
 	// stopper ???
	stopper, err := setupAndInitializeLoggingAndProfiling(ctx)
 	// ...
	// grpc 出场
	grpcutil.SetSeverity(log.Severity_WARNING)
	log.Info(ctx, "starting cockroach node")
	var s *server.Server
	errChan := make(chan error, 1)
	go func() {
		if err := func() error {
			// Instantiate the server.
			s, err = server.NewServer(serverCfg, stopper)
			// Attempt to start the server.
			if err := s.Start(ctx); err != nil {
				// ...
			}
			// ... 一堆的善后处理
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
	// 初始化server
	s := &Server{
		// ...
	}
	// ... 带拦截器的grpc
	s.grpc = rpc.NewServerWithInterceptor(s.rpcContext, s.Intercept())
	
	// gossip相关初始化
	s.gossip = gossip.New(
		// ...
	)

	// ...A DistSender provides methods to access Cockroach's monolithic, distributed key value store
	s.distSender = kv.NewDistSender(distSenderCfg, s.gossip)
	// ... ???
	s.tcsFactory = kv.NewTxnCoordSenderFactory(
		// ...
	 )
	// ... DB is a database handle to a single cockroach cluster
	s.db = client.NewDBWithContext(s.tcsFactory, s.clock, dbCtx)
	// ... NodeLiveness encapsulates information on node liveness and provides
	// an API for querying, updating, and invalidating node
	// liveness
	s.nodeLiveness = storage.NewNodeLiveness(
		// ...
	)
	// ... StorePool maintains a list of all known stores in the cluster and information on their health.
	s.storePool = storage.NewStorePool(
		// ...
	)

	// RaftTransport handles the rpc messages for raft
	s.raftTransport = storage.NewRaftTransport(
		s.cfg.AmbientCtx, st, storage.GossipAddressResolver(s.gossip), s.grpc, s.rpcContext,
	)
	// ... LeaseManager manages acquiring and releasing per-table leases
	s.leaseMgr = sql.NewLeaseManager(
		// ...
	)

	// ... Start begins a monitoring region.

	// creates a new engine for DistSQL processors to use when the
	// working set is larger than can be stored in memory
	tempEngine, err := engine.NewTempEngine(s.cfg.TempStorageConfig)
	// 应用关闭时的回调注册
	s.stopper.AddCloser(tempEngine)
	// Remove temporary directory linked to tempEngine after closing
	// tempEngine.
	s.stopper.AddCloser(stop.CloserFn(func() {
		err = os.RemoveAll(s.cfg.TempStorageConfig.Path)
	}))

	// Set up admin memory metrics for use by admin SQL executors.
	s.adminMemMetrics = sql.MakeMemMetrics("admin", cfg.HistogramWindowInterval())

	// 初始化时间序列DB ???
	s.tsDB = ts.NewDB(s.db, s.cfg.Settings)
	// 初始化时间序列DB对应的server
	s.tsServer = ts.MakeServer(s.cfg.AmbientCtx, s.tsDB, nodeCountFn, s.cfg.TimeSeriesServerConfig, s.stopper)

	// The InternalExecutor will be further initialized later, as we create more
	// of the server's components. There's a circular dependency - many things
	// need an InternalExecutor, but the InternalExecutor needs an ExecutorConfig,
	// which in turn needs many things. That's why everybody that needs an
	// InternalExecutor takes pointers to this one instance.
	sqlExecutor := sql.InternalExecutor{}

	// ... A Node manages a map of stores (by store ID) for which it serves traffic.
	s.node = NewNode(
		storeCfg, s.recorder, s.registry, s.stopper,
		txnMetrics, nil /* execCfg */, &s.rpcContext.ClusterID)
	// SessionRegistry stores a set of all sessions on this node
	s.sessionRegistry = sql.MakeSessionRegistry()
	// Registry creates Jobs and manages their leases and cancelation.
	s.jobRegistry = jobs.MakeRegistry(
		s.cfg.AmbientCtx, s.clock, s.db, &sqlExecutor, &s.nodeIDContainer, st, func(opName, user string) (interface{}, func()) {
			// This is a hack to get around a Go package dependency cycle. See comment
			// in sql/jobs/registry.go on planHookMaker.
			return sql.NewInternalPlanner(opName, nil, user, &sql.MemoryMetrics{}, &execCfg)
		})

	// implements the server for the distributed SQL APIs.
	s.distSQLServer = distsqlrun.NewServer(ctx, distSQLCfg)
	 // allocates and returns a new REST server for administrative APIs.
	s.admin = newAdminServer(s, &sqlExecutor)
	// provides a RESTful status API. ???
	s.status = newStatusServer(
		// ...
	)
	// allocates and returns a new REST server for authentication APIs.
	s.authentication = newAuthenticationServer(s, &sqlExecutor)
	for _, gw := range []grpcGatewayServer{s.admin, s.status, s.authentication, &s.tsServer} {
		gw.RegisterService(s.grpc)
	}
	// initServer manages the temporary init server used during bootstrapping.
	s.initServer = newInitServer(s)
	// VirtualSchemaHolder is a type used to provide convenient access to virtual database and table descriptors
	virtualSchemas, err := sql.NewVirtualSchemaHolder(ctx, st)
	// Set up Executor sql语句的执行着
	s.sqlExecutor = sql.NewExecutor(execCfg, s.stopper)
	// 监听pg sql请求 ???
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
