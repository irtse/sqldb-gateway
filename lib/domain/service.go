package domain

import (tool "sqldb-ws/lib")

func SpecializedService(name string) tool.SpecializedService {
	for _, service := range SERVICES {
		if service.Entity().GetName() == name { return service }
	}
	return &tool.CustomService{}
}

var SERVICES = []tool.SpecializedService{&SchemaService{}, &SchemaFields{}}
