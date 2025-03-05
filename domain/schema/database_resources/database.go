package database

import (
	"sqldb-ws/domain/schema/models"
	"strings"
)

/*
DB ROOT are all the ROOT database table needed in our generic API. They are restricted to modification
and can be impacted by a specialized service at DOMAIN level.
Their declarations is based on our Entity terminology, to help us in coding.
*/
// DBSchema express a table in the database, it's a template for a table
var DBSchema = models.SchemaModel{
	Name:     RootName("schema"),
	Label:    "template",
	Category: "template",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Level: models.LEVELRESPONSIBLE, Index: 0},
		{Name: models.LABELKEY, Type: models.BIGVARCHAR.String(), Required: true, Readonly: true, Index: 1},
		{Name: "category", Type: models.BIGVARCHAR.String(), Required: false, Default: "general", Readonly: true, Index: 2},
		{Name: "fields", Type: "onetomany", ForeignTable: RootName("schema_column"), Required: false, Index: 3},
	},
}

// DBSchemaField express a column in a table, it's a template for a column
var DBSchemaField = models.SchemaModel{
	Name:     RootName("schema_column"),
	Label:    "template field",
	Category: "template",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Required: true, Readonly: true, Index: 0},
		{Name: models.LABELKEY, Type: models.BIGVARCHAR.String(), Required: true, Index: 1},
		{Name: models.TYPEKEY, Type: models.DataTypeToEnum(), Required: true, Index: 2},
		{Name: "description", Type: models.TEXT.String(), Required: false, Index: 3},
		{Name: "placeholder", Type: models.VARCHAR.String(), Required: false, Index: 4},
		{Name: "default_value", Type: models.BIGVARCHAR.String(), Required: false, Index: 5, Label: "default"},
		{Name: "index", Type: models.INTEGER.String(), Required: true, Default: 1, Index: 6},
		{Name: "readonly", Type: models.BOOLEAN.String(), Required: true, Index: 7},
		{Name: "required", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 8},
		{Name: "read_level", Type: models.ENUMLEVEL.String(), Required: false, Default: models.LEVELNORMAL, Index: 9},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Index: 10, Label: "binded to template"},
		{Name: "constraints", Type: models.BIGVARCHAR.String(), Required: false, Level: models.LEVELRESPONSIBLE, Index: 11},
		{Name: "link_id", Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: false, Level: models.LEVELRESPONSIBLE, Index: 12, Label: "linked to"},
	},
}

// DBPermission express a permission in the database, ex: create, update, delete, read on a table
var DBPermission = models.SchemaModel{
	Name:     RootName("permission"),
	Label:    "permission",
	Category: "role & permission",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
		{Name: models.CREATEPERMS, Type: models.BOOLEAN.String(), Required: true, Index: 1},
		{Name: models.UPDATEPERMS, Type: models.BOOLEAN.String(), Required: true, Index: 2},
		{Name: models.DELETEPERMS, Type: models.BOOLEAN.String(), Required: true, Index: 3},
		{Name: models.READPERMS, Type: models.ENUMLEVELCOMPLETE.String(), Required: false, Default: models.LEVELNORMAL, Index: 4},
	},
}

// DBRole express a role in the database, ex: admin, user, guest with a set of permissions
var DBRole = models.SchemaModel{
	Name:     RootName("role"),
	Label:    "role",
	Category: "role & permission",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
		{Name: "description", Type: models.TEXT.String(), Required: false, Index: 1},
	},
}

// DBRolePermission express a role permission attribution in the database
var DBRolePermission = models.SchemaModel{
	Name:     RootName("role_permission"),
	Label:    "permission role attribution",
	Category: "role & permission",
	Fields: []models.FieldModel{
		{Name: RootID(DBRole.Name), Type: models.INTEGER.String(), ForeignTable: DBRole.Name, Required: true, Readonly: true, Index: 0, Label: "role"},
		{Name: RootID(DBPermission.Name), Type: models.INTEGER.String(), ForeignTable: DBPermission.Name, Required: true, Readonly: true, Index: 1, Label: "permission"},
	},
}

