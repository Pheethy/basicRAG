package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"

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

	ctx := context.Background()
	embedding, err := embedder.EmbedDocuments(ctx, lines)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Embedding for: %v", embedding)
}
