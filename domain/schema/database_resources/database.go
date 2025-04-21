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
		{Name: "can_owned", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 4},
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
		{Name: models.TYPEKEY, Type: models.VARCHAR.String(), Required: true, Index: 2},
		{Name: "description", Type: models.TEXT.String(), Required: false, Index: 3},
		{Name: "placeholder", Type: models.VARCHAR.String(), Required: false, Index: 4},
		{Name: "default_value", Type: models.BIGVARCHAR.String(), Required: false, Index: 5, Label: "default"},
		{Name: "index", Type: models.INTEGER.String(), Required: true, Default: 1, Index: 6},
		{Name: "readonly", Type: models.BOOLEAN.String(), Required: true, Index: 7},
		{Name: "required", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 8},
		{Name: "read_level", Type: models.ENUMLEVEL.String(), Required: false, Default: models.LEVELNORMAL, Index: 9},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Index: 10, Label: "binded to template"},
		{Name: "constraints", Type: models.BIGVARCHAR.String(), Required: false, Level: models.LEVELRESPONSIBLE, Index: 11},
		{Name: "link_id", Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: false, Index: 12, Label: "linked to"},
		{Name: "hidden", Type: models.BOOLEAN.String(), Default: false, Required: false, Index: 13, Label: "is hidden"},
		{Name: "translatable", Type: models.BOOLEAN.String(), Default: true, Required: false, Index: 14, Label: "is translatable"},
		{Name: "transform_function", Type: models.ENUMTRANSFORM.String(), Required: false, Index: 15, Label: "transformation function"},
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
	Category: "",
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
	CanOwned: true,
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
		{Name: "password", Type: models.TEXT.String(), Required: false, Default: "", Level: models.LEVELRESPONSIBLE, Index: 2},
		{Name: "token", Type: models.TEXT.String(), Required: false, Default: "", Level: models.LEVELRESPONSIBLE, Index: 3},
		{Name: "super_admin", Type: models.BOOLEAN.String(), Required: false, Default: false, Level: models.LEVELRESPONSIBLE, Index: 4},
	},
}

var DBEmailTemplate = models.SchemaModel{
	Name:     RootName("email_template"),
	Label:    "email template",
	Category: "email",
	Fields: []models.FieldModel{
		{Name: "subject", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 1},
		{Name: "template", Type: models.BIGINT.String(), Required: true, Index: 2},
		{Name: "waiting_response", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 3},
		{Name: "to_map_" + RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Label: "template attached", Index: 4},
	},
}

var DBEmailSended = models.SchemaModel{
	Name:     RootName("email_sended"),
	Label:    "email sended",
	Category: "email",
	Fields: []models.FieldModel{
		{Name: "from", Type: models.VARCHAR.String(), Required: true, Readonly: true, Index: 0},
		{Name: "to", Type: models.VARCHAR.String(), Required: true, Readonly: true, Index: 1},
		{Name: "subject", Type: models.VARCHAR.String(), Required: true, Readonly: true, Index: 2},
		{Name: "content", Type: models.BIGINT.String(), Required: true, Index: 3},
		{Name: RootID(DBEmailTemplate.Name), Type: models.INTEGER.String(), ForeignTable: DBEmailTemplate.Name, Required: true, Readonly: true, Label: "email attached", Index: 4},
		{Name: "code", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Index: 5},
		{Name: "mapped_with" + RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Label: "template attached", Index: 4},
		{Name: "mapped_with" + RootID("dest_table"), Type: models.INTEGER.String(), Required: true, Readonly: true, Label: "template attached", Index: 5},
	},
}

var DBEmailResponse = models.SchemaModel{
	Name:     RootName("email_response"),
	Label:    "email response",
	Category: "email",
	Fields: []models.FieldModel{
		{Name: "got_response", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 0},
		{Name: "comment", Type: models.VARCHAR.String(), Required: false, Index: 1},
		{Name: RootID(DBEmailSended.Name), Type: models.INTEGER.String(), ForeignTable: DBEmailSended.Name, Required: true, Readonly: true, Label: "email attached", Index: 2},
	},
}

