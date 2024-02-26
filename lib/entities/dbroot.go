package entities

/*
	DB ROOT are all the ROOT database table needed in our generic API. They are restricted to modification
	and can be impacted by a specialized service at DOMAIN level. 
	Their declarations is based on our Entity terminology, to help us in coding. 
*/
var NAMEATTR = "name"
var TYPEATTR = "type"

var CREATEPERMS = "write"
var UPDATEPERMS = "update"
var DELETEPERMS = "delete"
var READPERMS = "read"

var LEVELADMIN = "admin"
var LEVELMODERATOR = "moderator"
var LEVELRESPONSIBLE = "responsible"
var LEVELNORMAL = "normal"
var READLEVELACCESS = []string{ LEVELNORMAL, LEVELRESPONSIBLE, LEVELMODERATOR, LEVELADMIN, }

var DBSchema = TableEntity{
	Name : RootName("schema"),
	Label : "form",
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Constraint: "unique", Null : false, Readonly : true, Level: LEVELRESPONSIBLE },
		 TableColumnEntity{ Name: "label", Type: "varchar(255)", Null : true, Default : "general", Readonly : true },
	},
}

var DBSchemaField = TableEntity{
	Name : RootName("schema_column"),
	Label : "field",
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer", ForeignTable : DBSchema.Name, Null : false, Readonly: true, },
		 TableColumnEntity{ Name: "required", Type: "boolean", Null : true, Default: false, },
		 TableColumnEntity{ Name: "read_level", Type: "enum('"+ LEVELADMIN + "', '"+ LEVELMODERATOR + "', '"+ LEVELRESPONSIBLE + "', '"+ LEVELNORMAL + "')", Null : true, Default: "'" + LEVELNORMAL +"'" },
		 TableColumnEntity{ Name: "readonly", Type: "boolean", Null : false, },
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Null : false, Readonly : true },
		 TableColumnEntity{ Name: TYPEATTR, Type: "varchar(255)", Null : false, },
		 TableColumnEntity{ Name: "index", Type: "integer", Null : true, Default: 1 },
		 TableColumnEntity{ Name: "label", Type: "varchar(255)", Null : false },
		 TableColumnEntity{ Name: "placeholder", Type: "varchar(255)", Null : true, Default : "" },
		 TableColumnEntity{ Name: "default_value", Type: "varchar(255)", Null : true, },
		 TableColumnEntity{ Name: "description", Type: "varchar(255)", Null : true, Default : "no description..." },
        // link define a select.
		 TableColumnEntity{ Name: "constraint", Type: "varchar(255)", Null : true, Level: LEVELRESPONSIBLE, },
		 TableColumnEntity{ Name: "link", Type: "varchar(255)", Null : true, Level: LEVELRESPONSIBLE, },
		 TableColumnEntity{ Name: "link_sql_dir", Type: "varchar(255)", Null : true, Level: LEVELRESPONSIBLE, },
		 TableColumnEntity{ Name: "link_sql_order", Type: "varchar(255)", Null : true, Level: LEVELRESPONSIBLE, },
		 TableColumnEntity{ Name: "link_sql_view", Type: "varchar(255)", Null : true, Level: LEVELRESPONSIBLE, },
	},
}

var DBPermission = TableEntity{
	Name : RootName("permission"),
	Label : "permission",
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Constraint: "unique", Null : false, Readonly : true },
		 TableColumnEntity{ Name: CREATEPERMS, Type: "boolean", Null : false,  },
		 TableColumnEntity{ Name: UPDATEPERMS, Type: "boolean", Null : false,  },
		 TableColumnEntity{ Name: DELETEPERMS, Type: "boolean", Null : false },
		 TableColumnEntity{ Name: READPERMS, Type: "enum('"+ LEVELADMIN + "', '"+ LEVELMODERATOR + "', '"+ LEVELRESPONSIBLE + "', '"+ LEVELNORMAL + "')", Null : false, Default: "normal" },
	},
}

var DBRole = TableEntity{
	Name : RootName("role"),
	Label : "role",
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Constraint: "unique", Null : false, Readonly : true },
		 TableColumnEntity{ Name: "description", Type: "text", Null : true, Default : "no description..." },
	},
}

var DBRolePermission = TableEntity{
	Name : RootName("role_permission"),
	Label : "permission role attribution",
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: RootID(DBRole.Name), Type: "integer", ForeignTable: DBRole.Name, Null : false, Readonly : true },
		 TableColumnEntity{ Name: RootID(DBPermission.Name), Type: "integer", ForeignTable: DBPermission.Name, Null : false, Readonly : true },
	},
}

