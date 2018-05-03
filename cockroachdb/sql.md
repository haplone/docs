
# summary

```
server/server.go#Start 
	-> sql/pgwire/server.go#ServerConn2 
	-> sql/pgwire/conn.go#serverConn 
	-> sql/pgwire/conn.go#serverImpl 
	-> sql/conn_executor.go#ServerConn
	-> sql/conn_executor.go#run
	-> sql/conn_executor_exec.go#execStmt
	-> sql/conn_executor_exec.go#execStmtInOpenState
	// planOptimize distSql
	-> sql/conn_executor_exec.go#dispatchToExecutionEngine
	-> sql/conn_executor_exec.go#execWithLocalEngine
	-> sql/plan.go#start
	-> sql/plan.go#startPlan
	-> sql/plan.go#startExec
	-> sql/walk.go#walkPlan
	-> sql/walk.go#visit
	-> sql/values.go#startExec(valueNode)
	

```

github.com/cockroachdb/cockroach/pkg/sql/distsqlrun.(*Flow).Start(0xc4209482c0, 0x2539ce0, 0xc425589440, 0x233c500, 0x251fe60, 0xc425576d80)
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/distsqlrun/flow.go:435 +0x34
github.com/cockroachdb/cockroach/pkg/sql.(*DistSQLPlanner).Run(0xc420564300, 0xc424d9f050, 0xc424d831e0, 0xc424d9f080, 0xc425576d80, 0xc424d7a450)
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/distsql_running.go:219 +0x978
github.com/cockroachdb/cockroach/pkg/sql.(*DistSQLPlanner).PlanAndRun(0xc420564300, 0x2539ce0, 0xc4254e7c80, 0xc424d831e0, 0x2530120, 0xc42542cb00, 0xc425576d80, 0xc424d7a450)
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/distsql_running.go:538 +0x255
github.com/cockroachdb/cockroach/pkg/sql.(*connExecutor).execWithDistSQLEngine(0xc424d7a000, 0x2539ce0, 0xc4254e7c80, 0xc424d7a3e0, 0x3, 0x7f3d31fbdd00, 0xc424d8d100, 0x1, 0x0)
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/conn_executor_exec.go:682 +0x1dc
github.com/cockroachdb/cockroach/pkg/sql.(*connExecutor).dispatchToExecutionEngine(0xc424d7a000, 0x2539ce0, 0xc4254e7c80, 0x253c4e0, 0xc425420e40, 0x0, 0x0, 0x0, 0x0, 0x0, ...)
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/conn_executor_exec.go:587 +0x6a8
github.com/cockroachdb/cockroach/pkg/sql.(*connExecutor).execStmtInOpenState(0xc424d7a000, 0x2539ce0, 0xc4254e7c80, 0x253c4e0, 0xc425420e40, 0x0, 0x0, 0x0, 0x0, 0x0, ...)
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/conn_executor_exec.go:351 +0xae9
github.com/cockroachdb/cockroach/pkg/sql.(*connExecutor).execStmt(0xc424d7a000, 0x2539ce0, 0xc4254e7c80, 0x253c4e0, 0xc425420e40, 0x0, 0x0, 0x0, 0x0, 0x0, ...)
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/conn_executor_exec.go:86 +0x56c
github.com/cockroachdb/cockroach/pkg/sql.(*connExecutor).run(0xc424d7a000, 0x2539c20, 0xc4251cc9c0, 0x0, 0x0)
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/conn_executor.go:914 +0x1e22
github.com/cockroachdb/cockroach/pkg/sql.(*Server).ServeConn(0xc4204a7420, 0x2539c20, 0xc4251cc9c0, 0x0, 0x0, 0xc424d6e009, 0x4, 0xc424d6e05c, 0xd, 0x2522f20, ...)
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/conn_executor.go:491 +0xf9e
github.com/cockroachdb/cockroach/pkg/sql/pgwire.(*conn).serveImpl.func3(0xc4204a7420, 0x2539c20, 0xc4251cc9c0, 0xc42513dc00, 0x5400, 0x15000, 0xc4205c3a50, 0xc424c409d0, 0xc4203defe0, 0xc424c409c0)
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/pgwire/conn.go:260 +0xfc
created by github.com/cockroachdb/cockroach/pkg/sql/pgwire.(*conn).serveImpl
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/pgwire/conn.go:259 +0xf14