var DBTrigger = models.SchemaModel{
	Name:     RootName("trigger"),
	Label:    "trigger",
	Category: "trigger",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
		{Name: "type", Type: models.ENUMTRIGGER.String(), Required: true, Readonly: true, Index: 1},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Label: "template attached", Index: 3},
		{Name: "on_write", Type: models.BOOLEAN.String(), Required: true, Readonly: false, Default: false, Label: "on creation", Index: 2},
		{Name: "on_update", Type: models.BOOLEAN.String(), Required: true, Readonly: false, Default: false, Label: "on update", Index: 3},
	},
}

var DBTriggerCondition = models.SchemaModel{
	Name:     RootName("trigger_condition"),
	Label:    "trigger condition",
	Category: "trigger",
	Fields: []models.FieldModel{
		{Name: "value", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 0},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Label: "template to check condition", Index: 5},
		{Name: RootID(DBSchemaField.Name), Type: models.INTEGER.String(), ForeignTable: DBSchemaField.Name, Required: true, Readonly: true, Label: "field to check condition", Index: 6},
		{Name: RootID(DBTrigger.Name), Type: models.INTEGER.String(), ForeignTable: DBTrigger.Name, Required: true, Readonly: true, Label: "related trigger", Index: 6},
	},
}

var DBTriggerRule = models.SchemaModel{
	Name:     RootName("trigger_rule"),
	Label:    "trigger rule",
	Category: "trigger",
	Fields: []models.FieldModel{
		{Name: "value", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 0},

		{Name: "from_" + RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: false, Readonly: true, Label: "template to extract value modification", Index: 2},
		{Name: "from_" + RootID(DBSchemaField.Name), Type: models.INTEGER.String(), ForeignTable: DBSchemaField.Name, Required: false, Readonly: true, Label: "field  to extract value modification", Index: 3},

		{Name: "to_" + RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Label: "template to apply modification", Index: 5},
		{Name: "to_" + RootID(DBSchemaField.Name), Type: models.INTEGER.String(), ForeignTable: DBSchemaField.Name, Required: true, Readonly: true, Label: "field to apply modification", Index: 6},
		{Name: RootID(DBTrigger.Name), Type: models.INTEGER.String(), ForeignTable: DBTrigger.Name, Required: true, Readonly: true, Label: "related trigger", Index: 6},
	},
}

var DBFieldAutoFill = models.SchemaModel{
	Name:     RootName("field_autofill"),
	Label:    "field autofill",
	Category: "schema",
	Fields: []models.FieldModel{
		{Name: "value", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 0},
		{Name: "first_own", Type: models.BOOLEAN.String(), Label: "first of our data", Required: false, Readonly: false, Default: false, Index: 1},

		{Name: "from_" + RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: false, Readonly: true, Label: "template to extract value modification", Index: 2},
		{Name: "from_" + RootID(DBSchemaField.Name), Type: models.INTEGER.String(), ForeignTable: DBSchemaField.Name, Required: false, Readonly: true, Label: "field  to extract value modification", Index: 3},
		{Name: "from_" + RootID("dest_table"), Type: models.INTEGER.String(), Required: false, Readonly: true, Label: "reference", Index: 4},

		{Name: RootID(DBSchemaField.Name), Type: models.INTEGER.String(), ForeignTable: DBSchemaField.Name, Required: true, Readonly: true, Label: "field to check condition", Index: 5},
	},
}

// Note rules : HIERARCHY IS NOT INNER ROLE. HIERARCHY DEFINE MASTER OF AN ENTITY OR A USER. IT'S AN AUTO WATCHER ON USER ASSIGNEE TASK.
var DBHierarchy = models.SchemaModel{
	Name:     RootName("hierarchy"),
	Label:    "hierarchy",
	Category: "user",
	CanOwned: true,
	Fields: []models.FieldModel{
		{Name: "parent_" + RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: true, Index: 0, Label: "hierarchical user"},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Index: 1, Label: "user with hierarchy"},
		{Name: RootID(DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: DBEntity.Name, Required: false, Index: 2, Label: "entity with hierarchy"},
		{Name: models.STARTKEY, Type: models.TIMESTAMP.String(), Required: false, Default: "CURRENT_TIMESTAMP", Index: 3},
		{Name: models.ENDKEY, Type: models.TIMESTAMP.String(), Required: false, Index: 4},
	},
}

