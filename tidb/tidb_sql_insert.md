
sql :
```sql
INSERT INTO t VALUES ("pingcap001", "pingcap", 3);
```

ast:
```go
type InsertStmt struct {
	dmlNode

	IsReplace   bool
	IgnoreErr   bool
	Table       *TableRefsClause
	Columns     []*ColumnName
	Lists       [][]ExprNode
	Setlist     []*Assignment
	Priority    mysql.PriorityEnum
	OnDuplicate []*Assignment
	Select      ResultSetNode
}
```


*ast.InsertStmt
-- *ast.TableRefsClause
-- -- *ast.Join
-- -- -- *ast.TableSource
-- -- -- -- *ast.TableName
-- *ast.ValueExpr
-- *ast.ValueExpr
-- *ast.ValueExpr

validator

visitor
plan/preprocess.go#preprocessor

check privilege 

```go
// Manager is the interface for providing privilege related operations.
type Manager interface {
	// ShowGrants shows granted privileges for user.
	ShowGrants(ctx sessionctx.Context, user *auth.UserIdentity) ([]string, error)

	// RequestVerification verifies user privilege for the request.
	// If table is "", only check global/db scope privileges.
	// If table is not "", check global/db/table scope privileges.
	RequestVerification(db, table, column string, priv mysql.PrivilegeType) bool
	// ConnectionVerification verifies user privilege for connection.
	ConnectionVerification(host, user string, auth, salt []byte) bool

	// DBIsVisible returns true is the database is visible to current user.
	DBIsVisible(db string) bool

	// UserPrivilegesTable provide data for INFORMATION_SCHEMA.USERS_PRIVILEGE table.
	UserPrivilegesTable() [][]types.Datum
}

// RequestVerification checks whether the user have sufficient privileges to do the operation.
func (p *MySQLPrivilege) RequestVerification(user, host, db, table, column string, priv mysql.PrivilegeType) bool {
	log.Printf("check privilege by four level")
	record1 := p.matchUser(user, host)
	if record1 != nil && record1.Privileges&priv > 0 {
		return true
	}

	record2 := p.matchDB(user, host, db)
	if record2 != nil && record2.Privileges&priv > 0 {
		return true
	}

	record3 := p.matchTables(user, host, db, table)
	if record3 != nil {
		if record3.TablePriv&priv > 0 {
			return true
		}
		if column != "" && record3.ColumnPriv&priv > 0 {
			return true
		}
	}

	record4 := p.matchColumns(user, host, db, table, column)
	if record4 != nil && record4.ColumnPriv&priv > 0 {
		return true
	}

	return priv == 0
}
```

plan

```go
// Insert represents an insert plan.
type Insert struct {
	baseSchemaProducer

	Table       table.Table
	tableSchema *expression.Schema
	Columns     []*ast.ColumnName
	Lists       [][]expression.Expression
	Setlist     []*expression.Assignment
	OnDuplicate []*expression.Assignment

	IsReplace bool
	Priority  mysql.PriorityEnum
	IgnoreErr bool

	// NeedFillDefaultValue is true when expr in value list reference other column.
	NeedFillDefaultValue bool

	GenCols InsertGeneratedColumns

	SelectPlan PhysicalPlan
}
```

rbo 逻辑优化
cbo 物理优化

executor

builder

```go
// ExecStmt implements the ast.Statement interface, it builds a plan.Plan to an ast.Statement.
// code_analysis to_specify
type ExecStmt struct {
	// InfoSchema stores a reference to the schema information.
	InfoSchema infoschema.InfoSchema
	// Plan stores a reference to the final physical plan.
	Plan plan.Plan
	// Expensive represents whether this query is an expensive one.
	Expensive bool
	// Cacheable represents whether the physical plan can be cached.
	Cacheable bool
	// Text represents the origin query text.
	Text string

	StmtNode ast.StmtNode

	Ctx            sessionctx.Context
	startTime      time.Time
	isPreparedStmt bool
}
```

```go
// InsertExec represents an insert executor.
// code_analysis insert_concept
type InsertExec struct {
	*InsertValues

	OnDuplicate []*expression.Assignment

	Priority  mysql.PriorityEnum
	IgnoreErr bool

	finished bool
	rowCount int

	// For duplicate key update
	uniqueKeysInRows [][]keyWithDupError
	dupKeyValues     map[string][]byte
	dupOldRowValues  map[string][]byte
}
```

volcano 执行

