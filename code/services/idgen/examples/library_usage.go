package main

import (
	"database/sql"
	"fmt"
	"log"

	"idgen"

	_ "github.com/lib/pq"
)

func main() {
	// Connect to database
	db, err := sql.Open("postgres", "postgres://postgres:jagadees@localhost:5432/idgen?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create ID generator instance
	generator := idgen.New(db)

	// Register a template
	config := &idgen.TemplateConfig{
		Template: "COURT-{COURTID}-{DATE:yyyyMMdd}-{SEQ:0000}",
		Sequence: &idgen.SequenceConfig{
			Scope: "daily",
			Start: 1,
			Padding: idgen.PaddingConfig{
				Length: 4,
				Char:   "0",
			},
		},
	}

	err = generator.RegisterTemplate("court_order", config)
	if err != nil {
		log.Fatal("Failed to register template:", err)
	}

	// Generate IDs with different variables
	variables1 := map[string]string{
		"COURTID": "HC001",
	}

	id1, err := generator.GenerateID("court_order", variables1)
	if err != nil {
		log.Fatal("Failed to generate ID:", err)
	}
	fmt.Println("Generated ID 1:", id1)

	// Generate another ID with same template
	id2, err := generator.GenerateID("court_order", variables1)
	if err != nil {
		log.Fatal("Failed to generate ID:", err)
	}
	fmt.Println("Generated ID 2:", id2)

	// Generate ID with different court
	variables2 := map[string]string{
		"COURTID": "SC002",
	}
	id3, err := generator.GenerateID("court_order", variables2)
	if err != nil {
		log.Fatal("Failed to generate ID:", err)
	}
	fmt.Println("Generated ID 3:", id3)
}