// DBEntityAttribution express an entity attribution in the database
var DBEntityUser = models.SchemaModel{
	Name:     RootName("entity_user"),
	Label:    "entity user attribution",
	Category: "",
	Fields: []models.FieldModel{
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: true, Readonly: true, Index: 0, Label: "user"},
		{Name: RootID(DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: DBEntity.Name, Required: true, Readonly: true, Index: 1, Label: "entity"},
		{Name: models.STARTKEY, Type: models.TIMESTAMP.String(), Required: false, Default: "CURRENT_TIMESTAMP", Index: 2},
		{Name: models.ENDKEY, Type: models.TIMESTAMP.String(), Required: false, Index: 3},
	},
}

// DBRoleAttribution express a role attribution in the database
var DBRoleAttribution = models.SchemaModel{
	Name:     RootName("role_attribution"),
	Label:    "role attribution",
	Category: "",
	Fields: []models.FieldModel{
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Readonly: true, Index: 0, Label: "user"},
		{Name: RootID(DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: DBEntity.Name, Required: false, Readonly: true, Index: 1, Label: "entity"},
		{Name: RootID(DBRole.Name), Type: models.INTEGER.String(), ForeignTable: DBRole.Name, Required: true, Readonly: true, Index: 2, Label: "role"},
		{Name: models.STARTKEY, Type: models.TIMESTAMP.String(), Required: false, Default: "CURRENT_TIMESTAMP", Index: 3},
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
	Category: "",
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
		{Name: "before_hierarchical_validation", Type: models.BOOLEAN.String(), Required: false, Readonly: false, Label: "must have a before hierarchical validation", Index: 12},
		{Name: "custom_progressing_status", Type: models.VARCHAR.String(), Required: false, Readonly: true, Label: "rename of the pending status", Index: 13},
		{Name: "view_" + RootID(DBFilter.Name), Type: models.INTEGER.String(), ForeignTable: DBFilter.Name, Required: false, Label: "filter to apply on step", Index: 10, Hidden: true},
		{Name: "readonly_not_assignee", Type: models.BOOLEAN.String(), Required: false, Default: false, Label: "readonly for not assignee", Index: 11, Hidden: true},
	},
}

// TODO RELATE FILTER TO TASK IF ONE

// DBRequest express a request in the database, a request is a set of tasks to achieve a goal
var DBRequest = models.SchemaModel{
	Name:     RootName("request"),
	Label:    "request",
	Category: "request",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Required: true, Readonly: true, Index: 0},
		{Name: "state", Type: models.VARCHAR.String(), Required: false, Default: models.STATEPENDING, Level: models.LEVELRESPONSIBLE, Index: 1},
		{Name: "is_close", Type: models.BOOLEAN.String(), Required: false, Default: false, Level: models.LEVELRESPONSIBLE, Index: 2},
		{Name: "current_index", Type: models.FLOAT8.String(), Required: false, Default: 0, Index: 3},
		{Name: "closing_date", Type: models.TIMESTAMP.String(), Required: false, Readonly: true, Level: models.LEVELRESPONSIBLE, Index: 5},
		{Name: RootID("dest_table"), Type: models.INTEGER.String(), Required: false, Readonly: true, Label: "reference", Index: 6},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Label: "template attached", Index: 7},
		{Name: RootID(DBWorkflow.Name), Type: models.INTEGER.String(), ForeignTable: DBWorkflow.Name, Required: true, Label: "request type", Index: 8},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Label: "created by", Index: 9},
		{Name: "is_meta", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 10, Hidden: true},
	},
}

// DBWorkflow express a workflow in the database, a workflow is a set of steps to achieve a request
var DBConsent = models.SchemaModel{
	Name:     RootName("consent"),
	Label:    "consent",
	Category: "consent",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Constraint: "unique", Type: models.VARCHAR.String(), Required: true, Readonly: true, Index: 0},
		{Name: "optionnal", Type: models.BOOLEAN.String(), Required: true, Default: false, Index: 1},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Label: "template attached", Index: 2},
	},
}

var DBConsentResponse = models.SchemaModel{
	Name:     RootName("consent_response"),
	Label:    "consent response",
	Category: "consent",
	Fields: []models.FieldModel{
		{Name: "is_consenting", Label: "consentant", Type: models.BOOLEAN.String(), Required: true, Readonly: false, Index: 0},
		{Name: RootID(DBConsent.Name), Type: models.INTEGER.String(), ForeignTable: DBConsent.Name, Required: true, Readonly: true, Label: "consent template attached", Index: 2, Hidden: true},
	},
}

