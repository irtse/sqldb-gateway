package entities

import ()

var TABLENAMEATTR = "table_name"
var COLNAMEATTR = "col_name"
var NAMEATTR = "name"
var TYPEATTR = "type"

var CREATEPERMS = "write"
var UPDATEPERMS = "update"
var DELETEPERMS = "delete"
var READPERMS = "read"

var DBSchema = TableEntity{
	Name : RootName("schema"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: TYPEATTR, Type: "varchar(255)", NotNull : false, },
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Constraint: "unique", NotNull : true, },
		 TableColumnEntity{ Name: "title", Type: "varchar(255)", NotNull : true, },
		 TableColumnEntity{ Name: "header", Type: "varchar(255)", NotNull : false, },
	},
}

var DBSchemaField = TableEntity{
	Name : RootName("schema_column"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer", ForeignTable : DBSchema.Name, NotNull : true, },
		 TableColumnEntity{ Name: "required", Type: "boolean", NotNull : true, },
		 TableColumnEntity{ Name: "readonly", Type: "boolean", NotNull : true, },
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Constraint: "unique", NotNull : true, },
		 TableColumnEntity{ Name: TYPEATTR, Type: "varchar(255)", NotNull : true, },
		 TableColumnEntity{ Name: "order", Type: "varchar(255)", NotNull : false, },
		 TableColumnEntity{ Name: "label", Type: "varchar(255)", NotNull : true, },
		 TableColumnEntity{ Name: "placeholder", Type: "varchar(255)", NotNull : false, },
		 TableColumnEntity{ Name: "default_value", Type: "varchar(255)", NotNull : false, },
		 TableColumnEntity{ Name: "description", Type: "varchar(255)", NotNull : false, },
        // link define a select.
		 TableColumnEntity{ Name: "link_anchor", Type: "varchar(255)", NotNull : false, },
		 TableColumnEntity{ Name: "link_order", Type: "varchar(255)", NotNull : false, },
		 TableColumnEntity{ Name: "link_columns", Type: "varchar(255)", NotNull : false, },
		 TableColumnEntity{ Name: "link_restriction", Type: "varchar(255)", NotNull : false, },
	},
}

var DBPermission = TableEntity{
	Name : RootName("permission"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Constraint: "unique", NotNull : true, },
		 TableColumnEntity{ Name: TABLENAMEATTR, Type: "varchar(255)", NotNull : true, },
		 TableColumnEntity{ Name: COLNAMEATTR, Type: "varchar(255)", NotNull : false, },
		 TableColumnEntity{ Name: CREATEPERMS, Type: "boolean", NotNull : true, },
		 TableColumnEntity{ Name: UPDATEPERMS, Type: "boolean", NotNull : true, },
		 TableColumnEntity{ Name: DELETEPERMS, Type: "boolean", NotNull : true, },
		 TableColumnEntity{ Name: READPERMS, Type: "boolean", NotNull : true, },
	},
}

var DBRole = TableEntity{
	Name : RootName("role"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Constraint: "unique", NotNull : true, },
		 TableColumnEntity{ Name: "description", Type: "text", NotNull : false, },
	},
}

var DBRolePermission = TableEntity{
	Name : RootName("role_permission"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: RootID(DBRole.Name), Type: "integer", ForeignTable: DBRole.Name, NotNull : true, },
		 TableColumnEntity{ Name: RootID(DBPermission.Name), Type: "integer", ForeignTable: DBPermission.Name, NotNull : true, },
	},
}

var DBEntity = TableEntity{
	Name : RootName("entity"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: TYPEATTR, Type: "varchar(255)", NotNull : false, },
		 TableColumnEntity{ Name: "parent_id", Type: "varchar(255)", ForeignTable: RootName("entity"), NotNull : false, },
		 TableColumnEntity{ Name: "description", Type: "text", NotNull : false, },
	},
}

var DBUser = TableEntity{
	Name : RootName("user"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: "login", Type: "varchar(255)", Constraint: "unique", NotNull : true, },
		 TableColumnEntity{ Name: "password", Type: "varchar(255)", NotNull : true, },
		 TableColumnEntity{ Name: "token", Type: "varchar(255)", NotNull : false, },
		 TableColumnEntity{ Name: "super_admin", Type: "boolean", NotNull : false, },
	},
}
// Note rules : HIERARCHY IS NOT INNER ROLE. HIERARCHY DEFINE MASTER OF AN ENTITY OR A USER. IT'S AN AUTO WATCHER ON USER ASSIGNEE TASK.
var DBHierarchy = TableEntity{
	Name : RootName("hierarchy"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, NotNull : false, },
		 TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, NotNull : false, },
		 TableColumnEntity{ Name: "parent_" + RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, NotNull : true, },
		 TableColumnEntity{ Name: "start_date", Type: "timestamp",  NotNull : false, },
		 TableColumnEntity{ Name: "end_date", Type: "timestamp",  NotNull : false, },
	},
}

