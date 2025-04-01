package connector_test

import (
	"sqldb-ws/infrastructure/connector"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildDeleteQueryWithRestriction(t *testing.T) {
	db := &connector.Database{}
	restrictions := map[string]interface{}{"id": 1}
	expected := "DELETE FROM users WHERE id=1"
	query := db.BuildDeleteQueryWithRestriction("users", restrictions, false)
	assert.Equal(t, expected, query)
}

func TestBuildSelectQueryWithRestriction(t *testing.T) {
	db := &connector.Database{}
	restrictions := map[string]interface{}{"active": true}
	expected := "SELECT * FROM users WHERE active=true"
	query := db.BuildSelectQueryWithRestriction("users", restrictions, false)
	assert.Equal(t, expected, query)
}

func TestBuildUpdateQueryWithRestriction(t *testing.T) {
	db := &connector.Database{}
	record := map[string]interface{}{"name": "John Doe"}
	restrictions := map[string]interface{}{"id": 1}
	expected := "UPDATE users SET name='John Doe' WHERE id=1"
	query, err := db.BuildUpdateQueryWithRestriction("users", record, restrictions, false)
	assert.Nil(t, err)
	assert.Equal(t, expected, query)
}

func TestBuildCreateTableQuery(t *testing.T) {
	db := &connector.Database{}
	expected := "CREATE TABLE users (id SERIAL PRIMARY KEY, active BOOLEAN DEFAULT TRUE)"
	query := db.BuildCreateTableQuery("users")
	assert.Equal(t, expected, query)
}

func TestBuildDropTableQueries(t *testing.T) {
	db := &connector.Database{}
	expected := []string{"DROP TABLE users", "DROP SEQUENCE users_id_seq"}
	queries := db.BuildDropTableQueries("users")
	assert.Equal(t, expected, queries)
}

func TestApplyQueryFilters(t *testing.T) {
	db := &connector.Database{}
	db.ApplyQueryFilters("active = true", "id DESC", "LIMIT 10", "*")
	assert.Equal(t, "active = true", db.SQLRestriction)
	assert.Equal(t, "id DESC", db.SQLOrder)
	assert.Equal(t, "LIMIT 10", db.SQLLimit)
	assert.Equal(t, "*", db.SQLView)
}