// DBTask express a task in the database, a task is an activity to achieve a step in a request
var DBTask = models.SchemaModel{
	Name:     RootName("task"),
	Label:    "activity",
	Category: "request",
	Fields: []models.FieldModel{
		{Name: RootID("dest_table"), Type: models.INTEGER.String(), Required: false, Readonly: true, Label: "reference", Index: 0},
		{Name: models.NAMEKEY, Label: "task to be done", Type: models.VARCHAR.String(), Required: true, Readonly: true, Index: 1},
		{Name: "description", Type: models.BIGVARCHAR.String(), Required: false, Index: 11},
		{Name: "state", Type: models.ENUMSTATE.String(), Required: false, Default: models.STATEPENDING, Index: 2},
		{Name: "is_close", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 3, Hidden: true},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Readonly: true, Label: "assigned to entity", Index: 3},
		{Name: RootID(DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: DBEntity.Name, Required: false, Readonly: true, Label: "assigned to user", Index: 4},
		{Name: "urgency", Type: models.ENUMURGENCY.String(), Required: false, Default: models.LEVELNORMAL, Readonly: true, Index: 5},
		{Name: "priority", Type: models.ENUMURGENCY.String(), Required: false, Default: models.LEVELNORMAL, Readonly: true, Index: 6},
		{Name: "closing_date", Type: models.TIMESTAMP.String(), Required: false, Readonly: true, Index: 7},
		{Name: "closing_by" + RootID(DBUser.Name), Type: models.TIMESTAMP.String(), Required: false, Readonly: true, Index: 8},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Label: "template attached", Index: 9},
		{Name: RootID(DBRequest.Name), Type: models.INTEGER.String(), ForeignTable: DBRequest.Name, Required: true, Readonly: true, Label: "request attached", Index: 10},
		{Name: RootID(DBWorkflowSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBWorkflowSchema.Name, Required: false, Hidden: true, Readonly: true, Label: "workflow attached", Index: 11},
		{Name: "nexts", Type: models.BIGVARCHAR.String(), Required: false, Default: "all", Hidden: true, Index: 12},
		{Name: "meta_" + RootID(DBRequest.Name), Type: models.INTEGER.String(), ForeignTable: DBRequest.Name, Required: false, Hidden: true, Readonly: true, Label: "meta request attached", Index: 13},
		{Name: "binded_dbtask", Type: models.INTEGER.String(), ForeignTable: "dbtask", Required: false, Readonly: true, Label: "binded task", Hidden: true, Index: 14},
		{Name: "passive", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 14, Hidden: true},
	},
}

var DBComment = models.SchemaModel{
	Name:     RootName("comment"),
	Label:    "commentary",
	Category: "",
	Fields: []models.FieldModel{
		{Name: "content", Type: models.VARCHAR.String(), Required: true, Readonly: true, Index: 0},
		{Name: "index", Type: models.INTEGER.String(), Required: false, Readonly: true, Default: 0, Level: models.LEVELRESPONSIBLE, Index: 1},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Readonly: true, Label: "comment by", Level: models.LEVELRESPONSIBLE, Index: 2},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Label: "template attached", Index: 3},
	},
}

var DBCommunicationTemplate = models.SchemaModel{
	Name:     RootName("communication_template"),
	Label:    "communication template",
	Category: "",
	Fields: []models.FieldModel{
		{Name: "content", Type: models.VARCHAR.String(), Required: true, Readonly: false, Index: 0},
		{Name: "from_" + RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: true, Readonly: false, Label: "from user", Level: models.LEVELRESPONSIBLE, Index: 2},
		{Name: "to_" + RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: true, Readonly: false, Label: "to user", Level: models.LEVELRESPONSIBLE, Index: 3},
		{Name: "title", Type: models.VARCHAR.String(), Required: true, Readonly: false, Index: 4},
		{Name: "format", Type: models.VARCHAR.String(), Required: true, Readonly: false, Index: 5},
		{Name: "plateform", Type: models.ENUMPLATFORM.String(), Required: true, Readonly: false, Index: 6},
	},
}

