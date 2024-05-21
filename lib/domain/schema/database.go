package schema

/*
	DB ROOT are all the ROOT database table needed in our generic API. They are restricted to modification
	and can be impacted by a specialized service at DOMAIN level. 
	Their declarations is based on our Entity terminology, to help us in coding. 
*/
var DBSchema = SchemaModel{
	Name : RootName("schema"),
	Label : "template",
	Category : "template",
	Fields : []FieldModel{
		FieldModel{ Name: NAMEKEY, Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Level: LEVELRESPONSIBLE, Index : 0 },
		FieldModel{ Name: LABELKEY, Type: BIGVARCHAR.String(), Required : true, Readonly : true, Index : 1 },
		FieldModel{ Name: "category", Type: BIGVARCHAR.String(), Required : false, Default : "general", Readonly : true, Index : 2 },
		FieldModel{ Name: "fields", Type: "onetomany",  ForeignTable: RootName("schema_column"), Required : false, Index: 3 },
	},
}

var DBSchemaField = SchemaModel{
	Name : RootName("schema_column"),
	Label : "template field",
	Category : "template",
	Fields : []FieldModel{
		FieldModel{ Name: NAMEKEY, Type: VARCHAR.String(), Required : true, Readonly : true, Index: 0 },
		FieldModel{ Name: LABELKEY, Type: BIGVARCHAR.String(), Required : true, Index: 1 },
		FieldModel{ Name: TYPEKEY, Type: DataTypeToEnum(), Required : true, Index: 2 },
		FieldModel{ Name: "description", Type: TEXT.String(), Required : false, Index: 3  },
		FieldModel{ Name: "placeholder", Type: VARCHAR.String(), Required : false, Index: 4 },
		FieldModel{ Name: "default_value", Type: BIGVARCHAR.String(), Required : false, Index: 5, Label: "default", },
		FieldModel{ Name: "index", Type: INTEGER.String(), Required : true, Default: 1, Index: 6 },
		FieldModel{ Name: "readonly", Type: BOOLEAN.String(), Required : true, Index: 7 },
		FieldModel{ Name: "required", Type: BOOLEAN.String(), Required : false, Default: false, Index: 8 },
		FieldModel{ Name: "read_level", Type: ENUMLEVEL.String(), Required : false, Default: LEVELNORMAL, Index: 9 },
		FieldModel{ Name: RootID(DBSchema.Name), Type: INTEGER.String(), ForeignTable : DBSchema.Name, Required : true, Readonly: true, Index: 10, Label: "binded to template" },
		FieldModel{ Name: "constraints", Type: BIGVARCHAR.String(), Required : false, Level: LEVELRESPONSIBLE, Index: 11 },
		FieldModel{ Name: "link_id", Type: INTEGER.String(), ForeignTable: DBSchema.Name, Required : false, Level: LEVELRESPONSIBLE, Index: 12, Label: "linked to", },
	},
}

var DBPermission = SchemaModel{
	Name : RootName("permission"),
	Label : "permission",
	Category : "role & permission",
	Fields : []FieldModel{
		FieldModel{ Name: NAMEKEY, Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index: 0 },
		FieldModel{ Name: CREATEPERMS, Type: BOOLEAN.String(), Required : true, Index: 1  },
		FieldModel{ Name: UPDATEPERMS, Type: BOOLEAN.String(), Required : true, Index: 2  },
		FieldModel{ Name: DELETEPERMS, Type: BOOLEAN.String(), Required : true, Index: 3 },
		FieldModel{ Name: READPERMS, Type: ENUMLEVELCOMPLETE.String(), Required : false, Default: LEVELNORMAL, Index: 4 },
	},
}

var DBRole = SchemaModel{
	Name : RootName("role"),
	Label : "role",
	Category : "role & permission",
	Fields : []FieldModel{
		FieldModel{ Name: NAMEKEY, Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index: 0 },
		FieldModel{ Name: "description", Type: TEXT.String(), Required : false, Index: 1 },
	},
}

