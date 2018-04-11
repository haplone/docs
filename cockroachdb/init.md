
# 使用cobra来处理系统启动的命令行交互

* main.go中import 	"github.com/cockroachdb/cockroach/pkg/cli"
* 在cli/cli.go中使用init方法对md进行初始化

```go
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
```
* 在cli/start.go中有StartCmd的定义

```go
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