// DBEntity express an entity in the database, ex: user, task, project
var DBEntity = models.SchemaModel{
	Name:     RootName("entity"),
	Label:    "entity",
	Category: "entity",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Required: true, Readonly: true, Index: 0},
		{Name: "description", Type: models.TEXT.String(), Required: false, Index: 1},
		{Name: "parent_id", Type: models.INTEGER.String(), ForeignTable: RootName("entity"), Required: false, Index: 2, Label: "parent entity"},
	},
}

// DBUser express a user in the database, with email, password, token, super_admin
var DBUser = models.SchemaModel{
	Name:     RootName("user"),
	Label:    "user",
	Category: "user",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
		{Name: "email", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 1},
		{Name: "password", Type: models.TEXT.String(), Required: true, Level: models.LEVELRESPONSIBLE, Index: 2},
		{Name: "token", Type: models.TEXT.String(), Required: false, Default: "", Index: 3},
		{Name: "super_admin", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 4},
	},
}

// Note rules : HIERARCHY IS NOT INNER ROLE. HIERARCHY DEFINE MASTER OF AN ENTITY OR A USER. IT'S AN AUTO WATCHER ON USER ASSIGNEE TASK.
var DBHierarchy = models.SchemaModel{
	Name:     RootName("hierarchy"),
	Label:    "hierarchy",
	Category: "user",
	Fields: []models.FieldModel{
		{Name: "parent_" + RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: true, Index: 0, Label: "hierarchical user"},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Index: 1, Label: "user with hierarchy"},
		{Name: RootID(DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: DBEntity.Name, Required: false, Index: 2, Label: "entity with hierarchy"},
		{Name: models.STARTKEY, Type: models.TIMESTAMP.String(), Required: false, Default: "CURRENT_models.TIMESTAMP", Index: 3},
		{Name: models.ENDKEY, Type: models.TIMESTAMP.String(), Required: false, Index: 4},
	},
}

// DBEntityAttribution express an entity attribution in the database
var DBEntityUser = models.SchemaModel{
	Name:     RootName("entity_user"),
	Label:    "entity user attribution",
	Category: "entity",
	Fields: []models.FieldModel{
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: true, Readonly: true, Index: 0, Label: "user"},
		{Name: RootID(DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: DBEntity.Name, Required: true, Readonly: true, Index: 1, Label: "entity"},
		{Name: models.STARTKEY, Type: models.TIMESTAMP.String(), Required: false, Default: "CURRENT_models.TIMESTAMP", Index: 2},
		{Name: models.ENDKEY, Type: models.TIMESTAMP.String(), Required: false, Index: 3},
	},
}

// DBRoleAttribution express a role attribution in the database
var DBRoleAttribution = models.SchemaModel{
	Name:     RootName("role_attribution"),
	Label:    "role attribution",
	Category: "role & permission",
	Fields: []models.FieldModel{
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Readonly: true, Index: 0, Label: "user"},
		{Name: RootID(DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: DBEntity.Name, Required: false, Readonly: true, Index: 1, Label: "entity"},
		{Name: RootID(DBRole.Name), Type: models.INTEGER.String(), ForeignTable: DBRole.Name, Required: true, Readonly: true, Index: 2, Label: "role"},
		{Name: models.STARTKEY, Type: models.TIMESTAMP.String(), Required: false, Default: "CURRENT_models.TIMESTAMP", Index: 3},
		{Name: models.ENDKEY, Type: models.TIMESTAMP.String(), Required: false, Index: 4},
	},
}

// DBWorkflow express a workflow in the database, a workflow is a set of steps to achieve a request
var DBWorkflow = models.SchemaModel{
	Name:     RootName("workflow"),
	Label:    "workflow",
	Category: "workflow",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Constraint: "unique", Type: models.VARCHAR.String(), Required: true, Readonly: true, Index: 0},
		{Name: "description", Type: models.BIGVARCHAR.String(), Required: false, Index: 1},
		{Name: "is_meta", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 2, Label: "is a meta request"},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Label: "template entry", Index: 3},
		{Name: "steps", Type: "onetomany", ForeignTable: RootName("workflow_schema"), Required: false, Index: 4},
	},
}

