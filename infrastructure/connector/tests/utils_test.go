package connector_test

import (
	"sqldb-ws/infrastructure/connector"
	"testing"
)

func TestQuote(t *testing.T) {
	got := connector.Quote("test")
	want := "'test'"
	if got != want {
		t.Errorf("Quote() = %v, want %v", got, want)
	}
}

func TestRemoveLastChar(t *testing.T) {
	got := connector.RemoveLastChar("hello!")
	want := "hello"
	if got != want {
		t.Errorf("RemoveLastChar() = %v, want %v", got, want)
	}
}

func TestFormatMathViewQuery(t *testing.T) {
	got := connector.FormatMathViewQuery("SUM", "amount")
	want := "SUM(amount) as result"
	if got != want {
		t.Errorf("FormatMathViewQuery() = %v, want %v", got, want)
	}
}

func TestFormatSQLRestrictionWhereInjection(t *testing.T) {
	mockFunc := func(s string) (string, string, error) {
		return "int", "table", nil
	}
	got := connector.FormatSQLRestrictionWhereInjection("id:5", mockFunc)
	want := "( id IN (SELECT id FROM table WHERE id = 5) )"
	if got != want {
		t.Errorf("FormatSQLRestrictionWhereInjection() = %v, want %v", got, want)
	}
}

func TestMakeSqlItem(t *testing.T) {
	got := connector.MakeSqlItem("", "int", "", "id", "5", "=")
	want := "id = 5"
	if got != want {
		t.Errorf("MakeSqlItem() = %v, want %v", got, want)
	}
}

func TestFormatLimit(t *testing.T) {
	got := connector.FormatLimit("10", "5")
	want := "LIMIT 10 OFFSET 5"
	if got != want {
		t.Errorf("FormatLimit() = %v, want %v", got, want)
	}
}

func TestFormatOperatorSQLRestriction(t *testing.T) {
	got := connector.FormatOperatorSQLRestriction("LIKE", "AND", "name", "john", "text")
	want := "name::text LIKE '%john%'"
	if got != want {
		t.Errorf("FormatOperatorSQLRestriction() = %v, want %v", got, want)
	}
}

func TestFormatSQLRestrictionByList(t *testing.T) {
	got := connector.FormatSQLRestrictionByList("", []interface{}{"id=1", "name='John'"}, false)
	want := "id=1 AND name='John'"
	if got != want {
		t.Errorf("FormatSQLRestrictionByList() = %v, want %v", got, want)
	}
}

func TestFormatSQLOrderBy(t *testing.T) {
	mockFunc := func(s string) bool { return true }
	got := connector.FormatSQLOrderBy([]string{"name"}, []string{"ASC"}, mockFunc)
	want := []string{"name ASC"}
	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("FormatSQLOrderBy() = %v, want %v", got, want)
	}
}

func TestFormatForSQL(t *testing.T) {
	got := connector.FormatForSQL("text", "hello")
	want := "'hello'"
	if got != want {
		t.Errorf("FormatForSQL() = %v, want %v", got, want)
	}
}

func TestSQLInjectionProtector(t *testing.T) {
	got := connector.SQLInjectionProtector("SELECT * FROM users")
	want := "SELECT * FROM users"
	if got != want {
		t.Errorf("SQLInjectionProtector() = %v, want %v", got, want)
	}
}

func TestFormatEnumName(t *testing.T) {
	got := connector.FormatEnumName("Hello, World")
	want := "hello_world"
	if got != want {
		t.Errorf("FormatEnumName() = %v, want %v", got, want)
	}
}

// More test cases covering different edge cases...

func TestFormatSQLRestrictionWhereByMap(t *testing.T) {
	got := connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{"id": 1, "name": "John"}, false)
	want := "id=1 AND name=John"
	if got != want {
		t.Errorf("FormatSQLRestrictionWhereByMap() = %v, want %v", got, want)
	}
}

func TestFormatSQLRestrictionWhere(t *testing.T) {
	verifyFunc := func() bool { return true }
	got, _ := connector.FormatSQLRestrictionWhere("", "id=1", verifyFunc)
	want := "id=1"
	if got != want {
		t.Errorf("FormatSQLRestrictionWhere() = %v, want %v", got, want)
	}
}

func TestFormatSQLRestrictionWhereInjectionEmpty(t *testing.T) {
	mockFunc := func(s string) (string, string, error) {
		return "", "", nil
	}
	got := connector.FormatSQLRestrictionWhereInjection("", mockFunc)
	if got != "" {
		t.Errorf("FormatSQLRestrictionWhereInjection() with empty input should return empty, got %v", got)
	}
}

func TestFormatSQLRestrictionByListEmpty(t *testing.T) {
	got := connector.FormatSQLRestrictionByList("", []interface{}{}, false)
	if got != "" {
		t.Errorf("FormatSQLRestrictionByList() with empty list should return empty, got %v", got)
	}
}

func TestFormatSQLOrderByEmpty(t *testing.T) {
	mockFunc := func(s string) bool { return true }
	got := connector.FormatSQLOrderBy([]string{}, []string{}, mockFunc)
	want := []string{"id DESC"}
	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("FormatSQLOrderBy() with empty input should return default, got %v", got)
	}
}
