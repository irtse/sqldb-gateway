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
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Constraint: "unique", Null : false, },
		 TableColumnEntity{ Name: TYPEATTR, Type: "varchar(255)", Null : true, },
		 TableColumnEntity{ Name: "title", Type: "varchar(255)", Null : true, Default : "unknown title" },
		 TableColumnEntity{ Name: "header", Type: "varchar(255)", Null : true, Default : "" },
	},
}

var DBSchemaField = TableEntity{
	Name : RootName("schema_column"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer", ForeignTable : DBSchema.Name, Null : false, },
		 TableColumnEntity{ Name: "required", Type: "boolean", Null : true, Default : false },
		 TableColumnEntity{ Name: "hidden", Type: "boolean", Null : true, Default : false },
		 TableColumnEntity{ Name: "readonly", Type: "boolean", Null : true, Default : false },
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Constraint: "unique", Null : false, },
		 TableColumnEntity{ Name: TYPEATTR, Type: "varchar(255)", Null : false, },
		 TableColumnEntity{ Name: "index", Type: "integer", Null : true, Default: 1 },
		 TableColumnEntity{ Name: "label", Type: "varchar(255)", Null : true, Default : "" },
		 TableColumnEntity{ Name: "placeholder", Type: "varchar(255)", Null : true, Default : "" },
		 TableColumnEntity{ Name: "default_value", Type: "varchar(255)", Null : true, },
		 TableColumnEntity{ Name: "description", Type: "varchar(255)", Null : true, Default : "no description..." },
        // link define a select.
		 TableColumnEntity{ Name: "link_sql_anchor", Type: "varchar(255)", Null : true, },
		 TableColumnEntity{ Name: "link_sql_order", Type: "varchar(255)", Null : true, },
		 TableColumnEntity{ Name: "link_sql_columns", Type: "varchar(255)", Null : true, },
		 TableColumnEntity{ Name: "link_sql_restriction", Type: "varchar(255)", Null : true, },
	},
}

var DBPermission = TableEntity{
	Name : RootName("permission"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Constraint: "unique", Null : false, },
		 TableColumnEntity{ Name: TABLENAMEATTR, Type: "varchar(255)", Null : false, },
		 TableColumnEntity{ Name: COLNAMEATTR, Type: "varchar(255)", Null : true, },
		 TableColumnEntity{ Name: CREATEPERMS, Type: "boolean", Null : true, Default : false },
		 TableColumnEntity{ Name: UPDATEPERMS, Type: "boolean", Null : true, Default : false },
		 TableColumnEntity{ Name: DELETEPERMS, Type: "boolean", Null : true, Default : false },
		 TableColumnEntity{ Name: READPERMS, Type: "boolean", Null : true, Default : false },
	},
}

var DBRole = TableEntity{
	Name : RootName("role"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Constraint: "unique", Null : false, },
		 TableColumnEntity{ Name: "description", Type: "text", Null : true, Default : "no description..." },
	},
}

var DBRolePermission = TableEntity{
	Name : RootName("role_permission"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: RootID(DBRole.Name), Type: "integer", ForeignTable: DBRole.Name, Null : false, },
		 TableColumnEntity{ Name: RootID(DBPermission.Name), Type: "integer", ForeignTable: DBPermission.Name, Null : false, },
	},
}

var DBEntity = TableEntity{
	Name : RootName("entity"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: TYPEATTR, Type: "varchar(255)", Null : true, },
		 TableColumnEntity{ Name: "parent_id", Type: "varchar(255)", Null : true, },
		 TableColumnEntity{ Name: "description", Type: "text", Null : true, },
	},
}

var DBUser = TableEntity{
	Name : RootName("user"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: "login", Type: "varchar(255)", Constraint: "unique", Null : false, },
		 TableColumnEntity{ Name: "password", Type: "varchar(255)", Null : false, },
		 TableColumnEntity{ Name: "token", Type: "varchar(255)", Null : true, Default : "" },
		 TableColumnEntity{ Name: "super_admin", Type: "boolean", Null : true, Default : false  },
	},
}
// Note rules : HIERARCHY IS NOT INNER ROLE. HIERARCHY DEFINE MASTER OF AN ENTITY OR A USER. IT'S AN AUTO WATCHER ON USER ASSIGNEE TASK.
var DBHierarchy = TableEntity{
	Name : RootName("hierarchy"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, Null : true, },
		 TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, Null : true, },
		 TableColumnEntity{ Name: "parent_" + RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, Null : false, },
		 TableColumnEntity{ Name: "start_date", Type: "timestamp",  Null : true, Default : "CURRENT_TIMESTAMP" },
		 TableColumnEntity{ Name: "end_date", Type: "timestamp",  Null : true, },
	},
}