// DBWorkflowSchema express a workflow schema in the database, a workflow schema is a step in a workflow
var DBWorkflowSchema = models.SchemaModel{
	Name:     RootName("workflow_schema"),
	Label:    "workflow schema attribution",
	Category: "workflow",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Required: true, Constraint: "unique", Readonly: true, Index: 0},
		{Name: "description", Type: models.TEXT.String(), Required: false, Index: 1},
		{Name: "index", Type: models.INTEGER.String(), Required: true, Default: 1, Index: 2},
		{Name: "urgency", Type: models.ENUMURGENCY.String(), Required: false, Default: models.LEVELNORMAL, Index: 3},
		{Name: "priority", Type: models.ENUMURGENCY.String(), Required: false, Default: models.LEVELNORMAL, Index: 4},
		{Name: "optionnal", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 5},
		{Name: "hub", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 6},
		{Name: RootID(DBWorkflow.Name), Type: models.INTEGER.String(), ForeignTable: DBWorkflow.Name, Required: true, Readonly: true, Label: "workflow attached", Index: 7},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Label: "template attached", Index: 8},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Label: "user assignee", Index: 9},
		{Name: RootID(DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: DBEntity.Name, Required: false, Label: "entity assignee", Index: 10},
		{Name: "wrapped_" + RootID(DBWorkflow.Name), Type: models.INTEGER.String(), ForeignTable: DBWorkflow.Name, Required: false, Readonly: true, Label: "wrapping workflow", Index: 11},
	},
}

// DBRequest express a request in the database, a request is a set of tasks to achieve a goal
var DBRequest = models.SchemaModel{
	Name:     RootName("request"),
	Label:    "request",
	Category: "request",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Required: true, Readonly: true, Index: 0},
		{Name: "state", Type: models.ENUMSTATE.String(), Required: false, Default: models.STATEPENDING, Level: models.LEVELRESPONSIBLE, Index: 1},
		{Name: "is_close", Type: models.BOOLEAN.String(), Required: false, Default: false, Level: models.LEVELRESPONSIBLE, Index: 2},
		{Name: "current_index", Type: models.INTEGER.String(), Required: false, Default: 0, Index: 3},
		{Name: "created_date", Type: models.TIMESTAMP.String(), Required: false, Default: "CURRENT_models.TIMESTAMP", Readonly: true, Level: models.LEVELRESPONSIBLE, Index: 4},
		{Name: "closing_date", Type: models.TIMESTAMP.String(), Required: false, Readonly: true, Level: models.LEVELRESPONSIBLE, Index: 5},
		{Name: RootID("dest_table"), Type: models.INTEGER.String(), Required: false, Readonly: true, Level: models.LEVELRESPONSIBLE, Label: "reference", Index: 6},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Level: models.LEVELRESPONSIBLE, Label: "template attached", Index: 7},
		{Name: RootID(DBWorkflow.Name), Type: models.INTEGER.String(), ForeignTable: DBWorkflow.Name, Required: true, Label: "request type", Index: 8},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Level: models.LEVELRESPONSIBLE, Label: "created by", Index: 9},
		{Name: "is_meta", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 10, Level: models.LEVELRESPONSIBLE},
	},
}

