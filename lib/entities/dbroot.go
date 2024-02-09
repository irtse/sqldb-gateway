package entities

/*
	DB ROOT are all the ROOT database table needed in our generic API. They are restricted to modification
	and can be impacted by a specialized service at DOMAIN level. 
	Their declarations is based on our Entity terminology, to help us in coding. 
*/
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
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Constraint: "unique", Null : false, Readonly : true },
		 TableColumnEntity{ Name: "title", Type: "varchar(255)", Null : true, Default : "unknown title", Readonly : true },
		 TableColumnEntity{ Name: "header", Type: "varchar(255)", Null : true, Default : "", Readonly : true },
	},
}

var DBSchemaField = TableEntity{
	Name : RootName("schema_column"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer", ForeignTable : DBSchema.Name, Null : false },
		 TableColumnEntity{ Name: "required", Type: "boolean", Null : false },
		 TableColumnEntity{ Name: "hidden", Type: "boolean", Null : false },
		 TableColumnEntity{ Name: "readonly", Type: "boolean", Null : false },
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Null : false, Readonly : true },
		 TableColumnEntity{ Name: TYPEATTR, Type: "varchar(255)", Null : false, },
		 TableColumnEntity{ Name: "index", Type: "integer", Null : true, Default: 1 },
		 TableColumnEntity{ Name: "kind", Type: "enum('LINK_SELECT', 'INPUT', 'LINK_ADD')", Null : true, Default : "INPUT" },
		 TableColumnEntity{ Name: "label", Type: "varchar(255)", Null : false },
		 TableColumnEntity{ Name: "placeholder", Type: "varchar(255)", Null : true, Default : "" },
		 TableColumnEntity{ Name: "default_value", Type: "varchar(255)", Null : true, },
		 TableColumnEntity{ Name: "description", Type: "varchar(255)", Null : true, Default : "no description..." },
        // link define a select.
		 TableColumnEntity{ Name: "link", Type: "varchar(255)", Null : true, },
		 TableColumnEntity{ Name: "link_sql_dir", Type: "varchar(255)", Null : true, },
		 TableColumnEntity{ Name: "link_sql_order", Type: "varchar(255)", Null : true, },
		 TableColumnEntity{ Name: "link_sql_view", Type: "varchar(255)", Null : true, },
	},
}

var DBPermission = TableEntity{
	Name : RootName("permission"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Constraint: "unique", Null : false, Readonly : true },
		 TableColumnEntity{ Name: TABLENAMEATTR, Type: "varchar(255)", Null : false, Readonly : true },
		 TableColumnEntity{ Name: COLNAMEATTR, Type: "varchar(255)", Null : true, Readonly : true },
		 TableColumnEntity{ Name: CREATEPERMS, Type: "boolean", Null : false,  },
		 TableColumnEntity{ Name: UPDATEPERMS, Type: "boolean", Null : false,  },
		 TableColumnEntity{ Name: DELETEPERMS, Type: "boolean", Null : false },
		 TableColumnEntity{ Name: READPERMS, Type: "boolean", Null : false, },
		 TableColumnEntity{ Name: RootName("role_permission"), Type: "manytomany" },
	},
}

var DBRole = TableEntity{
	Name : RootName("role"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Constraint: "unique", Null : false, Readonly : true },
		 TableColumnEntity{ Name: "description", Type: "text", Null : true, Default : "no description..." },
		 TableColumnEntity{ Name: RootName("role_permission"), Type: "manytomany" },
		 TableColumnEntity{ Name: RootName("role_attribution"), Type: "manytomany" },
	},
}

var DBRolePermission = TableEntity{
	Name : RootName("role_permission"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: RootID(DBRole.Name), Type: "integer", ForeignTable: DBRole.Name, Null : false, Readonly : true },
		 TableColumnEntity{ Name: RootID(DBPermission.Name), Type: "integer", ForeignTable: DBPermission.Name, Null : false, Readonly : true },
	},
}

var DBEntity = TableEntity{
	Name : RootName("entity"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Null : true, Readonly : true },
		 TableColumnEntity{ Name: "parent_id", Type: "varchar(255)", Null : true, },
		 TableColumnEntity{ Name: "description", Type: "text", Null : true, },
		 TableColumnEntity{ Name: RootName("entity_user"), Type: "manytomany" },
		 TableColumnEntity{ Name: RootName("role_attribution"), Type: "manytomany" },
		 TableColumnEntity{ Name: RootName("hierarchy"), Type: "manytomany" },
	},
}

