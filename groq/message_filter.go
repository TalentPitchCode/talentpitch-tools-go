package groq

import (
	"context"
)

// FilterMessageWithAI checks if a message is malicious using Groq AI
// This is a convenience wrapper around CheckMessageContent
// Returns:
//   - isMalicious: true if message should be rejected
//   - errorCode: error code for rejection reason
//   - reason: reason for rejection
//   - error: any error during the check
func (c *Client) FilterMessageWithAI(ctx context.Context, messageText string) (isMalicious bool, errorCode string, reason string, err error) {
	return c.CheckMessageContent(ctx, messageText)
}