github.com/cockroachdb/cockroach/pkg/sql/sqlbase.DecodeTableValue(0xc425576e18, 0x2547a40, 0x392b4c0, 0xc425566113, 0x4, 0x9, 0x20, 0x20, 0x212be60, 0x6a04f4, ...)
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/sqlbase/table.go:1352 +0x8a
github.com/cockroachdb/cockroach/pkg/sql/sqlbase.(*EncDatum).EnsureDecoded(0xc425584ab0, 0xc424d8d800, 0xc425576e18, 0xc4203f11e0, 0x2)
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/sqlbase/encoded_datum.go:203 +0xb2
github.com/cockroachdb/cockroach/pkg/sql.(*distSQLReceiver).Push(0xc425576d80, 0xc425584ab0, 0x2, 0x2, 0x0, 0xc4200fe7b8)
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/distsql_running.go:450 +0x940
github.com/cockroachdb/cockroach/pkg/sql/distsqlrun.Run(0x2539c20, 0xc42556a040, 0x253ab20, 0xc425584a80, 0x251fe60, 0xc425576d80)
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/distsqlrun/base.go:138 +0x5c
github.com/cockroachdb/cockroach/pkg/sql/distsqlrun.(*tableReader).Run(0xc425584a80, 0xc420948538)
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/distsqlrun/tablereader.go:132 +0x6b
created by github.com/cockroachdb/cockroach/pkg/sql/distsqlrun.(*Flow).Start
        /data/code/golang/src/github.com/cockroachdb/cockroach/pkg/sql/distsqlrun/flow.go:473 +0x3aa



// sql/executor_statement_metrics.go
const (
	// When the session is created (pgwire). Used to compute
	// the session age.
	sessionInit sessionPhase = iota

	// Executor phases.
	sessionQueryReceived    // Query is received.
	sessionStartParse       // Parse starts.
	sessionEndParse         // Parse ends.
	plannerStartLogicalPlan // Planning starts.
	plannerEndLogicalPlan   // Planning ends.
	plannerStartExecStmt    // Execution starts.
	plannerEndExecStmt      // Execution ends.

	// sessionNumPhases must be listed last so that it can be used to
	// define arrays sufficiently large to hold all the other values.
	sessionNumPhases
)


sql/distsqlrun/tablereader.go#Next
sql/sqlbase/rowfetcher.go#processKV

// sql/conn_executoer.go
 A connExecutor is in charge of executing queries received on a given client
 connection. The connExecutor implements a state machine (dictated by the
 Postgres/pgwire session semantics). The state machine is supposed to run
 asynchronously wrt the client connection: it receives input statements
 through a stmtBuf and produces results through a clientComm interface. The
 connExecutor maintains a cursor over the statementBuffer and executes
 statements / produces results for one statement at a time. The cursor points
 at all times to the statement that the connExecutor is currently executing.
 Results for statements before the cursor have already been produced (but not
 necessarily delivered to the client). Statements after the cursor are queued
 for future execution. Keeping already executed statements in the buffer is
 useful in case of automatic retries (in which case statements from the
 retried transaction have to be executed again); the connExecutor is in charge
 of removing old statements that are no longer needed for retries from the
 (head of the) buffer. Separately, the implementer of the clientComm interface
 (e.g. the pgwire module) is in charge of keeping track of what results have
 been delivered to the client and what results haven't (yet).

 The connExecutor has two main responsibilities: to dispatch queries to the
 execution engine(s) and relay their results to the clientComm, and to
 implement the state machine maintaining the various aspects of a connection's
 state. The state machine implementation is further divided into two aspects:
 maintaining the transaction status of the connection (outside of a txn,
 inside a txn, in an aborted txn, in a txn awaiting client restart, etc.) and
 maintaining the cursor position (i.e. correctly jumping to whatever the
 "next" statement to execute is in various situations).

 The cursor normally advances one statement at a time, but it can also skip
 some statements (remaining statements in a query string are skipped once an
 error is encountered) and it can sometimes be rewound when performing
 automatic retries. Rewinding can only be done if results for the rewound
 statements have not actually been delivered to the client; see below.