// DBTask express a task in the database, a task is an activity to achieve a step in a request
var DBTask = models.SchemaModel{
	Name:     RootName("task"),
	Label:    "activity",
	Category: "request",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Required: true, Readonly: true, Index: 0},
		{Name: "description", Type: models.BIGVARCHAR.String(), Required: false, Index: 1},
		{Name: "state", Type: models.ENUMSTATE.String(), Required: false, Default: models.STATEPENDING, Index: 2},
		{Name: "is_close", Type: models.BOOLEAN.String(), Required: false, Default: false, Level: models.LEVELRESPONSIBLE, Index: 3},
		{Name: "urgency", Type: models.ENUMURGENCY.String(), Required: false, Default: models.LEVELNORMAL, Readonly: true, Index: 4},
		{Name: "priority", Type: models.ENUMURGENCY.String(), Required: false, Default: models.LEVELNORMAL, Readonly: true, Index: 5},
		{Name: "comment", Type: models.TEXT.String(), Required: false, Index: 6},
		{Name: RootID(DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: DBEntity.Name, Required: false, Readonly: true, Label: "created by", Level: models.LEVELRESPONSIBLE, Index: 8},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Readonly: true, Label: "created by", Level: models.LEVELRESPONSIBLE, Index: 8},
		{Name: "created_date", Type: models.TIMESTAMP.String(), Required: false, Default: "CURRENT_models.TIMESTAMP", Readonly: true, Index: 9},
		{Name: "closing_date", Type: models.TIMESTAMP.String(), Required: false, Readonly: true, Index: 10},
		{Name: RootID("dest_table"), Type: models.INTEGER.String(), Required: false, Readonly: true, Label: "reference", Index: 11},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Label: "template attached", Index: 12},
		{Name: RootID(DBRequest.Name), Type: models.INTEGER.String(), ForeignTable: DBRequest.Name, Required: true, Readonly: true, Label: "request attached", Index: 13},
		{Name: RootID(DBWorkflowSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBWorkflowSchema.Name, Required: false, Readonly: true, Label: "workflow attached", Index: 14},
		{Name: "nexts", Type: models.BIGVARCHAR.String(), Required: false, Default: "all", Level: models.LEVELRESPONSIBLE, Index: 15},
		{Name: "meta_" + RootID(DBRequest.Name), Type: models.INTEGER.String(), ForeignTable: DBRequest.Name, Required: false, Readonly: true, Label: "meta request attached", Index: 16},
	},
}

// DBFilter express a filter in the database, a filter is a set of conditions to filter a view on a table
var DBFilter = models.SchemaModel{
	Name:     RootName("filter"),
	Label:    "filter",
	Category: "filter",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Required: true, Index: 0},
		{Name: "is_view", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 1},
		{Name: "is_selected", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 2},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: false, Index: 3},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Index: 4},
		{Name: RootID(DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: DBEntity.Name, Required: false, Index: 5},
		{Name: "elder", Type: models.ENUMLIFESTATE.String(), Required: false, Default: "all", Index: 6},
		{Name: "dashboard_restricted", Type: models.BOOLEAN.String(), Required: true, Default: false, Index: 7},
	},
}

// DBFilterField express a filter field in the database, a filter field is a condition to filter a view on a table
var DBFilterField = models.SchemaModel{
	Name:     RootName("filter_field"),
	Label:    "filter field",
	Category: "filter",
	Fields: []models.FieldModel{
		{Name: RootID(DBSchemaField.Name), Type: models.INTEGER.String(), ForeignTable: DBSchemaField.Name, Required: false, Index: 0},
		{Name: "value", Type: models.BIGVARCHAR.String(), Required: false, Index: 1},
		{Name: "operator", Type: models.ENUMOPERATOR.String(), Required: false, Index: 2},
		{Name: "separator", Type: models.ENUMSEPARATOR.String(), Required: false, Index: 3},
		{Name: "dir", Type: models.BIGVARCHAR.String(), Required: false, Index: 4},
		{Name: "index", Type: models.INTEGER.String(), Required: false, Default: 1, Index: 5},
		{Name: "width", Type: models.DECIMAL.String(), Required: false, Index: 6},
		{Name: RootID(DBFilter.Name), Type: models.INTEGER.String(), ForeignTable: DBFilter.Name, Required: false, Index: 7},
	},
}

