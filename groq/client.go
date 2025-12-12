package groq

import (
	"fmt"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
)

// PromptTemplate is a function that generates a prompt from a message text
type PromptTemplate func(messageText string) string

// Client wraps the Groq OpenAI-compatible client
type Client struct {
	client        *openai.Client
	model         string
	promptBuilder PromptTemplate
	blockedTerms  []string
}

// Config holds configuration for the Groq client
type Config struct {
	// APIKey is the Groq API key (read from GROQ_API_KEY env var if empty)
	APIKey string
	// Model is the Groq model to use (read from GROQ_MODEL env var if empty, defaults to "llama-3.1-8b-instant")
	Model string
	// BaseURL is the Groq API base URL (defaults to "https://api.groq.com/openai/v1")
	BaseURL string
	// PromptTemplate is a function that generates the prompt for content moderation
	// If not provided, a default prompt will be used
	PromptTemplate PromptTemplate
	// BlockedTerms is a list of offensive terms to check before using AI
	// If not provided, a default list will be used
	// If empty slice is provided, blocked terms checking will be disabled
	BlockedTerms []string
}

// NewClient creates a new Groq client with the given configuration
// If APIKey or Model are empty, they will be read from environment variables
// GROQ_API_KEY and GROQ_MODEL respectively
func NewClient(cfg Config) *Client {
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("GROQ_API_KEY")
	}

	if apiKey == "" {
		log.Printf("GROQ_API_KEY not set, Groq client will not be initialized")
		return nil
	}

	model := cfg.Model
	if model == "" {
		model = os.Getenv("GROQ_MODEL")
		if model == "" {
			model = "llama-3.1-8b-instant"
		}
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.groq.com/openai/v1"
	}

	// Create default config and set custom base URL for Groq
	openaiConfig := openai.DefaultConfig(apiKey)
	openaiConfig.BaseURL = baseURL
	client := openai.NewClientWithConfig(openaiConfig)

	// Set prompt template (use default if not provided)
	promptBuilder := cfg.PromptTemplate
	if promptBuilder == nil {
		promptBuilder = defaultPromptTemplate
	}

	// Set blocked terms (use default if not provided)
	blockedTerms := cfg.BlockedTerms
	if blockedTerms == nil {
		// Use default blocked terms if not explicitly set
		blockedTerms = defaultBlockedTerms()
	}
	// If empty slice is provided, blocked terms checking is disabled

	log.Printf("Groq client initialized successfully with model: %s", model)

	return &Client{
		client:        client,
		model:         model,
		promptBuilder: promptBuilder,
		blockedTerms:  blockedTerms,
	}
}

// defaultPromptTemplate returns the default prompt template for content moderation
func defaultPromptTemplate(messageText string) string {
	return fmt.Sprintf(`Analyze the following message and determine if it contains malicious, inappropriate, spam, or harmful content.

Message: "%s"

Respond with ONLY a JSON object in this exact format:
{
  "is_malicious": true or false,
  "error_code": "ERROR_CODE" or null,
  "reason": "brief reason"
}

Error codes to use if malicious:
- CONTENT_SPAM: for spam messages
- CONTENT_INAPPROPRIATE: for inappropriate language or content
- CONTENT_HARASSMENT: for harassment or bullying
- CONTENT_SCAM: for scam or phishing attempts
- CONTENT_VIOLENCE: for violent or threatening content
- CONTENT_OTHER: for other malicious content

If the message is safe, set is_malicious to false and error_code to null.`, messageText)
}

// GetModel returns the configured model name
func (c *Client) GetModel() string {
	if c == nil {
		return "llama-3.1-8b-instant"
	}
	return c.model
}

// GetClient returns the underlying OpenAI client
func (c *Client) GetClient() *openai.Client {
	if c == nil {
		return nil
	}
	return c.client
}
