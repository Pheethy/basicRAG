package main

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
)

func ReadLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

type Document struct {
	Content   string `json:"content"`
	Embedding string `json:"embedding"`
}

func main() {
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}

	lines, err := ReadLines("./knowledged.txt")
	if err != nil {
		log.Fatal(err)
	}

	psqlDB, err := sqlx.Connect("postgres", "user=postgres password=pheet1234 host=127.0.0.1 port=5432 dbname=pheet_db_dev sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer psqlDB.Close()

	ctx := context.Background()
	embedding, err := embedder.EmbedDocuments(ctx, lines)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Embedding for: %v", embedding)

	docs := make([]*Document, 0)
	for index := range lines {
		bu, _ := json.Marshal(embedding[index])
		docs = append(docs, &Document{
			Content:   lines[index],
			Embedding: string(bu),
		})
	}

	if err := CreateDocument(psqlDB, ctx, docs); err != nil {
		log.Fatal(err)
	}
}

func CreateDocument(a *sqlx.DB, ctx context.Context, documents []*Document) error {
	tx, err := a.Beginx()
	if err != nil {
		return err
	}

	scriptSQL := `
		INSERT INTO documents (
			content,
			embedding
		) VALUES (
			$1::text,
			$2::vector
		)
	`

	for index := range documents {
		_, err := tx.Exec(scriptSQL,
			documents[index].Content,
			documents[index].Embedding,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