```
                                                   +---------------------+
                                                   |connExecutor         |
                                                   |                     |
                                                   +->execution+--------------+
                                                   ||  +                 |    |
                                                   ||  |fsm.Event        |    |
                                                   ||  |                 |    |
                                                   ||  v                 |    |
                                                   ||  fsm.Machine(TxnStateTransitions)
                                                   ||  +  +--------+     |    |
      +--------------------+                       ||  |  |txnState|     |    |
      |stmtBuf             |                       ||  |  +--------+     |    |
      |                    | statements are read   ||  |                 |    |
      | +-+-+ +-+-+ +-+-+  +------------------------+  |                 |    |
      | | | | | | | | | |  |                       |   |   +-------------+    |
  +---> +-+-+ +++-+ +-+-+  |                       |   |   |session data |    |
  |   |        ^           |                       |   |   +-------------+    |
  |   |        |   +-----------------------------------+                 |    |
  |   |        +   v       | cursor is advanced    |  advanceInfo        |    |
  |   |       cursor       |                       |                     |    |
  |   +--------------------+                       +---------------------+    |
  |                                                                           |
  |                                                                           |
  +-------------+                                                             |
                +--------+                                                    |
                | parser |                                                    |
                +--------+                                                    |
                |                                                             |
                |                                                             |
                |                                   +----------------+        |
        +-------+------+                            |execution engine<--------+
        | pgwire conn  |               +------------+(local/DistSQL) |
        |              |               |            +----------------+
        |   +----------+               |
        |   |clientComm<---------------+
        |   +----------+           results are produced
        |              |
        +-------^------+
                |
                |
        +-------+------+
        | SQL client   |
        +--------------+
```


 The connExecutor is disconnected from client communication (i.e. generally
 network communication - i.e. pgwire.conn); the module doing client
 communication is responsible for pushing statements into the buffer and for
 providing an implementation of the clientConn interface (and thus sending
 results to the client). The connExecutor does not control when
 results are delivered to the client, but still it does have some influence
 over that; this is because of the fact that the possibility of doing
 automatic retries goes away the moment results for the transaction in
 question are delivered to the client. The communication module has full
 freedom in sending results whenever it sees fit; however the connExecutor
 influences communication in the following ways:

 a) When deciding whether an automatic retry can be performed for a
 transaction, the connExecutor needs to:

   1) query the communication status to check that no results for the txn have
   been delivered to the client and, if this check passes:
   2) lock the communication so that no further results are delivered to the
   client, and, eventually:
   3) rewind the clientComm to a certain position corresponding to the start
   of the transaction, thereby discarding all the results that had been
   accumulated for the previous attempt to run the transaction in question.

 These steps are all orchestrated through clientComm.lockCommunication() and
 rewindCapability{}.

 b) The connExecutor sometimes ask the clientComm to deliver everything
 (most commonly in response to a Sync command).

 As of Feb 2018, the pgwire.conn delivers results synchronously to the client
 when its internal buffer overflows. In principle, delivery of result could be
 done asynchronously wrt the processing of commands (e.g. we could have a
 timing policy in addition to the buffer size). The first implementation of
 that showed a performance impact of involving a channel communication in the
 Sync processing path.


 Implementation notes:

 --- Error handling ---

 The key to understanding how the connExecutor handles errors is understanding
 the fact that there's two distinct categories of errors to speak of. There
 are "query execution errors" and there are the rest. Most things fall in the
 former category: invalid queries, queries that fail constraints at runtime,
 data unavailability errors, retriable errors (i.e. serializability
 violations) "internal errors" (e.g. connection problems in the cluster). This
 category of errors doesn't represent dramatic events as far as the connExecutor
 is concerned: they produce "results" for the query to be passed to the client
 just like more successful queries do and they produce Events for the
 state machine just like the successful queries (the events in question
 are generally event{non}RetriableErr and they generally cause the
 state machine to move to the Aborted state, but the connExecutor doesn't
 concern itself with this). The way the connExecutor reacts to these errors is
 the same as how it reacts to a successful query completing: it moves the
 cursor over the incoming statements as instructed by the state machine and
 continues running statements.

 And then there's other errors that don't have anything to do with a
 particular query, but with the connExecutor itself. In other languages, these
 would perhaps be modeled as Exceptions: we want them to unwind the stack
 significantly. These errors cause the connExecutor.run() to break out of its
 loop and return an error. Example of such errors include errors in
 communication with the client (e.g. the network connection is broken) or the
 connection's context being canceled.

 All of connExecutor's methods only return errors for the 2nd category. Query
 execution errors are written to a CommandResult. Low-level methods don't
 operate on a CommandResult directly; instead they operate on a wrapper
 (resultWithStoredErr), which provides access to the query error for purposes
 of building the correct state machine event.

 --- Context management ---

 At the highest level, there's connExecutor.run() that takes a context. That
 context is supposed to represent "the connection's context": its lifetime is
 the client connection's lifetime and it is assigned to
 connEx.ctxHolder.connCtx. Below that, every SQL transaction has its own
 derived context because that's the level at which we trace operations. The
 lifetime of SQL transactions is determined by the txnState: the state machine
 decides when transactions start and end in txnState.performStateTransition().
 When we're inside a SQL transaction, most operations are considered to happen
 in the context of that txn. When there's no SQL transaction (i.e.
 stateNoTxn), everything happens in the connection's context.

 High-level code in connExecutor is agnostic of whether it currently is inside
 a txn or not. To deal with both cases, such methods don't explicitly take a
 context; instead they use connEx.Ctx(), which returns the appropriate ctx
 based on the current state.
 Lower-level code (everything from connEx.execStmt() and below which runs in
 between state transitions) knows what state its running in, and so the usual
 pattern of explicitly taking a context as an argument is used.