// DBFilter express a filter in the database, a filter is a set of conditions to filter a view on a table
var DBFilter = models.SchemaModel{
	Name:     RootName("filter"),
	Label:    "filter",
	Category: "",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Index: 0},
		{Name: "is_view", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 1},
		{Name: "is_selected", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 2},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: false, Index: 3},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Index: 4},
		{Name: RootID(DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: DBEntity.Name, Required: false, Index: 5},
		{Name: "elder", Type: models.ENUMLIFESTATE.String(), Required: false, Default: "all", Index: 6},
		{Name: "dashboard_restricted", Type: models.BOOLEAN.String(), Required: true, Default: false, Index: 7},
		{Name: "hidden", Type: models.BOOLEAN.String(), Required: false, Default: true, Index: 8},
	},
}

// DBFilterField express a filter field in the database, a filter field is a condition to filter a view on a table
var DBFilterField = models.SchemaModel{
	Name:     RootName("filter_field"),
	Label:    "filter field",
	Category: "",
	Fields: []models.FieldModel{
		{Name: RootID(DBSchemaField.Name), Type: models.INTEGER.String(), ForeignTable: DBSchemaField.Name, Required: false, Index: 0},
		{Name: "value", Type: models.BIGVARCHAR.String(), Required: false, Index: 1},
		{Name: "operator", Type: models.ENUMOPERATOR.String(), Required: false, Index: 2},
		{Name: "separator", Type: models.ENUMSEPARATOR.String(), Required: false, Index: 3},
		{Name: "dir", Type: models.BIGVARCHAR.String(), Required: false, Index: 4},
		{Name: "index", Type: models.INTEGER.String(), Required: false, Default: 1, Index: 5},
		{Name: "width", Type: models.DECIMAL.String(), Required: false, Index: 6},
		{Name: "is_own", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 7},
		{Name: RootID(DBFilter.Name), Type: models.INTEGER.String(), ForeignTable: DBFilter.Name, Required: false, Index: 8},
	},
}

// DBDashboardElement express a dashboard in the database, a dashboard is a set of views on a table
var DBDashboard = models.SchemaModel{
	Name:     RootName("dashboard"),
	Label:    "dashboard",
	Category: "",
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
	Category: "",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Required: true, Index: 0},
		{Name: "description", Type: models.BIGVARCHAR.String(), Required: false, Index: 1},
		{Name: "type", Type: models.ENUMTIME.String(), Required: false, Index: 2},
		{Name: "X", Type: models.INTEGER.String(), ForeignTable: DBDashboardLabel.Name, Required: true, Index: 3},
		{Name: "Y", Type: models.VARCHAR.String(), ForeignTable: DBDashboardLabel.Name, Required: false, Index: 4},
		{Name: "Z", Type: models.VARCHAR.String(), ForeignTable: DBDashboardLabel.Name, Required: false, Index: 5},
		{Name: RootID(DBDashboardMathField.Name), Type: models.INTEGER.String(), ForeignTable: DBDashboardMathField.Name, Required: false, Index: 6},
		{Name: RootID(DBFilter.Name), Type: models.INTEGER.String(), ForeignTable: DBFilter.Name, Required: false, Index: 7},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Index: 8},                          // results if multiple must be ordered by
		{Name: "order_by_" + RootID(DBSchemaField.Name), Type: models.INTEGER.String(), ForeignTable: DBSchemaField.Name, Required: false, Index: 9}, // results if multiple must be ordered by
		{Name: RootID(DBDashboard.Name), Type: models.INTEGER.String(), ForeignTable: DBDashboard.Name, Required: true, Index: 10},
	},
}

// DBDashboardMathField express a dashboard math field in the database, a dashboard math field is a math operation on a column
var DBDashboardLabel = models.SchemaModel{
	Name:     RootName("dashboard_math_field"),
	Label:    "dashboard math field",
	Category: "",
	Fields: []models.FieldModel{
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: false, Index: 1},
		{Name: RootID(DBSchemaField.Name), Type: models.INTEGER.String(), ForeignTable: DBSchemaField.Name, Required: false, Index: 2},
		{Name: "type", Type: models.VARCHAR.String(), Required: false, Index: 3},
	},
}