var DBRolePermission = SchemaModel{
	Name : RootName("role_permission"),
	Label : "permission role attribution",
	Category : "role & permission",
	Fields : []FieldModel{
		FieldModel{ Name: RootID(DBRole.Name), Type: INTEGER.String(), ForeignTable: DBRole.Name, Required : true, Readonly : true, Index: 0, Label: "role" },
		FieldModel{ Name: RootID(DBPermission.Name), Type: INTEGER.String(), ForeignTable: DBPermission.Name, Required : true, Readonly : true, Index: 1, Label: "permission" },
	},
}

var DBEntity = SchemaModel{
	Name : RootName("entity"),
	Label : "entity",
	Category : "entity",
	Fields : []FieldModel{
		FieldModel{ Name: NAMEKEY, Type: VARCHAR.String(), Required : true, Readonly : true, Index: 0 },
		FieldModel{ Name: "description", Type: TEXT.String(), Required : false, Index: 1 },
		FieldModel{ Name: "parent_id", Type: INTEGER.String(), ForeignTable: RootName("entity"), Required : false, Index: 2, Label: "parent entity", },
	},
}

var DBUser = SchemaModel{
	Name : RootName("user"),
	Label : "user",
	Category : "user",
	Fields : []FieldModel{
		FieldModel{ Name: NAMEKEY, Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index: 0  },
		FieldModel{ Name: "email", Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index: 1  },
		FieldModel{ Name: "password", Type: TEXT.String(), Required : true, Level: LEVELRESPONSIBLE, Index: 2 },
		FieldModel{ Name: "token", Type: TEXT.String(), Required : false, Default : "", Index: 3 },
		FieldModel{ Name: "super_admin", Type: BOOLEAN.String(), Required : false, Default : false, Index: 4 },
	},
}
// Note rules : HIERARCHY IS NOT INNER ROLE. HIERARCHY DEFINE MASTER OF AN ENTITY OR A USER. IT'S AN AUTO WATCHER ON USER ASSIGNEE TASK.
var DBHierarchy = SchemaModel{
	Name : RootName("hierarchy"),
	Label : "hierarchy",
	Category : "user",
	Fields : []FieldModel{
		FieldModel{ Name: "parent_" + RootID(DBUser.Name), Type: INTEGER.String(), ForeignTable: DBUser.Name, Required : true, Index: 0, Label: "hierarchical user" },
		FieldModel{ Name: RootID(DBUser.Name), Type: INTEGER.String(), ForeignTable: DBUser.Name, Required : false, Index: 1, Label: "user with hierarchy" },
		FieldModel{ Name: RootID(DBEntity.Name), Type: INTEGER.String(), ForeignTable: DBEntity.Name, Required : false, Index: 2, Label: "entity with hierarchy" },
		FieldModel{ Name: STARTKEY, Type: TIMESTAMP.String(),  Required : false, Default : "CURRENT_TIMESTAMP", Index: 3 },
		FieldModel{ Name: ENDKEY, Type: TIMESTAMP.String(),  Required : false, Index: 4 },
	},
}

var DBEntityUser = SchemaModel{
	Name : RootName("entity_user"),
	Label : "entity user attribution",
	Category : "entity",
	Fields : []FieldModel{
		 FieldModel{ Name: RootID(DBUser.Name), Type: INTEGER.String(), ForeignTable: DBUser.Name, Required : true, Readonly : true, Index: 0, Label: "user" },
		 FieldModel{ Name: RootID(DBEntity.Name), Type: INTEGER.String(), ForeignTable: DBEntity.Name, Required : true, Readonly : true, Index: 1, Label: "entity" },
		 FieldModel{ Name: STARTKEY, Type: TIMESTAMP.String(),  Required : false, Default: "CURRENT_TIMESTAMP", Index: 2 },
		 FieldModel{ Name: ENDKEY, Type: TIMESTAMP.String(),  Required : false, Index: 3 },
	},
}
var DBRoleAttribution = SchemaModel{
	Name : RootName("role_attribution"),
	Label : "role attribution",
	Category : "role & permission",
	Fields : []FieldModel{
		 FieldModel{ Name: RootID(DBUser.Name), Type:INTEGER.String(), ForeignTable: DBUser.Name, Required : false, Readonly : true, Index: 0, Label: "user" },
		 FieldModel{ Name: RootID(DBEntity.Name), Type: INTEGER.String(), ForeignTable: DBEntity.Name, Required : false, Readonly : true, Index: 1, Label: "entity" },
		 FieldModel{ Name: RootID(DBRole.Name), Type: INTEGER.String(), ForeignTable: DBRole.Name, Required : true, Readonly : true, Index: 2, Label: "role" },
		 FieldModel{ Name: STARTKEY, Type: TIMESTAMP.String(),  Required : false, Default: "CURRENT_TIMESTAMP", Index: 3, },
		 FieldModel{ Name: ENDKEY, Type: TIMESTAMP.String(),  Required : false, Index: 4 },
	},
}