// DBDashboardElement express a dashboard in the database, a dashboard is a set of views on a table
var DBDashboard = models.SchemaModel{
	Name:     RootName("dashboard"),
	Label:    "dashboard",
	Category: "filter",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Required: true, Index: 0},
		{Name: "description", Type: models.BIGVARCHAR.String(), Required: false, Index: 1},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Index: 2},
		{Name: RootID(DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: DBEntity.Name, Required: false, Index: 3},
		{Name: "is_selected", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 4},
	},
}

// DBDashboardElement express a dashboard element in the database, a dashboard element is a view on a table with a filter
var DBDashboardElement = models.SchemaModel{
	Name:     RootName("dashboard_element"),
	Label:    "dashboard element",
	Category: "filter",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Required: true, Index: 0},
		{Name: "description", Type: models.BIGVARCHAR.String(), Required: false, Index: 1},
		{Name: RootID(DBFilter.Name), Type: models.INTEGER.String(), ForeignTable: DBFilter.Name, Required: false, Index: 2},
		{Name: "order_by_" + RootID(DBSchemaField.Name), Type: models.INTEGER.String(), Required: false, Index: 3}, // results if multiple must be ordered by
		{Name: RootID(DBDashboard.Name), Type: models.INTEGER.String(), ForeignTable: DBDashboard.Name, Required: true, Index: 4},
	},
}

// DBDashboardMathField express a dashboard math field in the database, a dashboard math field is a math operation on a column
var DBDashboardMathField = models.SchemaModel{
	Name:     RootName("dashboard_math_field"),
	Label:    "dashboard math field",
	Category: "filter",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Required: true, Index: 0},
		{Name: RootID(DBDashboardElement.Name), Type: models.INTEGER.String(), ForeignTable: DBDashboardElement.Name, Required: true, Index: 1},
		{Name: "column_math_func", Type: models.ENUMMATHFUNC.String(), Required: false, Index: 2}, // func applied on operation added on column value ex: COUNT
		{Name: "row_math_func", Type: models.VARCHAR.String(), Required: true, Index: 3},          // operation applied on row ex: field + 3
	},
}

// DBView express a view in the database, a view is a set of fields to display on a table
var DBView = models.SchemaModel{
	Name:     RootName("view"),
	Label:    "view",
	Category: "view",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Required: true, Constraint: "unique", Index: 0},
		{Name: "description", Type: models.BIGVARCHAR.String(), Required: false, Index: 1},
		{Name: "category", Type: models.VARCHAR.String(), Required: false, Index: 2},
		{Name: "index", Type: models.INTEGER.String(), Required: false, Default: 1, Index: 3},
		{Name: "indexable", Type: models.BOOLEAN.String(), Required: false, Default: true, Index: 4},
		{Name: "is_list", Type: models.BOOLEAN.String(), Required: false, Default: true, Index: 5},
		{Name: "is_shortcut", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 6},
		{Name: "is_empty", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 7},
		{Name: "readonly", Type: models.BOOLEAN.String(), Required: true, Index: 8},
		{Name: "view_" + RootID(DBFilter.Name), Type: models.INTEGER.String(), ForeignTable: DBFilter.Name, Required: false, Index: 9},
		{Name: RootID(DBFilter.Name), Type: models.INTEGER.String(), ForeignTable: DBFilter.Name, Required: false, Index: 10},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Index: 11},
		{Name: "own_view", Type: models.BOOLEAN.String(), Required: false, Index: 12},
		{Name: "only_not_empty", Type: models.BOOLEAN.String(), Required: false, Index: 13},
	},
}

// DBViewAttribution express a view attribution in the database for a user or an entity
var DBViewAttribution = models.SchemaModel{
	Name:     RootName("view_attribution"),
	Label:    "view attribution",
	Category: "view",
	Fields: []models.FieldModel{
		{Name: RootID(DBView.Name), Type: models.INTEGER.String(), ForeignTable: DBView.Name, Required: true, Index: 0},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Index: 1},
		{Name: RootID(DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: DBEntity.Name, Required: false, Index: 2},
	},
}

