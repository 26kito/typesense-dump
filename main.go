package main

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/typesense/typesense-go/v3/typesense"
	"github.com/typesense/typesense-go/v3/typesense/api"
	"github.com/typesense/typesense-go/v3/typesense/api/pointer"
)

var client *typesense.Client

func main() {
	// Initialize Fiber app
	app := fiber.New()

	// Initialize Typesense Client
	client = typesense.NewClient(
		typesense.WithServer("http://localhost:8108"),
		typesense.WithAPIKey("xyz"), // Use the same API key as in Docker
	)

	// Route to create a collection
	app.Post("/create-collection", createCollection)

	// Route to add a document
	app.Post("/add-document", addDocumentsFromCSV)

	// Route to search documents
	app.Get("/search", searchDocuments)

	// Start server
	log.Fatal(app.Listen(":3000"))
}

func createCollection(c *fiber.Ctx) error {
	schema := &api.CollectionSchema{
		Name: "books",
		Fields: []api.Field{
			{Name: "title", Type: "string", Facet: pointer.False(), Index: pointer.True()},
			{Name: "author", Type: "string", Facet: pointer.False(), Index: pointer.True()},
			{Name: "year", Type: "int32"},
		},
	}

	_, err := client.Collections().Create(context.Background(), schema)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Collection created successfully"})
}

func addDocumentsFromCSV(c *fiber.Ctx) error {
	// Open the CSV file
	file, err := os.Open("books.csv")
	if err != nil {
		log.Println("Error opening CSV:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to open CSV file"})
	}
	defer file.Close()

	// Read CSV
	reader := csv.NewReader(file)
	_, err = reader.Read() // Skip header
	if err != nil {
		log.Println("Error reading CSV header:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to read CSV header"})
	}

	// Store documents for bulk insert
	var documents []map[string]interface{}

	// Read each row
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("Error reading CSV row:", err)
			continue
		}

		// Convert row to document
		doc := map[string]interface{}{
			"id":     record[0],
			"title":  record[1],
			"author": record[2],
			"year":   record[3],
		}

		documents = append(documents, doc)
	}

	// Convert []map[string]interface{} to []interface{}
	var docsInterface []interface{}
	for _, doc := range documents {
		docsInterface = append(docsInterface, doc)
	}

	// Bulk insert into Typesense
	_, err = client.Collection("books").Documents().Import(
		context.Background(),
		docsInterface,
		&api.ImportDocumentsParams{},
	)

	if err != nil {
		log.Println("Error inserting documents:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to insert documents"})
	}

	log.Println("Successfully inserted", len(documents), "documents")
	return c.JSON(fiber.Map{"message": "Documents inserted successfully", "count": len(documents)})
}

func searchDocuments(c *fiber.Ctx) error {
	query := c.Query("q") // Get query parameter from request

	searchParams := &api.SearchCollectionParams{
		Q:       &query,
		QueryBy: pointer.String("title,author"),
	}

	result, err := client.Collection("books").Documents().Search(context.Background(), searchParams)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}