var DBWorkflow = SchemaModel{
	Name : RootName("workflow"),
	Label : "workflow",
	Category : "workflow",
	Fields : []FieldModel{
		FieldModel{ Name: NAMEKEY, Constraint : "unique", Type: VARCHAR.String(),  Required : true, Readonly : true, Index: 0 },
		FieldModel{ Name: "description", Type: BIGVARCHAR.String(), Required : false, Index: 1 },
		FieldModel{ Name: "is_meta", Type: BOOLEAN.String(),  Required : false, Default: false, Index: 2, Label: "is a meta request", },
		FieldModel{ Name: RootID(DBSchema.Name), Type: INTEGER.String(), ForeignTable: DBSchema.Name, Required : true, Readonly : true, Label: "template entry", Index: 3 },
		FieldModel{ Name: "steps", Type: "onetomany",  ForeignTable: RootName("workflow_schema"), Required : false, Index: 4 },
	},
}

var DBWorkflowSchema = SchemaModel{
	Name : RootName("workflow_schema"),
	Label : "workflow schema attribution",
	Category : "workflow",
	Fields : []FieldModel{
		FieldModel{ Name: NAMEKEY, Type: VARCHAR.String(), Required : true, Constraint: "unique", Readonly: true, Index: 0 },
		FieldModel{ Name: "description", Type: TEXT.String(), Required : false, Index: 1 },
		FieldModel{ Name: "index", Type: INTEGER.String(), Required : true, Default: 1, Index: 2 },
		FieldModel{ Name: "urgency", Type: ENUMURGENCY.String(), Required : false, Default: LEVELNORMAL, Index: 3 },
		FieldModel{ Name: "priority", Type: ENUMURGENCY.String(), Required : false, Default: LEVELNORMAL, Index: 4 },
		FieldModel{ Name: "optionnal", Type: BOOLEAN.String(), Required : false, Default: false, Index: 5 },
		FieldModel{ Name: "hub", Type: BOOLEAN.String(), Required : false, Default: false, Index: 6 },
		FieldModel{ Name: RootID(DBWorkflow.Name), Type: INTEGER.String(), ForeignTable: DBWorkflow.Name, Required : true, Readonly : true, Label: "workflow attached", Index: 7 },
		FieldModel{ Name: RootID(DBSchema.Name), Type: INTEGER.String(), ForeignTable: DBSchema.Name, Required : true, Readonly : true, Label: "template attached", Index: 8 },
		FieldModel{ Name: RootID(DBUser.Name), Type: INTEGER.String(), ForeignTable: DBUser.Name, Required : false, Label: "user assignee", Index: 9 },
		FieldModel{ Name: RootID(DBEntity.Name), Type: INTEGER.String(), ForeignTable: DBEntity.Name, Required : false, Label: "entity assignee", Index: 10 },
		FieldModel{ Name: "wrapped_" + RootID(DBWorkflow.Name), Type: INTEGER.String(), ForeignTable: DBWorkflow.Name, Required : false, Readonly : true, Label: "wrapping workflow", Index: 11 },
	},
}

