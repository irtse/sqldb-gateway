package service

import (
	"html/template"
	"os"
	conn "sqldb-ws/infrastructure/connector"

	"github.com/rs/zerolog/log"
)

/*
Infrastructure is meant as DDD pattern, as a generic accessor to database and distant services.
Main Procedure of services at Infrastructure level.
*/
type InfraServiceItf interface {
	GetName() string
	Verify(string) (string, bool)
	Math(algo string, restriction ...string) ([]map[string]interface{}, error)
	Get(restriction ...string) ([]map[string]interface{}, error)
	Create(record map[string]interface{}) ([]map[string]interface{}, error)
	Update(record map[string]interface{}, restriction ...string) ([]map[string]interface{}, error)
	Delete(restriction ...string) ([]map[string]interface{}, error)
	Template(restriction ...string) (interface{}, error)
	GenerateFromTemplate(string) error
}
type InfraService struct {
	Name               string                     `json:"name"`
	User               string                     `json:"-"`
	Results            []map[string]interface{}   `json:"-"`
	SuperAdmin         bool                       `json:"-"`
	NoLog              bool                       `json:"-"`
	SpecializedService InfraSpecializedServiceItf `json:"-"`
	DB                 conn.DB
	InfraServiceItf
}

func (service *InfraService) GetName() string {
	return service.Name
}

// Main Service Builder
func (service *InfraService) Fill(name string, admin bool, user string) {
	service.Name = name
	service.User = user
	service.SuperAdmin = admin
}

// Common Service action of generation by template (TO USE)
func (service *InfraService) GenerateFromTemplate(templateName string) error {
	templ, err := service.Template()
	if err != nil {
		return err
	}
	tFile, err := template.ParseFiles(templateName)
	if err != nil {
		return err
	}
	file, err := os.Create(service.Name)
	if err != nil {
		return err
	}
	if tFile.Execute(file, templ) != nil {
		return err
	}
	return nil
}

// Common service error
func (service *InfraService) DBError(res []map[string]interface{}, err error) ([]map[string]interface{}, error) {
	if !service.NoLog && os.Getenv("log") == "enable" {
		log.Error().Msg(err.Error())
	}
	return res, err
}

type InfraSpecializedServiceItf interface {
	GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string)
	SpecializedCreateRow(record map[string]interface{}, tableName string)
	SpecializedUpdateRow(results []map[string]interface{}, record map[string]interface{})
	SpecializedDeleteRow(results []map[string]interface{}, tableName string)
	VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool)
}

type InfraSpecializedService struct{}

func (s *InfraSpecializedService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	return "", "", "", ""
}

func (s *InfraSpecializedService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	return record, nil, false
}

func (s *InfraSpecializedService) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {
	// EMPTY AND PROUD TO BE
}
func (s *InfraSpecializedService) SpecializedUpdateRow(results []map[string]interface{}, record map[string]interface{}) {
	// EMPTY AND PROUD TO BE
}
func (s *InfraSpecializedService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
}