# serverConn

// sql/pgwire/conn.go

 serveConn creates a conn that will serve the netConn. It returns once the
 network connection is closed.

 Internally, a connExecutor will be created to execute commands. Commands read
 from the network are buffered in a stmtBuf which is consumed by the
 connExecutor. The connExecutor produces results which are buffered and
 sometimes synchronously flushed to the network.

 The reader goroutine (this one) outlives the connExecutor's goroutine (the
 "processor goroutine").
 However, they can both signal each other to stop. Here's how the different
 cases work:
 1) The reader receives a ClientMsgTerminate protocol packet: the reader
 closes the stmtBuf and also cancels the command processing context. These
 actions will prompt the command processor to finish.
 2) The reader gets a read error from the network connection: like above, the
 reader closes the command processor.
 3) The reader's context is canceled (happens when the server is draining but
 the connection was busy and hasn't quit yet): the reader notices the canceled
 context and, like above, closes the processor.
 4) The processor encouters an error. This error can come from various fatal
 conditions encoutered internally by the processor, or from a network
 communication error encountered while flushing results to the network.
 The processor will cancel the reader's context and terminate.
 Note that query processing errors are different; they don't cause the
 termination of the connection.

 Draining notes:

 The reader notices that the server is draining by polling the draining()
 closure passed to serveConn. At that point, the reader delegates the
 responsibility of closing the connection to the statement processor: it will
 push a DrainRequest to the stmtBuf which signal the processor to quit ASAP.
 The processor will quit immediately upon seeing that command if it's not
 currently in a transaction. If it is in a transaction, it will wait until the
 first time a Sync command is processed outside of a transaction - the logic
 being that we want to stop when we're both outside transactions and outside
 batches.


# conn

// sql/pgwire/conn.go

 conn implements a pgwire network connection (version 3 of the protocol,
 implemented by Postgres v7.4 and later). 
 
 conn.serve() reads protocol messages, transforms them into commands that it pushes onto a `StmtBuf` (where
 they'll be picked up and executed by the `connExecutor`).
 The connExecutor produces results for the commands, which are delivered to
 the client through the sql.ClientComm interface, implemented by this conn
 (code is in command_result.go).


# StmtBuf

// sql/conn_io.go


 StmtBuf maintains a list of commands that a SQL client has sent for execution
 over a network connection. The commands are SQL queries to be executed,
 statements to be prepared, etc. At any point in time the buffer contains
 outstanding commands that have yet to be executed, and it can also contain
 some history of commands that we might want to retry - in the case of a
 retriable error, we'd like to retry all the commands pertaining to the
 current SQL transaction.

 The buffer is supposed to be used by one reader and one writer. The writer
 adds commands to the buffer using Push(). The reader reads one command at a
 time using curCmd(). The consumer is then supposed to create command results
 (the buffer is not involved in this).
 The buffer internally maintains a cursor representing the reader's position.
 The reader has to manually move the cursor using advanceOne(),
 seekToNextBatch() and rewind().
 In practice, the writer is a module responsible for communicating with a SQL
 client (i.e. pgwire.conn) and the reader is a connExecutor.

 The StmtBuf supports grouping commands into "batches" delimited by sync
 commands. A reader can then at any time chose to skip over commands from the
 current batch. This is used to implement Postgres error semantics: when an
 error happens during processing of a command, some future commands might need
 to be skipped. Batches correspond either to multiple queries received in a
 single query string (when the SQL client sends a semicolon-separated list of
 queries as part of the "simple" protocol), or to different commands pipelined
 by the cliend, separated from "sync" messages.

 push() can be called concurrently with curCmd().

 The connExecutor will use the buffer to maintain a window around the
 command it is currently executing. It will maintain enough history for
 executing commands again in case of an automatic retry. The connExecutor is
 in charge of trimming completed commands from the buffer when it's done with
 them.