var DBRequest = SchemaModel{
	Name : RootName("request"),
	Label : "request",
	Category : "request",
	Fields : []FieldModel{
		FieldModel{ Name: NAMEKEY, Type: VARCHAR.String(), Required : true, Readonly : true, Index: 0 },
		FieldModel{ Name: "state", Type: ENUMSTATE.String(),  Required : false, Default: STATEPENDING, Level: LEVELRESPONSIBLE, Index: 1 },
		FieldModel{ Name: "is_close", Type: BOOLEAN.String(),  Required : false, Default: false, Level: LEVELRESPONSIBLE, Index: 2 },
		FieldModel{ Name: "current_index", Type: INTEGER.String(),  Required : false, Default: 0, Index: 3 },
		FieldModel{ Name: "created_date", Type: TIMESTAMP.String(),  Required : false, Default : "CURRENT_TIMESTAMP", Readonly : true, Level: LEVELRESPONSIBLE, Index: 4 },
		FieldModel{ Name: "closing_date", Type: TIMESTAMP.String(),  Required : false, Readonly : true, Level: LEVELRESPONSIBLE, Index: 5 },
		FieldModel{ Name: RootID("dest_table"), Type: INTEGER.String(), Required : false, Readonly : true, Level: LEVELRESPONSIBLE, Label: "reference", Index: 6 },
		FieldModel{ Name: RootID(DBSchema.Name), Type: INTEGER.String(), ForeignTable: DBSchema.Name, Required : true, Readonly : true, Level: LEVELRESPONSIBLE, Label: "template attached", Index: 7 },
		FieldModel{ Name: RootID(DBWorkflow.Name), Type: INTEGER.String(), ForeignTable: DBWorkflow.Name, Required : true, Label: "request type", Index: 8 },
		FieldModel{ Name: RootID(DBUser.Name), Type: INTEGER.String(), ForeignTable: DBUser.Name, Required : false, Level: LEVELRESPONSIBLE, Label: "created by", Index: 9 },
		FieldModel{ Name: "is_meta", Type: BOOLEAN.String(), Required : false, Default : false, Index: 10, Level: LEVELRESPONSIBLE, },
	},
}

var DBTask = SchemaModel{
	Name : RootName("task"),
	Label : "activity",
	Category : "request",
	Fields : []FieldModel{
		FieldModel{ Name: NAMEKEY, Type: VARCHAR.String(), Required : true, Readonly : true, Index: 0 },
		FieldModel{ Name: "description", Type: BIGVARCHAR.String(),  Required : false, Index: 1 },
		FieldModel{ Name: "state", Type: ENUMSTATE.String(),  Required : false, Default: STATEPENDING, Index: 2 },
		FieldModel{ Name: "is_close", Type: BOOLEAN.String(),  Required : false, Default: false, Level: LEVELRESPONSIBLE, Index: 3 },
		FieldModel{ Name: "urgency", Type: ENUMURGENCY.String(), Required : false, Default: LEVELNORMAL, Readonly : true, Index: 4 },
		FieldModel{ Name: "priority", Type: ENUMURGENCY.String(), Required : false, Default: LEVELNORMAL, Readonly : true, Index: 5 },
		FieldModel{ Name: "comment", Type: TEXT.String(), Required : false, Index: 6 },
		FieldModel{ Name: RootID(DBEntity.Name), Type: INTEGER.String(), ForeignTable: DBEntity.Name, Required : false, Readonly : true, Label: "created by", Level: LEVELRESPONSIBLE, Index: 8 },
		FieldModel{ Name: RootID(DBUser.Name), Type: INTEGER.String(), ForeignTable: DBUser.Name, Required : false, Readonly : true, Label: "created by", Level: LEVELRESPONSIBLE, Index: 8 },
		FieldModel{ Name: "created_date", Type: TIMESTAMP.String(),  Required : false, Default : "CURRENT_TIMESTAMP", Readonly : true, Index: 9 },
		FieldModel{ Name: "closing_date", Type: TIMESTAMP.String(),  Required : false, Readonly : true, Index: 10 },
		FieldModel{ Name: RootID("dest_table"), Type: INTEGER.String(), Required : false, Readonly : true, Label: "reference", Index: 11 },
		FieldModel{ Name: RootID(DBSchema.Name), Type: INTEGER.String(), ForeignTable: DBSchema.Name, Required : true, Readonly : true, Label: "template attached", Index: 12 },
		FieldModel{ Name: RootID(DBRequest.Name), Type: INTEGER.String(), ForeignTable: DBRequest.Name, Required : true, Readonly : true, Label: "request attached", Index: 13 },
		FieldModel{ Name: RootID(DBWorkflowSchema.Name), Type: INTEGER.String(),  ForeignTable: DBWorkflowSchema.Name, Required : false, Readonly : true, Label: "workflow attached", Index: 14 },
		FieldModel{ Name: "nexts", Type: BIGVARCHAR.String(),  Required : false, Default : "all", Level: LEVELRESPONSIBLE , Index: 15 },
		FieldModel{ Name: "meta_" + RootID(DBRequest.Name), Type: INTEGER.String(), ForeignTable: DBRequest.Name, Required : false, Readonly : true, Label: "meta request attached", Index: 16 },
	},
}

