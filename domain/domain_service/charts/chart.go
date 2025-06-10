package charts

/*
import (
	"fmt"
	"slices"
	filter "sqldb-ws/domain/filter"
	"sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
)

type Chart struct {
	Domain utils.DomainITF
}

func (l *Chart) GetAxisDatas(dashboardElement utils.Record, axis string) (map[string]float64, []string) { // string is label
	labels := []string{}
	datas := map[string]float64{}
	if sch, err := schema.GetSchemaByID(utils.GetInt(dashboardElement, ds.SchemaDBField)); err == nil {
		restr, sd, ed := l.GetFilterString(sch, utils.GetString(dashboardElement, ds.FilterDBField))
		if m, err := l.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBDashboardMathField.Name, map[string]interface{}{
			utils.SpecialIDParam: dashboardElement[ds.DashboardMathDBField],
		}, false); err == nil && len(m) > 0 {
			return l.getAxisDatas(axis, dashboardElement, m[0], sd, ed, restr, datas, labels, restr)
		}
	}
	return datas, labels
}

func (l *Chart) getAxisDatas(axis string, dashboardElement utils.Record, mathElement utils.Record, starDate string, endDate string,
	filter []string, datas map[string]float64, labels []string, restr []string) (map[string]float64, []string) {
	algo, _ := mathElement["column_math_func"]
	function, _ := mathElement["row_math_func"]
	if i, ok := dashboardElement[axis]; ok {
		if res, err := l.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBDashboardLabel.Name, map[string]interface{}{
			utils.SpecialIDParam: i,
		}, false); err == nil && len(res) > 0 {
			labelRef := res[0]
			if slices.Contains([]string{"day", "week", "month", "year"}, utils.GetString(labelRef, "type")) {

			} else if utils.GetString(labelRef, "type") == "value" {

			} else if _, ok := labelRef[ds.SchemaFieldDBField]; ok {
				datas, labels = l.getDatasBySchemaColumn(
					algo,
					utils.GetInt(mathElement, ds.SchemaFieldDBField), utils.GetInt(mathElement, ds.SchemaFieldDBField),
					utils.GetInt(labelRef, ds.SchemaFieldDBField), utils.GetInt(labelRef, ds.SchemaFieldDBField), datas, labels, restr)
			}
		}
	}
	return datas, labels
}



func (l *Chart) getLabelsByColID(colID int64, sch sm.SchemaModel) []string {
	labels := []string{}

	f, err := sch.GetFieldByID(colID)
	if err != nil {
		return labels
	}
	l.Domain.GetDb().ClearQueryFilter().SetSQLGroupBy(f.Name)
	l.Domain.GetDb().ClearQueryFilter().SetSQLView(f.Name)
	if res, err := l.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(sch.Name, map[string]interface{}{}, false); err == nil {
		for _, r := range res {
			if _, ok := r[f.Name]; ok {
				labels = append(labels, f.Name)
			}
		}
	}
	return labels
}

func (l *Chart) getDatasBySchemaColumn(algo interface{}, mathColID int64, mathSchemaID int64, labelColID int64, labelSchemaID int64, datas map[string]float64, labels []string,
	restr []string) (map[string]float64, []string) {
	schLabel, err := schema.GetSchemaByID(labelSchemaID)
	if err != nil {
		return datas, labels
	}
	schMath, err := schema.GetSchemaByID(mathSchemaID)
	if err != nil {
		return datas, labels
	}
	f, err := schMath.GetFieldByID(mathColID)
	if err != nil {
		return datas, labels
	}
	for _, label := range l.getLabels(labelColID, schLabel) {
		if algo != nil {
			if res, err := l.Domain.GetDb().ClearQueryFilter().SimpleMathQuery(fmt.Sprintf("%v", algo), schMath.Name, map[string]interface{}{
				f.Name: label,
			}, false); err == nil && len(res) > 0 && res[0]["result"] != nil {
				datas[label] = float64(utils.GetInt(res[0], "result"))
			}
		}
	}
	return datas, labels
}

func (l *Chart) GetFilterString(schema sm.SchemaModel, filterID string) ([]string, string, string) {
	alterRestr := []string{}
	startDate, ok := l.Domain.GetParams().Get("start_date")
	if ok {
		alterRestr = append(alterRestr, connector.FormatSQLRestrictionWhereByMap("",
			map[string]interface{}{
				utils.SpecialIDParam: l.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBDataAccess.Name, map[string]interface{}{
					"access_date": "> " + connector.Quote(startDate),
					"write":       true,
				}, false),
			}, false))
	}
	endDate, ok := l.Domain.GetParams().Get("end_date")
	if ok {
		alterRestr = append(alterRestr, connector.FormatSQLRestrictionWhereByMap("",
			map[string]interface{}{
				utils.SpecialIDParam: l.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBDataAccess.Name, map[string]interface{}{
					"access_date": "< " + connector.Quote(endDate),
					"write":       true,
				}, false),
			}, false))
	}
	if line, ok := l.Domain.GetParams().Get(utils.RootFilterLine); ok && schema.Name != ds.DBView.Name {
		alterRestr = append(alterRestr, connector.FormatSQLRestrictionWhereInjection(line, schema.GetTypeAndLinkForField))
	} else {
		filter.NewFilterService(l.Domain).ProcessFilterRestriction(filterID, schema.ID)
	}
	return alterRestr, startDate, endDate
}
// si le label est une valeu
/*
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
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Index: 1},
		{Name: RootID(DBSchemaField.Name), Type: models.INTEGER.String(), ForeignTable: DBSchemaField.Name, Required: true, Index: 2},

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
		{Name: "column_math_func", Type: models.ENUMMATHFUNC.String(), Required: false, Index: 2}, // func applied on operation added on column value ex: COUNT
		{Name: "row_math_func", Type: models.VARCHAR.String(), Required: false, Index: 3},
	},
}
*/
