
# summary

```
server/server.go#Start 
	-> sql/pgwire/server.go#ServerConn2 
	-> sql/pgwire/conn.go#serverConn 
	-> sql/pgwire/conn.go#serverImpl 
	-> sql/conn_executor.go#ServerConn
	-> sql/conn_executor.go#run
```

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