// DBDashboardMathField express a dashboard math field in the database, a dashboard math field is a math operation on a column
var DBDashboardMathField = models.SchemaModel{
	Name:     RootName("dashboard_math_field"),
	Label:    "dashboard math field",
	Category: "",
	Fields: []models.FieldModel{
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Index: 1},
		{Name: RootID(DBSchemaField.Name), Type: models.INTEGER.String(), ForeignTable: DBSchemaField.Name, Required: true, Index: 2},
		{Name: "column_math_func", Type: models.ENUMMATHFUNC.String(), Required: false, Index: 3}, // func applied on operation added on column value ex: COUNT
		{Name: "row_math_func", Type: models.VARCHAR.String(), Required: false, Index: 4},
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
		{Name: "shortcut_on_schema", Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Level: models.LEVELRESPONSIBLE, Required: false, Index: 6},
		{Name: "is_empty", Type: models.BOOLEAN.String(), Required: false, Default: false, Index: 7},
		{Name: "readonly", Type: models.BOOLEAN.String(), Required: true, Index: 8},
		{Name: "view_" + RootID(DBFilter.Name), Type: models.INTEGER.String(), ForeignTable: DBFilter.Name, Required: false, Index: 9},
		{Name: RootID(DBFilter.Name), Type: models.INTEGER.String(), ForeignTable: DBFilter.Name, Required: false, Index: 10},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Index: 11},
		{Name: "own_view", Type: models.BOOLEAN.String(), Required: false, Index: 12},
		{Name: "only_not_empty", Type: models.BOOLEAN.String(), Required: false, Index: 13},
		{Name: "foldered", Type: models.INTEGER.String(), ForeignTable: DBSchemaField.Name, Required: false, Level: models.LEVELRESPONSIBLE, Index: 14},
		{Name: "permit_on_action", Type: models.VARCHAR.String(), Level: models.LEVELRESPONSIBLE, Required: false, Index: 15},
	},
}

// DBViewAttribution express a view attribution in the database for a user or an entity
var DBViewAttribution = models.SchemaModel{
	Name:     RootName("view_attribution"),
	Label:    "view attribution",
	Category: "",
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
	Category: "",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Type: models.VARCHAR.String(), Required: true, Index: 0},
		{Name: "description", Type: models.BIGVARCHAR.String(), Required: false, Index: 1},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Readonly: true, Label: "assigned user", Index: 2},
		{Name: RootID(DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: DBEntity.Name, Required: false, Readonly: true, Label: "assigned entity", Index: 3},
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
		{Name: "access_date", Type: models.TIMESTAMP.String(), Required: false, Default: "CURRENT_TIMESTAMP", Readonly: true, Label: "access date", Index: 2},
		{Name: RootID("dest_table"), Type: models.INTEGER.String(), Required: false, Readonly: true, Label: "reference", Index: 3},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Label: "template attached", Index: 4},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Readonly: true, Label: "related user", Index: 5},
	},
}

var DBDelegation = models.SchemaModel{
	Name:     RootName("delegation"),
	Label:    "delegation",
	Category: "user",
	CanOwned: true,
	Fields: []models.FieldModel{
		{Name: "delegated_" + RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: true, Index: 0, Label: "delegated to user"},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Index: 1, Label: "user with hierarchy"},
		{Name: models.STARTKEY, Type: models.TIMESTAMP.String(), Required: false, Default: "CURRENT_TIMESTAMP", Index: 2},
		{Name: models.ENDKEY, Type: models.TIMESTAMP.String(), Required: false, Index: 3},
		{Name: RootID(DBTask.Name), Type: models.INTEGER.String(), ForeignTable: DBTask.Name, Required: false, Index: 4, Label: "task delegated"},
	},
}