// DBNotification express a notification in the database, a notification is a message to a user or an entity
var DBNotification = models.SchemaModel{
	Name:     RootName("notification"),
	Label:    "notification",
	Category: "notification",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Required: true, Index: 0},
		{Name: "description", Type: models.BIGVARCHAR.String(), Required: false, Index: 1},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Readonly: true, Label: "user assignee", Index: 2},
		{Name: RootID(DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: DBEntity.Name, Required: false, Readonly: true, Label: "entity assignee", Index: 3},
		{Name: RootID("dest_table"), Type: models.INTEGER.String(), Required: false, Readonly: true, Label: "reference", Index: 4}, // reference to a table if needed
		{Name: "link_id", Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Readonly: true, Label: "template attached", Index: 5},
	},
}

// DBDataAccess express a data access in the database, a data access is a log of access to a table
var DBDataAccess = models.SchemaModel{
	Name:     RootName("data_access"),
	Label:    "data access",
	Category: "history",
	Fields: []models.FieldModel{
		{Name: "update", Type: models.BOOLEAN.String(), Required: false, Default: false, Readonly: true, Label: "updated", Index: 0},
		{Name: "write", Type: models.BOOLEAN.String(), Required: false, Default: false, Readonly: true, Label: "created", Index: 1},
		{Name: "access_date", Type: models.TIMESTAMP.String(), Required: false, Default: "CURRENT_models.TIMESTAMP", Readonly: true, Label: "access date", Index: 2},
		{Name: RootID("dest_table"), Type: models.INTEGER.String(), Required: false, Readonly: true, Label: "reference", Index: 3},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Label: "template attached", Index: 4},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Readonly: true, Label: "related user", Index: 5},
	},
}

var OWNPERMISSIONEXCEPTION = []string{DBFilter.Name, DBFilterField.Name, DBNotification.Name,
	DBDashboard.Name, DBDashboardElement.Name, DBDashboardMathField.Name}
var AllPERMISSIONEXCEPTION = []string{DBNotification.Name, DBViewAttribution.Name}
var POSTPERMISSIONEXCEPTION = []string{DBRequest.Name}
var PUPERMISSIONEXCEPTION = []string{DBTask.Name}
var PERMISSIONEXCEPTION = []string{DBView.Name, DBTask.Name, DBRequest.Name, DBWorkflow.Name, DBEntity.Name, DBSchema.Name, DBSchemaField.Name} // override permission checkup

var ROOTTABLES = []models.SchemaModel{DBWorkflow, DBView, DBSchema, DBSchemaField, DBUser, DBPermission, DBEntity,
	DBRole, DBDataAccess, DBNotification, DBEntityUser, DBRoleAttribution,
	DBRequest, DBTask, DBWorkflowSchema, DBRolePermission, DBHierarchy, DBViewAttribution, DBFilter, DBFilterField,
	DBDashboard, DBDashboardElement, DBDashboardMathField,
}

func IsRootDB(name string) bool {
	if len(name) > 1 {
		return strings.Contains(name[:2], "db")
	} else {
		return false
	}
}
func RootID(name string) string {
	if IsRootDB(name) {
		return name + "_id"
	} else {
		return RootName(name) + "_id"
	}
}

func RootName(name string) string { return "db" + name }

var SchemaDBField = RootID(DBSchema.Name)
var SchemaFieldDBField = RootID(DBSchemaField.Name)
var RequestDBField = RootID(DBRequest.Name)
var WorkflowDBField = RootID(DBWorkflow.Name)
var WorkflowSchemaDBField = RootID(DBWorkflowSchema.Name)
var UserDBField = RootID(DBUser.Name)
var EntityDBField = RootID(DBEntity.Name)
var DestTableDBField = RootID("dest_table")
var FilterDBField = RootID(DBFilter.Name)
var ViewFilterDBField = "view_" + RootID(DBFilter.Name)
var ViewDBField = RootID(DBView.Name)
var DashboardDBField = RootID(DBDashboard.Name)
var DashboardElementDBField = RootID(DBDashboardElement.Name)