var DBEntity = TableEntity{
	Name : RootName("entity"),
	Label : "entity",
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Null : true, Readonly : true },
		 TableColumnEntity{ Name: "parent_id", Type: "varchar(255)", Null : true, },
		 TableColumnEntity{ Name: "description", Type: "text", Null : true, },
	},
}

var DBUser = TableEntity{
	Name : RootName("user"),
	Label : "user",
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Constraint: "unique", Null : false, Readonly : true },
		 TableColumnEntity{ Name: "email", Type: "varchar(255)", Constraint: "unique", Null : false, Readonly : true },
		 TableColumnEntity{ Name: "password", Type: "varchar(255)", Null : false, Level: LEVELRESPONSIBLE },
		 TableColumnEntity{ Name: "token", Type: "varchar(255)", Null : true, Default : "" },
		 TableColumnEntity{ Name: "super_admin", Type: "boolean", Null : false,  },
	},
}
// Note rules : HIERARCHY IS NOT INNER ROLE. HIERARCHY DEFINE MASTER OF AN ENTITY OR A USER. IT'S AN AUTO WATCHER ON USER ASSIGNEE TASK.
var DBHierarchy = TableEntity{
	Name : RootName("hierarchy"),
	Label : "hierarchy",
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
	Label : "entity user attribution",
	Columns : []TableColumnEntity{
		 TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, Null : true, Readonly : true },
		 TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, Null : true, Readonly : true },
		 TableColumnEntity{ Name: "start_date", Type: "timestamp",  Null : true, Default: "CURRENT_TIMESTAMP"},
		 TableColumnEntity{ Name: "end_date", Type: "timestamp",  Null : true, },
	},
}

var DBRoleAttribution = TableEntity{
	Name : RootName("role_attribution"),
	Label : "role attribution",
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
	Label : "workflow",
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: NAMEATTR, Constraint : "unique", Type: "varchar(255)",  Null : false, Readonly : true, },
		TableColumnEntity{ Name: "description", Type: "varchar(255)",  Null : true, },
		TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer", ForeignTable: DBSchema.Name, Null : false, Readonly : true, Label: "form entry", },
	},
}
var DBWorkflowSchema = TableEntity{
	Name : RootName("workflow_schema"),
	Label : "workflow schema attribution",
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBWorkflow.Name), Type: "integer", ForeignTable: DBWorkflow.Name, Null : false, Readonly : true, Label: "workflow attached", },
		TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer", ForeignTable: DBSchema.Name, Null : false, Readonly : true, Label: "form attached", },
		TableColumnEntity{ Name: "index", Type: "integer", Null : true, Default: 1 },
		TableColumnEntity{ Name: "description", Type: "text", Null : true, },
		TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)", Null : false, Constraint: "UNIQUE", Readonly: true, },
		TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, Null : true, Label: "user assignee", },
		TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, Null : true, Label: "entity assignee", },
		TableColumnEntity{ Name: "urgency", Type: "enum('low', 'medium', 'high')",  Null : true, Default: "medium",},
		TableColumnEntity{ Name: "priority", Type: "enum('low', 'medium', 'high')",  Null : true, Default: "medium", },
	},
}

var DBRequest = TableEntity{
	Name : RootName("request"),
	Label : "request",
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)",  Null : true, Readonly : true,},
		TableColumnEntity{ Name: "state", Type: "enum('pending', 'progressing', 'rejected', 'completed')",  Null : true, Default: "pending", Level: LEVELRESPONSIBLE},
		TableColumnEntity{ Name: "is_close", Type: "boolean",  Null : true, Default: false, Level: LEVELRESPONSIBLE },
		TableColumnEntity{ Name: "current_index", Type: "integer",  Null : true, Default: 0, Level: LEVELRESPONSIBLE },
		TableColumnEntity{ Name: RootID(DBWorkflow.Name), Type: "integer", ForeignTable: DBWorkflow.Name, Null : false, Label: "workflow attached", },
		TableColumnEntity{ Name: RootID("dest_table"), Type: "integer", Null : true, Readonly : true, Level: LEVELRESPONSIBLE, Label: "reference", },
		TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer", ForeignTable: DBSchema.Name, Null : false, Readonly : true, Level: LEVELRESPONSIBLE, Label: "form attached", },
		TableColumnEntity{ Name: RootID("created_by"), Type: "integer", ForeignTable: DBUser.Name, Null : true, Readonly : true, Level: LEVELRESPONSIBLE },
		TableColumnEntity{ Name: "created_date", Type: "timestamp",  Null : true, Default : "CURRENT_TIMESTAMP", Readonly : true, Level: LEVELRESPONSIBLE },
	},
}