var DBShare = models.SchemaModel{
	Name:     RootName("share"),
	Label:    "share",
	Category: "user",
	CanOwned: true,
	Fields: []models.FieldModel{
		{Name: "shared_" + RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: true, Index: 0, Label: "shared to user"},
		{Name: RootID(DBUser.Name), Type: models.INTEGER.String(), ForeignTable: DBUser.Name, Required: false, Index: 1, Label: "user with hierarchy"},
		{Name: models.STARTKEY, Type: models.TIMESTAMP.String(), Required: false, Default: "CURRENT_TIMESTAMP", Index: 2},
		{Name: models.ENDKEY, Type: models.TIMESTAMP.String(), Required: false, Index: 3},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Index: 4, Label: "template delegated"},
		{Name: RootID("dest_table"), Type: models.INTEGER.String(), Required: true, Readonly: true, Label: "reference", Index: 5},
		{Name: "read_access", Type: models.BOOLEAN.String(), Required: false, Default: true, Index: 6, Label: "read access"},
		{Name: "update_access", Type: models.BOOLEAN.String(), Required: false, Default: true, Index: 7, Label: "update access"},
		{Name: "delete_access", Type: models.BOOLEAN.String(), Required: false, Default: true, Index: 9, Label: "delete access"},
	},
} // TODO PERMISSION

var OWNPERMISSIONEXCEPTION = []string{DBFilter.Name, DBFilterField.Name, DBNotification.Name, DBDelegation.Name,
	DBDashboard.Name, DBDashboardElement.Name, DBDashboardMathField.Name, DBShare.Name}
var AllPERMISSIONEXCEPTION = []string{DBNotification.Name, DBViewAttribution.Name, DBUser.Name}
var POSTPERMISSIONEXCEPTION = []string{DBEmailSended.Name, DBRequest.Name, DBConsentResponse.Name}
var PUPERMISSIONEXCEPTION = []string{DBTask.Name, DBEmailResponse.Name}
var PERMISSIONEXCEPTION = []string{DBView.Name, DBTask.Name, DBRequest.Name, DBWorkflow.Name, DBEntity.Name, DBSchema.Name, DBSchemaField.Name} // override permission checkup

var ROOTTABLES = []models.SchemaModel{DBSchemaField, DBUser, DBWorkflow, DBView, DBRequest, DBSchema, DBPermission, DBEntity,
	DBRole, DBDataAccess, DBNotification, DBEntityUser, DBRoleAttribution, DBShare,
	DBConsent, DBTask, DBWorkflowSchema, DBRolePermission, DBHierarchy, DBViewAttribution, DBFilter, DBFilterField,
	DBDashboard, DBDashboardElement, DBDashboardMathField, DBDashboardLabel,
	DBComment, DBDelegation,
	DBCommunicationTemplate, DBConsentResponse, DBEmailTemplate,
	DBTrigger, DBTriggerRule, DBTriggerCondition,
	DBFieldAutoFill,
	DBEmailSended, DBEmailTemplate, DBEmailResponse,
}

var NOAUTOLOADROOTTABLES = []models.SchemaModel{DBSchema, DBSchemaField, DBPermission, DBView, DBWorkflow}
var NOAUTOLOADROOTTABLESSTR = []string{DBSchema.Name, DBSchemaField.Name, DBPermission.Name, DBView.Name, DBWorkflow.Name}

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

var ConsentDBField = RootID(DBConsent.Name)
var SchemaDBField = RootID(DBSchema.Name)
var SchemaFieldDBField = RootID(DBSchemaField.Name)
var RequestDBField = RootID(DBRequest.Name)
var TaskDBField = RootID(DBTask.Name)
var NotificationDBField = RootID(DBNotification.Name)
var DataAccessDBField = RootID(DBDataAccess.Name)
var WorkflowDBField = RootID(DBWorkflow.Name)
var WorkflowSchemaDBField = RootID(DBWorkflowSchema.Name)
var UserDBField = RootID(DBUser.Name)
var EntityDBField = RootID(DBEntity.Name)
var DestTableDBField = RootID("dest_table")
var FilterDBField = RootID(DBFilter.Name)
var FilterFieldDBField = RootID(DBFilterField.Name)
var ViewFilterDBField = "view_" + RootID(DBFilter.Name)
var ViewDBField = RootID(DBView.Name)
var DashboardDBField = RootID(DBDashboard.Name)
var DashboardMathDBField = RootID(DBDashboardMathField.Name)
var DashboardElementDBField = RootID(DBDashboardElement.Name)
var ViewAttributionDBField = RootID(DBViewAttribution.Name)
var TriggerDBField = RootID(DBTrigger.Name)
var EmailTemplateDBField = RootID(DBEmailTemplate.Name)
var EmailSendedDBField = RootID(DBEmailSended.Name)