var DBFilter = SchemaModel{
	Name : RootName("filter"),
	Label : "filter",
	Category : "filter",
	Fields : []FieldModel{
		FieldModel{ Name: NAMEKEY, Type: VARCHAR.String(),  Required : true, Index: 0 },
		FieldModel{ Name: "is_view", Type: BOOLEAN.String(),  Required : false, Default: false, Index: 1 },
		FieldModel{ Name: "is_selected", Type: BOOLEAN.String(),  Required : false, Default: false, Index: 2 },
		FieldModel{ Name: RootID(DBSchema.Name), Type: INTEGER.String(), ForeignTable : DBSchema.Name, Required : false, Index: 3 },
		FieldModel{ Name: RootID(DBUser.Name), Type: INTEGER.String(), ForeignTable : DBUser.Name, Required : false, Index: 4 },
		FieldModel{ Name: RootID(DBEntity.Name), Type: INTEGER.String(), ForeignTable : DBEntity.Name, Required : false, Index: 5 },
	},
}

var DBFilterField = SchemaModel{
	Name : RootName("filter_field"),
	Label : "filter field",
	Category : "filter",
	Fields : []FieldModel{
		FieldModel{ Name: RootID(DBSchemaField.Name), Type: INTEGER.String(), ForeignTable: DBSchemaField.Name, Required : true, Index: 0 },
		FieldModel{ Name: "value", Type: BIGVARCHAR.String(), Required : false, Index: 1 },
		FieldModel{ Name: "operator", Type: ENUMOPERATOR.String(), Required : false, Index: 2 },
		FieldModel{ Name: "separator", Type: ENUMSEPARATOR.String(), Required : false, Index: 3 },
		FieldModel{ Name: "dir", Type: BIGVARCHAR.String(), Required : false, Index: 4 },
		FieldModel{ Name: "index", Type: INTEGER.String(), Required : false, Default: 1, Index: 5 },
		FieldModel{ Name: RootID(DBFilter.Name), Type: INTEGER.String(), ForeignTable: DBFilter.Name, Required : false, Index: 6 },
	},
}

var DBView = SchemaModel{
	Name : RootName("view"),
	Label : "view",
	Category : "view",
	Fields : []FieldModel{
		FieldModel{ Name: NAMEKEY, Type: VARCHAR.String(),  Required : true, Constraint: "unique", Index: 0 },
		FieldModel{ Name: "description", Type: BIGVARCHAR.String(),  Required : false, Index: 1 },
		FieldModel{ Name: "category", Type: VARCHAR.String(),  Required : false, Index: 2 },
		FieldModel{ Name: "index", Type: INTEGER.String(), Required : false, Default: 1, Index: 3 },
		FieldModel{ Name: "indexable", Type: BOOLEAN.String(), Required : false, Default: true, Index: 4 },
		FieldModel{ Name: "is_list", Type: BOOLEAN.String(), Required : false, Default: true, Index: 5 },
		FieldModel{ Name: "is_shortcut", Type: BOOLEAN.String(), Required : false, Default: false, Index: 6 },
		FieldModel{ Name: "is_empty", Type: BOOLEAN.String(), Required : false, Default: false, Index: 7 },
		FieldModel{ Name: "readonly", Type: BOOLEAN.String(), Required : true, Index: 8 },
		FieldModel{ Name: "view_" + RootID(DBFilter.Name), Type: INTEGER.String(), ForeignTable : DBFilter.Name, Required : false, Index: 9 },
		FieldModel{ Name: RootID(DBFilter.Name), Type: INTEGER.String(), ForeignTable : DBFilter.Name, Required : false, Index: 10 },
		FieldModel{ Name: RootID(DBSchema.Name), Type: INTEGER.String(), ForeignTable : DBSchema.Name, Required : true, Index: 11 },
	},
}

