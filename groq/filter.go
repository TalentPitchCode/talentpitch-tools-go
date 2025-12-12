package groq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// CheckMessageContent uses Groq to analyze message content and determine if it's malicious
// Returns:
//   - isMalicious: true if the message should be rejected
//   - errorCode: error code for the rejection reason
//   - reason: brief reason for rejection
//   - error: any error that occurred during the check
func (c *Client) CheckMessageContent(ctx context.Context, messageText string) (isMalicious bool, errorCode string, reason string, err error) {
	if c == nil || c.client == nil {
		// If Groq client is not initialized, allow the message (fail open)
		log.Printf("Groq client not initialized, allowing message")
		return false, "", "", nil
	}

	model := c.GetModel()

	// Use the configured prompt template
	prompt := c.promptBuilder(messageText)

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.1, // Low temperature for more consistent moderation
			MaxTokens:   150, // Short response
		},
	)

	if err != nil {
		log.Printf("Error calling Groq API: %v", err)
		// Fail open - allow message if API call fails
		return false, "", "", fmt.Errorf("error calling Groq API: %w", err)
	}

	if len(resp.Choices) == 0 {
		log.Printf("No response from Groq API")
		return false, "", "", fmt.Errorf("no response from Groq API")
	}

	// Parse the JSON response
	responseText := resp.Choices[0].Message.Content
	log.Printf("Groq moderation response: %s", responseText)

	// Clean the response text (remove markdown code blocks if present)
	responseText = strings.TrimSpace(responseText)
	if strings.HasPrefix(responseText, "```json") {
		responseText = strings.TrimPrefix(responseText, "```json")
		responseText = strings.TrimSuffix(responseText, "```")
	} else if strings.HasPrefix(responseText, "```") {
		responseText = strings.TrimPrefix(responseText, "```")
		responseText = strings.TrimSuffix(responseText, "```")
	}
	responseText = strings.TrimSpace(responseText)

	// Parse JSON response
	var moderationResult struct {
		IsMalicious bool   `json:"is_malicious"`
		ErrorCode   string `json:"error_code"`
		Reason      string `json:"reason"`
	}

	if err := json.Unmarshal([]byte(responseText), &moderationResult); err != nil {
		log.Printf("Error parsing Groq JSON response: %v, response: %s", err, responseText)
		// If we can't parse, do a simple check for malicious indicators
		if strings.Contains(strings.ToLower(responseText), "is_malicious") && strings.Contains(strings.ToLower(responseText), "true") {
			return true, "CONTENT_OTHER", "", nil
		}
		// Fail open - allow message if we can't parse
		return false, "", "", nil
	}

	if moderationResult.IsMalicious {
		errorCode := moderationResult.ErrorCode
		if errorCode == "" {
			errorCode = "CONTENT_OTHER"
		}
		log.Printf("Message flagged as malicious: error_code=%s, reason=%s", errorCode, moderationResult.Reason)
		return true, errorCode, moderationResult.Reason, nil
	}

	return false, "", "", nil
}