var DBUser = TableEntity{
	Name : RootName("user"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Constraint: "unique", Null : false, Readonly : true },
		 TableColumnEntity{ Name: "email", Type: "varchar(255)", Constraint: "unique", Null : false, Readonly : true },
		 TableColumnEntity{ Name: "password", Type: "varchar(255)", Null : false, Hidden: true },
		 TableColumnEntity{ Name: "token", Type: "varchar(255)", Null : true, Default : "" },
		 TableColumnEntity{ Name: "super_admin", Type: "boolean", Null : false,  },
		 TableColumnEntity{ Name: RootName("entity_user"), Type: "manytomany" },
		 TableColumnEntity{ Name: RootName("role_attribution"), Type: "manytomany" },
		 TableColumnEntity{ Name: RootName("hierarchy"), Type: "manytomany" },
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
		 TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, Null : true, Readonly : true },
		 TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, Null : true, Readonly : true },
		 TableColumnEntity{ Name: "start_date", Type: "timestamp",  Null : true, Default: "CURRENT_TIMESTAMP"},
		 TableColumnEntity{ Name: "end_date", Type: "timestamp",  Null : true, },
	},
}

var DBRoleAttribution = TableEntity{
	Name : RootName("role_attribution"),
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, Null : true, Readonly : true },
		 TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, Null : true, Readonly : true },
		 TableColumnEntity{ Name: RootID(DBRole.Name), Type: "integer", ForeignTable: DBRole.Name, Null : false, Readonly : true },
		 TableColumnEntity{ Name: "start_date", Type: "timestamp",  Null : true, Default: "CURRENT_TIMESTAMP" },
		 TableColumnEntity{ Name: "end_date", Type: "timestamp",  Null : true, },
	},
}

var DBWorkflow = TableEntity{
	Name : RootName("workflow"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: NAMEATTR, Constraint : "unique", Type: "varchar(255)",  Null : false, Readonly : true, },
		TableColumnEntity{ Name: "description", Type: "varchar(255)",  Null : true, },
		TableColumnEntity{ Name: RootName("workflow_schema"), Type: "manytomany" },
	},
}
var DBWorkflowSchema = TableEntity{
	Name : RootName("workflow_schema"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBWorkflow.Name), Type: "integer", ForeignTable: DBWorkflow.Name, Null : false, Readonly : true },
		TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer", ForeignTable: DBSchema.Name, Null : false, Readonly : true },
		TableColumnEntity{ Name: "index", Type: "integer", Null : true, Default: 1 },
	},
}

var DBTask = TableEntity{
	Name : RootName("task"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer", ForeignTable: DBSchema.Name, Null : false, Readonly : true },
		TableColumnEntity{ Name: RootID("opened_by"), Type: "integer", ForeignTable: DBUser.Name, Null : true, Readonly : true },
		TableColumnEntity{ Name: RootID("created_by"), Type: "integer", ForeignTable: DBUser.Name, Null : true, Readonly : true },
		TableColumnEntity{ Name: "opened_date", Type: "timestamp",  Null : true, Readonly : true },
		TableColumnEntity{ Name: "comment", Type: "text", Null : false, },
		TableColumnEntity{ Name: "created_date", Type: "timestamp",  Null : true, Default : "CURRENT_TIMESTAMP", Readonly : true },
		TableColumnEntity{ Name: "state", Type: "enum('completed', 'in progress', 'pending', 'close')",  Null : true, Default: "pending" },
		TableColumnEntity{ Name: "urgency", Type: "enum('low', 'medium', 'high')",  Null : true, Default: "medium", Readonly : true },
		TableColumnEntity{ Name: "priority", Type: "enum('low', 'medium', 'high')",  Null : true, Default: "medium" },
		TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)",  Null : false, Readonly : true },
		TableColumnEntity{ Name: "description", Type: "varchar(255)",  Null : true, Default: "no description..." },
		TableColumnEntity{ Name: "header", Type: "varchar(255)",  Null : true, Readonly : true },
		TableColumnEntity{ Name: RootID("dest_table"), Type: "integer", Null : true, Readonly : true },
		TableColumnEntity{ Name: RootID(DBWorkflow.Name), Type: "integer", ForeignTable: DBWorkflow.Name, Null : true, Readonly : true },
		TableColumnEntity{ Name: RootName("task_assignee"), Type: "manytomany" },
		TableColumnEntity{ Name: RootName("task_verifyer"), Type: "manytomany" },
		TableColumnEntity{ Name: RootName("task_watcher"), Type: "manytomany" },
	},
}

var DBTaskAssignee = TableEntity{
	Name : RootName("task_assignee"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, Null : true, Readonly : true },
		TableColumnEntity{ Name: RootID(DBTask.Name), Type: "integer", ForeignTable: DBTask.Name, Null : false, Readonly : true },
		TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, Null : true, Readonly : true },
		TableColumnEntity{ Name: "state", Type: "enum('in progress', 'pending', 'completed')",  Null : true, Default: "pending"},
		TableColumnEntity{ Name: "hidden", Type: "boolean",  Null : false, },
	},
}