var DBViewAttribution = SchemaModel{
	Name : RootName("view_attribution"),
	Label : "view attribution",
	Category : "view",
	Fields : []FieldModel{
		FieldModel{ Name: RootID(DBView.Name), Type: INTEGER.String(), ForeignTable : DBView.Name, Required : true, Index: 0 },
		FieldModel{ Name: RootID(DBUser.Name), Type: INTEGER.String(), ForeignTable : DBUser.Name, Required : false, Index: 1 },
		FieldModel{ Name: RootID(DBEntity.Name), Type: INTEGER.String(), ForeignTable : DBEntity.Name, Required : false, Index: 2 },
	},
}

var DBNotification = SchemaModel{
	Name : RootName("notification"),
	Label : "notification",
	Category : "notification",
	Fields : []FieldModel{
		FieldModel{ Name: NAMEKEY, Type: VARCHAR.String(),  Required : true, Index: 0 },
		FieldModel{ Name: "description", Type: BIGVARCHAR.String(),  Required : false, Index: 1 },
		FieldModel{ Name: RootID(DBUser.Name), Type: INTEGER.String(), ForeignTable: DBUser.Name, Required : false, Readonly : true, Label: "user assignee", Index: 2 },
		FieldModel{ Name: RootID(DBEntity.Name), Type: INTEGER.String(), ForeignTable: DBEntity.Name, Required : false, Readonly : true, Label: "entity assignee", Index: 3 },
		FieldModel{ Name: RootID("dest_table"), Type: INTEGER.String(), Required : false, Readonly : true, Label: "reference", Index: 4 },
		FieldModel{ Name: "link_id", Type: INTEGER.String(), ForeignTable: DBSchema.Name, Readonly : true, Label: "template attached", Index: 5 },
	},
}

var DBDataAccess = SchemaModel{
	Name : RootName("data_access"),
	Label : "data access",
	Category : "history",
	Fields : []FieldModel{
		FieldModel{ Name: "update", Type: BOOLEAN.String(), Required : false, Default: false, Readonly : true, Label: "updated", Index: 0 },
		FieldModel{ Name: "write", Type: BOOLEAN.String(), Required : false, Default: false, Readonly : true, Label: "created", Index: 1 },
		FieldModel{ Name: "access_date", Type: TIMESTAMP.String(), Required : false, Default: "CURRENT_TIMESTAMP", Readonly : true, Label: "access date", Index: 2 },
		FieldModel{ Name: RootID("dest_table"), Type: INTEGER.String(), Required : false, Readonly : true, Label: "reference", Index: 3 },
		FieldModel{ Name: RootID(DBSchema.Name), Type: INTEGER.String(), ForeignTable: DBSchema.Name, Required : true, Readonly : true, Label: "template attached", Index: 4 },
		FieldModel{ Name: RootID(DBUser.Name), Type: INTEGER.String(), ForeignTable: DBUser.Name, Required : false, Readonly : true, Label: "related user", Index: 5 },
	},
}

var OWNPERMISSIONEXCEPTION = []string{ DBFilter.Name, DBFilterField.Name, DBNotification.Name }
var AllPERMISSIONEXCEPTION = []string{ DBNotification.Name, DBViewAttribution.Name }
var POSTPERMISSIONEXCEPTION = []string{ DBRequest.Name }
var PUPERMISSIONEXCEPTION = []string{ DBTask.Name }
var PERMISSIONEXCEPTION = []string{ DBView.Name, DBTask.Name, DBRequest.Name, DBWorkflow.Name, DBEntity.Name, DBSchema.Name, DBSchemaField.Name } // override permission checkup

var ROOTTABLES = []SchemaModel{ DBWorkflow, DBView, DBSchema, DBSchemaField, DBUser, DBPermission, DBEntity, 
	DBRole, DBDataAccess, DBNotification, DBEntityUser, DBRoleAttribution,
	DBRequest, DBTask, DBWorkflowSchema, DBRolePermission, DBHierarchy, DBViewAttribution, DBFilter, DBFilterField }
