package models

import "strings"

// Generate templates from a scheme
type Link struct {
	Source      string
	Destination string
}

func BuildLinks(schema []TableEntity) []Link {
	var links []Link
	for _, ti := range schema {
		for column := range ti.AssColumns {
			if strings.HasSuffix(column, "_id") {
				tokens := strings.Split(column, "_")
				linkedtable := tokens[len(tokens)-2]
				var link Link
				link.Source = ti.Name
				link.Destination = linkedtable
				links = append(links, link)
			}
		}
	}
	return links
}
