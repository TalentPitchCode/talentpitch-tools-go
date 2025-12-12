package groq

// MaliciousMessageSaver is an interface that must be implemented by projects
// that want to save malicious messages to their own database
type MaliciousMessageSaver interface {
	// SaveMaliciousMessage saves a rejected message to the database
	// Parameters:
	//   - fromUserID: ID of the user who sent the message
	//   - toUserID: ID of the user who was supposed to receive the message
	//   - messageText: The content of the rejected message
	//   - errorCode: Error code for the rejection reason (e.g., "CONTENT_SPAM", "CONTENT_INAPPROPRIATE")
	//   - reason: Brief reason for rejection
	//   - currentTime: Current timestamp in the format "2006-01-02 15:04:05"
	// Returns:
	//   - error: Any error that occurred during the save operation
	SaveMaliciousMessage(fromUserID int, toUserID int, messageText string, errorCode string, reason string, currentTime string) error
}

// SaveMaliciousMessage is a convenience function that uses the provided saver
// to save a malicious message. This allows projects to implement their own
// database logic while using the shared filtering functionality.
func SaveMaliciousMessage(saver MaliciousMessageSaver, fromUserID int, toUserID int, messageText string, errorCode string, reason string, currentTime string) error {
	if saver == nil {
		return nil // No saver provided, skip saving
	}
	return saver.SaveMaliciousMessage(fromUserID, toUserID, messageText, errorCode, reason, currentTime)
}