var DBEntityUser = TableEntity{
	Name : RootName("entity_user"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, NotNull : false, },
		 TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, NotNull : false, },
		 TableColumnEntity{ Name: RootID(DBRole.Name), Type: "integer", ForeignTable: DBRole.Name, NotNull : true, },
		 TableColumnEntity{ Name: "start_date", Type: "timestamp",  NotNull : false, Default: "CURRENT_TIMESTAMP"},
		 TableColumnEntity{ Name: "end_date", Type: "timestamp",  NotNull : false, },
	},
}

var DBRoleAttribution = TableEntity{
	Name : RootName("role_attribution"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, NotNull : false, },
		 TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, NotNull : false, },
		 TableColumnEntity{ Name: RootID(DBRole.Name), Type: "integer", ForeignTable: DBRole.Name, NotNull : true, },
		 TableColumnEntity{ Name: "start_date", Type: "timestamp",  NotNull : false, },
		 TableColumnEntity{ Name: "end_date", Type: "timestamp",  NotNull : false, },
	},
}

var DBWorkflow = TableEntity{
	Name : RootName("workflow"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)",  NotNull : true, },
		TableColumnEntity{ Name: "description", Type: "varchar(255)",  NotNull : false, },
	},
}
var DBWorkflowSchema = TableEntity{
	Name : RootName("workflow_schema"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBWorkflow.Name), Type: "integer", ForeignTable: DBWorkflow.Name, NotNull : true, },
		TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer", ForeignTable: DBSchema.Name, NotNull : true, },
		TableColumnEntity{ Name: "order", Type: "integer", NotNull : true, },
	},
}

var DBTask = TableEntity{
	Name : RootName("task"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer", ForeignTable: DBSchema.Name, NotNull : false, },
		TableColumnEntity{ Name: "opened_by", Type: "integer", ForeignTable: DBUser.Name, NotNull : false, },
		TableColumnEntity{ Name: "created_by", Type: "integer", ForeignTable: DBUser.Name, NotNull : false, },
		TableColumnEntity{ Name: "opened_date", Type: "timestamp",  NotNull : true, },
		TableColumnEntity{ Name: "created_date", Type: "timestamp",  NotNull : true, Default : "CURRENT_TIMESTAMP"},
		TableColumnEntity{ Name: "state", Type: "enum('close', 'open', 'pending')",  NotNull : true, Default: "pending" },
		TableColumnEntity{ Name: "urgency", Type: "enum('low', 'medium', 'high')",  NotNull : true, Default: "medium" },
		TableColumnEntity{ Name: "priority", Type: "enum('low', 'medium', 'high')",  NotNull : true, Default: "medium" },
		TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)",  NotNull : true, },
		TableColumnEntity{ Name: "description", Type: "varchar(255)",  NotNull : false, },
		TableColumnEntity{ Name: "header", Type: "varchar(255)",  NotNull : false, },
		TableColumnEntity{ Name: RootID(DBWorkflow.Name), Type: "integer", ForeignTable: DBWorkflow.Name, NotNull : false, },
	},
}

var DBWorkflowTask = TableEntity{
	Name : RootName("workflow_schema_task"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBWorkflowSchema.Name), Type: "integer", ForeignTable: DBWorkflowSchema.Name, NotNull : true, },
		TableColumnEntity{ Name: RootID(DBTask.Name), Type: "integer", ForeignTable: DBTask.Name, NotNull : true, },
	},
}

var DBTaskAssignee = TableEntity{
	Name : RootName("task_assignee"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, NotNull : true, },
		TableColumnEntity{ Name: RootID(DBTask.Name), Type: "integer", ForeignTable: DBTask.Name, NotNull : true, },
		TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, NotNull : true, },
		TableColumnEntity{ Name: "state", Type: "enum('open', 'pending', 'complete')",  NotNull : true, Default: "pending"},
		TableColumnEntity{ Name: "hidden", Type: "boolean",  NotNull : true, Default: false, },
	},
}

var DBTaskVerifyer = TableEntity{
	Name : RootName("task_verifyer"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, NotNull : false, },
		TableColumnEntity{ Name: RootID(DBTask.Name), Type: "integer", ForeignTable: DBTask.Name, NotNull : true, },
		TableColumnEntity{ Name: "state", Type: "enum('pending', 'dismiss', 'complete')",  NotNull : true, Default: "pending"},
		TableColumnEntity{ Name: "hidden", Type: "boolean",  NotNull : true, Default: false, },
	},
}

var DBTaskWatcher = TableEntity{
	Name : RootName("task_watcher"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, NotNull : false, },
		TableColumnEntity{ Name: RootID(DBTask.Name), Type: "integer", ForeignTable: DBTask.Name, NotNull : true, },
		TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, NotNull : false, },
		TableColumnEntity{ Name: "hidden", Type: "boolean",  NotNull : true, Default: false, },
	},
}

var ROOTTABLES = []TableEntity{DBSchema, DBSchemaField, DBPermission, DBRole, DBRolePermission, DBEntity, DBUser, 
	    DBEntityUser, DBRoleAttribution, DBTask, DBWorkflow, DBWorkflowTask, DBTaskAssignee, DBTaskVerifyer, DBTaskWatcher }