var DBTaskVerifyer = TableEntity{
	Name : RootName("task_verifyer"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, Null : false, },
		TableColumnEntity{ Name: RootID(DBTask.Name), Type: "integer", ForeignTable: DBTask.Name, Null : false, },
		TableColumnEntity{ Name: "state", Type: "enum('pending', 'rejected', 'complete')",  Null : true, Default: "pending"},
		TableColumnEntity{ Name: "hidden", Type: "boolean",  Null : false, },
	},
}

var DBTaskWatcher = TableEntity{
	Name : RootName("task_watcher"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, Null : true, },
		TableColumnEntity{ Name: RootID(DBTask.Name), Type: "integer", ForeignTable: DBTask.Name, Null : false, },
		TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, Null : true, },
		TableColumnEntity{ Name: "hidden", Type: "boolean",  Null : false, },
	},
}

var DBView = TableEntity{
	Name : RootName("view"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)",  Null : false, },
		TableColumnEntity{ Name: "category", Type: "varchar(100)",  Null : true, },
		TableColumnEntity{ Name: "is_list", Type: "boolean", Null : false, Default: true },
		TableColumnEntity{ Name: "indexable", Type: "boolean", Null : false, Default: true },
		TableColumnEntity{ Name: "description", Type: "varchar(255)",  Null : true,  Default: "no description...", },
		TableColumnEntity{ Name: "is_empty", Type: "boolean", Null : false }, // EMPTY VIEW OR ...
		TableColumnEntity{ Name: "readonly", Type: "boolean", Null : false }, // SOLO VIEW OR ... 
		TableColumnEntity{ Name: "index", Type: "integer", Null : true, Default: 1 },
		TableColumnEntity{ Name: "sql_order", Type: "varchar(255)", Null : true, },
		TableColumnEntity{ Name: "sql_view", Type: "varchar(255)", Null : true, },
		TableColumnEntity{ Name: "sql_dir", Type: "varchar(255)", Null : true, },
		TableColumnEntity{ Name: "through_perms", Type: "integer",  ForeignTable : DBSchema.Name, Null : true, },
		TableColumnEntity{ Name: RootID("view"), Type: "integer", ForeignTable : RootName("view"), Null : true, },
		TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer", ForeignTable : DBSchema.Name, Null : false, },
		TableColumnEntity{ Name: RootName("view_action"), Type: "manytomany" },
	},
}

var DBAction = TableEntity{
	Name : RootName("action"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)",  Null : false, Readonly : true },
		TableColumnEntity{ Name: "description", Type: "varchar(255)",  Null : true, Default: "no description...", },
		TableColumnEntity{ Name: "parameters", Type: "varchar(255)",  Null : true },
		TableColumnEntity{ Name: "method", Type: "enum('POST', 'GET', 'PUT', 'DELETE')",  Null : true, Default : "GET" },
		TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer",  ForeignTable : DBSchema.Name, Null : false, },
		TableColumnEntity{ Name: "sql_dir", Type: "varchar(255)", Null : true, },
		TableColumnEntity{ Name: "sql_order", Type: "varchar(255)", Null : true, },
		TableColumnEntity{ Name: "sql_view", Type: "varchar(255)", Null : true, },
		TableColumnEntity{ Name: "extra_path", Type: "varchar(255)", Null : true, Default : ""  },
		TableColumnEntity{ Name: RootID("view"), Type: "integer", ForeignTable : DBView.Name, Null : true, },
		TableColumnEntity{ Name: RootName("view_action"), Type: "manytomany" },
	},
}

var DBViewAction = TableEntity{
	Name : RootName("view_action"),
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBView.Name), Type: "integer", ForeignTable : DBView.Name, Null : false, },
		TableColumnEntity{ Name: RootID(DBAction.Name), Type: "integer", ForeignTable : DBAction.Name, Null : false, },
	},
}

var DBRESTRICTED = []TableEntity{ DBSchema, DBSchemaField, } // override permission checkup
var PERMISSIONEXCEPTION = []TableEntity{ DBUser, DBPermission, DBEntity, DBRole, DBView, DBAction, 
										 DBEntityUser, DBRoleAttribution, } // override permission checkup
var ROOTTABLES = []TableEntity{ DBUser, DBPermission, DBEntity, DBRole, DBView, DBAction, 
	DBEntityUser, DBRoleAttribution, DBWorkflow, DBTask, DBWorkflowSchema, DBTaskAssignee, 
	DBTaskVerifyer, DBTaskWatcher,  DBViewAction, DBRolePermission, }