var DBTask = TableEntity{
	Name : RootName("task"),
	Label : "activity",
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer", ForeignTable: DBSchema.Name, Null : false, Readonly : true, Label: "form attached", },
		TableColumnEntity{ Name: RootID(DBRequest.Name), Type: "integer", ForeignTable: DBRequest.Name, Null : false, Readonly : true, Label: "request attached", },
		TableColumnEntity{ Name: RootID(DBUser.Name), Type: "integer", ForeignTable: DBUser.Name, Null : true, Readonly : true, Label: "user assignee", },
		TableColumnEntity{ Name: RootID(DBEntity.Name), Type: "integer", ForeignTable: DBEntity.Name, Null : true, Readonly : true, Label: "entity assignee", },
		TableColumnEntity{ Name: RootID("created_by"), Type: "integer", ForeignTable: DBUser.Name, Null : true, Readonly : true, Level: LEVELRESPONSIBLE },
		TableColumnEntity{ Name: "comment", Type: "text", Null : true, Default : "", Readonly : true, },
		TableColumnEntity{ Name: "created_date", Type: "timestamp",  Null : true, Default : "CURRENT_TIMESTAMP", Readonly : true },
		TableColumnEntity{ Name: "state", Type: "enum('pending', 'progressing', 'dismiss', 'completed')",  Null : true, Default: "pending" },
		TableColumnEntity{ Name: "is_close", Type: "boolean",  Null : true, Default: false, Level: LEVELRESPONSIBLE },
		TableColumnEntity{ Name: "urgency", Type: "enum('low', 'medium', 'high')",  Null : true, Default: "medium", Readonly : true, },
		TableColumnEntity{ Name: "priority", Type: "enum('low', 'medium', 'high')",  Null : true, Default: "medium", Readonly : true, },
		TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)",  Null : false, Readonly : true },
		TableColumnEntity{ Name: "description", Type: "varchar(255)",  Null : true, Default: "no description...", Readonly : true, },
		TableColumnEntity{ Name: RootID(DBWorkflowSchema.Name), Type: "integer",  ForeignTable: DBWorkflowSchema.Name, Null : true, Readonly : true, Level: LEVELRESPONSIBLE, Label: "workflow attached", },
		TableColumnEntity{ Name: RootID("dest_table"), Type: "integer", Null : true, Readonly : true, Label: "reference", },
	},
}

var DBView = TableEntity{
	Name : RootName("view"),
	Label : "view",
	Columns : []TableColumnEntity{
		TableColumnEntity{ Name: NAMEATTR, Type: "varchar(255)",  Null : false, },
		TableColumnEntity{ Name: "is_list", Type: "boolean", Null : false, Default: true },
		TableColumnEntity{ Name: "can_pull_schema", Type: "boolean", Null : false, Default: false },
		TableColumnEntity{ Name: "is_empty", Type: "boolean", Null : false, Default: false },
		TableColumnEntity{ Name: "indexable", Type: "boolean", Null : false, Default: true },
		TableColumnEntity{ Name: "description", Type: "varchar(255)",  Null : true,  Default: "no description...", },
		TableColumnEntity{ Name: "readonly", Type: "boolean", Null : false }, // SOLO VIEW OR ... 
		TableColumnEntity{ Name: "index", Type: "integer", Null : true, Default: 1 },
		TableColumnEntity{ Name: "sql_restriction", Type: "varchar(255)", Null : true, },
		TableColumnEntity{ Name: "sql_order", Type: "varchar(255)", Null : true, },
		TableColumnEntity{ Name: "sql_view", Type: "varchar(255)", Null : true, },
		TableColumnEntity{ Name: "sql_dir", Type: "varchar(255)", Null : true, },
		TableColumnEntity{ Name: "through_perms", Type: "integer",  ForeignTable : DBSchema.Name, Null : true, },
		TableColumnEntity{ Name: RootID("view"), Type: "integer", ForeignTable : RootName("view"), Null : true, },
		TableColumnEntity{ Name: RootID(DBSchema.Name), Type: "integer", ForeignTable : DBSchema.Name, Null : false, },
	},
}
var POSTPERMISSIONEXCEPTION = []TableEntity{ DBRequest }
var PUPERMISSIONEXCEPTION = []TableEntity{ DBTask }
var PERMISSIONEXCEPTION = []TableEntity{ DBView, DBTask, DBRequest, DBWorkflow } // override permission checkup
var ROOTTABLES = []TableEntity{ DBSchema, DBSchemaField, DBUser, DBPermission, DBEntity, DBRole, DBView, 
	DBEntityUser, DBRoleAttribution, DBWorkflow, DBRequest, DBTask, DBWorkflowSchema, DBRolePermission, DBHierarchy, }