var DBEntityUser = TableEntity{
	Name : RootName("entity_user"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, Null : true, },
		 TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, Null : true, },
		 TableColumnEntity{ Name: RootID(DBRole.Name), Type: "integer", ForeignTable: DBRole.Name, Null : false, },
		 TableColumnEntity{ Name: "start_date", Type: "timestamp",  Null : true, Default: "CURRENT_TIMESTAMP"},
		 TableColumnEntity{ Name: "end_date", Type: "timestamp",  Null : true, },
	},
}

var DBRoleAttribution = TableEntity{
	Name : RootName("role_attribution"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, Null : true, },
		 TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, Null : true, },
		 TableColumnEntity{ Name: RootID(DBRole.Name), Type: "integer", ForeignTable: DBRole.Name, Null : false, },
		 TableColumnEntity{ Name: "start_date", Type: "timestamp",  Null : true, Default: "CURRENT_TIMESTAMP" },
		 TableColumnEntity{ Name: "end_date", Type: "timestamp",  Null : true, },
	},
}

var DBWorkflow = TableEntity{
	Name : RootName("workflow"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)",  Null : false, },
		TableColumnEntity{ Name: "description", Type: "varchar(255)",  Null : true, },
	},
}
var DBWorkflowSchema = TableEntity{
	Name : RootName("workflow_schema"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBWorkflow.Name), Type: "integer", ForeignTable: DBWorkflow.Name, Null : false, },
		TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer", ForeignTable: DBSchema.Name, Null : false, },
		TableColumnEntity{ Name: "index", Type: "integer", Null : true, Default: 1 },
	},
}

var DBTask = TableEntity{
	Name : RootName("task"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer", ForeignTable: DBSchema.Name, Null : true, },
		TableColumnEntity{ Name: "opened_by", Type: "integer", ForeignTable: DBUser.Name, Null : true, },
		TableColumnEntity{ Name: "created_by", Type: "integer", ForeignTable: DBUser.Name, Null : true, },
		TableColumnEntity{ Name: "opened_date", Type: "timestamp",  Null : false, },
		TableColumnEntity{ Name: "created_date", Type: "timestamp",  Null : true, Default : "CURRENT_TIMESTAMP"},
		TableColumnEntity{ Name: "state", Type: "enum('completed', 'in progress', 'pending', 'close')",  Null : true, Default: "pending" },
		TableColumnEntity{ Name: "urgency", Type: "enum('low', 'medium', 'high')",  Null : true, Default: "medium" },
		TableColumnEntity{ Name: "priority", Type: "enum('low', 'medium', 'high')",  Null : true, Default: "medium" },
		TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)",  Null : false, },
		TableColumnEntity{ Name: "description", Type: "varchar(255)",  Null : true, Default: "no description..." },
		TableColumnEntity{ Name: "header", Type: "varchar(255)",  Null : true, },
		TableColumnEntity{ Name: RootID("dest_table"), Type: "integer", Null : true, },
		TableColumnEntity{ Name: RootID(DBWorkflow.Name), Type: "integer", ForeignTable: DBWorkflow.Name, Null : true, },
	},
}

var DBWorkflowTask = TableEntity{
	Name : RootName("workflow_schema_task"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBWorkflowSchema.Name), Type: "integer", ForeignTable: DBWorkflowSchema.Name, Null : false, },
		TableColumnEntity{ Name: RootID(DBTask.Name), Type: "integer", ForeignTable: DBTask.Name, Null : false, },
	},
}

var DBTaskAssignee = TableEntity{
	Name : RootName("task_assignee"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, Null : false, },
		TableColumnEntity{ Name: RootID(DBTask.Name), Type: "integer", ForeignTable: DBTask.Name, Null : false, },
		TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, Null : false, },
		TableColumnEntity{ Name: "state", Type: "enum('in progress', 'pending', 'completed')",  Null : true, Default: "pending"},
		TableColumnEntity{ Name: "hidden", Type: "boolean",  Null : true, Default: false, },
	},
}

