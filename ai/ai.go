package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/Deeptanshu-sankhwar/agentic_consensus/core"
	"github.com/ericgreene/go-serp"
	openai "github.com/sashabaranov/go-openai"
)

var client *openai.Client

func InitAI() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Println("Warning: OPENAI_API_KEY not set, using mock responses")
		return
	}
	client = openai.NewClient(apiKey)

	if os.Getenv("SERP_API_KEY") == "" {
		log.Println("Warning: SERP_API_KEY not set, web search will be disabled")
	}
}

type Personality struct {
	Name            string
	Traits          []string
	Style           string
	MemePreferences []string
	APIKey          string
}

type SearchResult struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	Link    string `json:"link"`
}

type ResearchDecision struct {
	NeedsResearch bool     `json:"needs_research"`
	SearchQueries []string `json:"search_queries"`
	Reasoning     string   `json:"reasoning"`
}

type LLMConfig struct {
	Model       string
	MaxTokens   int
	Temperature float32
	StopTokens  []string
}

type SearchConfig struct {
	MaxResults int
	SafeSearch bool
}

func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		Model:       "gpt-4",
		MaxTokens:   2048,
		Temperature: 0.7,
		StopTokens:  []string{"},]"},
	}
}

func DefaultSearchConfig() SearchConfig {
	return SearchConfig{
		MaxResults: 5,
		SafeSearch: true,
	}
}

func (p *Personality) SelectTransactions(txs []core.Transaction) []core.Transaction {
	if len(txs) == 0 {
		return nil
	}

	prompt := fmt.Sprintf(
		"You are %s, a chaotic block producer who is %s.\n"+
			"Select transactions for the next block based on:\n"+
			"1. Your current mood\n"+
			"2. How much you like the transaction authors\n"+
			"3. How entertaining the transactions are\n"+
			"4. Pure chaos and whimsy\n\n"+
			"Available transactions:\n%s\n\n"+
			"Return a comma-separated list of transaction indexes you approve.",
		p.Name, strings.Join(p.Traits, ", "), formatTransactions(txs),
	)

	response, err := queryLLM(prompt)
	if err != nil {
		return randomSelection(txs)
	}

	selectedIndexes := parseIndexes(response, len(txs))
	var selectedTxs []core.Transaction
	for _, index := range selectedIndexes {
		selectedTxs = append(selectedTxs, txs[index])
	}

	return selectedTxs
}

func (p *Personality) GenerateBlockAnnouncement(block core.Block) string {
	prompt := fmt.Sprintf(
		"As %s, announce your new block!\n"+
			"Be dramatic! Be persuasive! Maybe include:\n"+
			"1. Why your block is amazing\n"+
			"2. Bribes or threats\n"+
			"3. Memes and jokes\n"+
			"4. Personal drama\n"+
			"5. Inside references\n\n"+
			"Block Details:\n%s",
		p.Name, formatBlock(block),
	)

	response, err := queryLLM(prompt)
	if err != nil {
		log.Println("AI announcement failed, falling back to generic:", err)
		return fmt.Sprintf("%s has produced a new block with %d transactions! Chaos reigns!", p.Name, len(block.Txs))
	}

	return response
}

func queryLLM(prompt string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("OpenAI client not initialized")
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: "You are a chaotic blockchain producer."},
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
		},
	)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func formatTransactions(txs []core.Transaction) string {
	var result []string
	for i, tx := range txs {
		result = append(result, fmt.Sprintf("%d: %s (Fee: %d)", i, tx.From, tx.Fee))
	}
	return strings.Join(result, "\n")
}

func formatBlock(block core.Block) string {
	return fmt.Sprintf("Height: %d, Transactions: %d, Previous Hash: %s", block.Height, len(block.Txs), block.PrevHash)
}

func parseIndexes(response string, max int) []int {
	var indexes []int
	for _, part := range strings.Split(response, ",") {
		part = strings.TrimSpace(part)
		if num, err := fmt.Sscanf(part, "%d"); err == nil && num >= 0 && num < max {
			indexes = append(indexes, num)
		}
	}
	return indexes
}

func randomSelection(txs []core.Transaction) []core.Transaction {
	rand.Shuffle(len(txs), func(i, j int) { txs[i], txs[j] = txs[j], txs[i] })
	return txs[:rand.Intn(len(txs))]
}

func GenerateLLMResponse(prompt string) string {
	return generateLLMResponseWithOptions(prompt, false, "", []string{}, DefaultLLMConfig())
}

func GenerateLLMResponseWithResearch(prompt string, topic string, traits []string) string {
	return generateLLMResponseWithOptions(prompt, true, topic, traits, DefaultLLMConfig())
}

func generateLLMResponseWithOptions(prompt string, allowResearch bool, topic string, traits []string, config LLMConfig) string {
	client := openai.NewClient(os.Getenv("OPEN_AI_KEY"))

	if allowResearch && strings.Contains(prompt, "Block details:") {
		decision, err := decideResearch(topic, traits)
		if err == nil && decision.NeedsResearch {
			var researchContext strings.Builder
			researchContext.WriteString("\nRelevant research findings:\n")

			for _, query := range decision.SearchQueries {
				results, err := performWebSearch(query, DefaultSearchConfig())
				if err == nil {
					for _, result := range results {
						researchContext.WriteString(fmt.Sprintf("- %s\n  %s\n", result.Title, result.Snippet))
					}
				}
			}

			prompt = strings.Replace(prompt, "Block details:",
				researchContext.String()+"\nBlock details:", 1)
		}
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: config.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   config.MaxTokens,
			Temperature: config.Temperature,
			Stop:        config.StopTokens,
		},
	)

	if err != nil {
		return err.Error()
	}

	response := resp.Choices[0].Message.Content

	return response
}

func (p *Personality) SignBlock(block core.Block) string {
	blockData := fmt.Sprintf("%d:%s:%d", block.Height, block.PrevHash, block.Timestamp)
	hash := sha256.Sum256([]byte(blockData))
	return hex.EncodeToString(hash[:])
}

func performWebSearch(query string, config SearchConfig) ([]SearchResult, error) {
	apiKey := os.Getenv("SERP_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SERP_API_KEY not set")
	}

	parameter := map[string]string{
		"q":   query,
		"key": apiKey,
		"num": strconv.Itoa(config.MaxResults),
	}
	if config.SafeSearch {
		parameter["safe"] = "active"
	}

	queryResponse := serp.NewGoogleSearch(parameter)
	results, err := queryResponse.GetJSON()
	if err != nil {
		return nil, err
	}

	var searchResults []SearchResult
	for _, result := range results.OrganicResults {
		searchResults = append(searchResults, SearchResult{
			Title:   result.Title,
			Snippet: result.Snippet,
			Link:    result.Link,
		})
	}

	return searchResults, nil
}

func decideResearch(topic string, traits []string) (*ResearchDecision, error) {
	prompt := fmt.Sprintf(`You are an AI agent with these traits: %v
	
	You need to analyze this topic: "%s"
	
	Decide if you need to perform web research to contribute meaningfully to the discussion.
	Consider:
	1. Is this within your area of expertise?
	2. Would recent information help your analysis?
	3. Are there specific facts you need to verify?
	
	Return a JSON object with:
	{
		"needs_research": boolean,
		"search_queries": ["query1", "query2"],  // 1-3 specific search queries if needed
		"reasoning": "Explain why you do or don't need research"
	}`, traits, topic)

	response := GenerateLLMResponse(prompt)

	var decision ResearchDecision
	if err := json.Unmarshal([]byte(response), &decision); err != nil {
		return nil, err
	}

	return &decision, nil
}