var DBTaskVerifyer = TableEntity{
	Name : RootName("task_verifyer"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, Null : true, },
		TableColumnEntity{ Name: RootID(DBTask.Name), Type: "integer", ForeignTable: DBTask.Name, Null : false, },
		TableColumnEntity{ Name: "state", Type: "enum('pending', 'dismiss', 'complete')",  Null : true, Default: "pending"},
		TableColumnEntity{ Name: "hidden", Type: "boolean",  Null : true, Default: false, },
	},
}

var DBTaskWatcher = TableEntity{
	Name : RootName("task_watcher"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, Null : true, },
		TableColumnEntity{ Name: RootID(DBTask.Name), Type: "integer", ForeignTable: DBTask.Name, Null : false, },
		TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, Null : true, },
		TableColumnEntity{ Name: "hidden", Type: "boolean",  Null : true, Default: false, },
	},
}

var DBView = TableEntity{
	Name : RootName("view"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: "name", Type: "varchar(255)",  Null : false, },
		TableColumnEntity{ Name: "category", Type: "varchar(100)",  Null : true, },
		TableColumnEntity{ Name: "description", Type: "varchar(255)",  Null : true,  Default: "no description...", },
		TableColumnEntity{ Name: "readonly", Type: "boolean",  Null : true, Default: false, },
		TableColumnEntity{ Name: "is_empty", Type: "boolean", Null : true, Default: false },
		TableColumnEntity{ Name: "index", Type: "integer", Null : true, Default: 1 },
		TableColumnEntity{ Name: "sql_order", Type: "varchar(255)", Null : true, },
		TableColumnEntity{ Name: "sql_view", Type: "varchar(255)", Null : true, },
		TableColumnEntity{ Name: "sql_restriction", Type: "varchar(255)", Null : true, },
		TableColumnEntity{ Name: "through_perms", Type: "integer",  ForeignTable : DBSchema.Name, Null : true, },
		TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer",
		                   ForeignTable : DBSchema.Name, Null : false, },
	},
}

var DBAction = TableEntity{
	Name : RootName("action"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: "icon", Type: "varchar(100)",  Null : true, Default: "", },
		TableColumnEntity{ Name: "name", Type: "varchar(255)",  Null : false, },
		TableColumnEntity{ Name: "description", Type: "varchar(255)",  Null : true, Default: "no description...", },
		TableColumnEntity{ Name: "category", Type: "varchar(100)",  Null : true, },
		TableColumnEntity{ Name: "method", Type: "enum('POST', 'GET', 'PUT', 'DELETE')",  Null : true, Default : "GET" },
		TableColumnEntity{ Name: "extra_path", Type: "varchar(255)", Null : true, Default : ""  },
		TableColumnEntity{ Name: "from_schema", Type: "integer",  ForeignTable : DBSchema.Name, Null : false, },
		TableColumnEntity{ Name: "to_schema", Type: "integer",  ForeignTable : DBSchema.Name, Null : true, },
		TableColumnEntity{ Name: RootID(DBView.Name), Type: "integer",  ForeignTable : DBView.Name, Null : true, },
		TableColumnEntity{ Name: "type", Type: "enum('LINK_SELECT', 'BUTTON', 'LINK_ADD')", Null : true, Default : "BUTTON" },
		TableColumnEntity{ Name: "link", Type: "integer", ForeignTable : DBSchema.Name, Null : true, },
		TableColumnEntity{ Name: "link_sql_order", Type: "varchar(255)", Null : true, },
		TableColumnEntity{ Name: "link_sql_columns", Type: "varchar(255)", Null : true, },
		TableColumnEntity{ Name: "link_sql_restriction", Type: "varchar(255)", Null : true, },
	},
}

var DBViewAction = TableEntity{
	Name : RootName("view_action"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBView.Name), Type: "integer", ForeignTable : DBView.Name, Null : false, },
		TableColumnEntity{ Name: RootID(DBAction.Name), Type: "integer", ForeignTable : DBAction.Name, Null : false, },
	},
}
var DBRESTRICTED = []TableEntity{ DBPermission, DBSchema, DBSchemaField,  } // override permission checkup
var PERMISSIONEXCEPTION = []TableEntity{ DBView, DBAction, DBTask, } // override permission checkup
var ROOTTABLES = []TableEntity{ DBWorkflow, DBRole, DBRolePermission, DBEntity, DBUser, DBRoleAttribution,
	    DBEntityUser,  DBTask, DBWorkflowSchema, DBWorkflowTask, DBTaskAssignee, 
		DBTaskVerifyer, DBTaskWatcher, DBView, DBAction, DBViewAction }
