package validators

import (
	"context"
	"log"

	"github.com/TalentPitchCode/talentpitch-tools-go/groq"
	"github.com/go-playground/validator/v10"
)

// AcceptableMessageValidator creates a validator function for the "acceptable" tag
// that checks if a message is acceptable using the Groq client
// The validator returns true if the message is NOT malicious (i.e., acceptable)
func AcceptableMessageValidator(groqClient *groq.Client) validator.Func {
	return func(fl validator.FieldLevel) bool {
		msg := fl.Field().String()
		
		// If message is empty and field is optional (omitempty), skip validation
		if msg == "" {
			return true
		}

		ctx := context.Background()
		isMalicious, _, _, err := groqClient.FilterMessageWithAI(ctx, msg)
		if err != nil {
			log.Printf("Error validating message with Groq: %v", err)
			// On error, reject the message (fail closed for security)
			return false
		}

		// Return true if message is NOT malicious (acceptable)
		return !isMalicious
	}
}

// RegisterAcceptableValidator is a convenience function that registers the "acceptable"
// validator tag with the provided validator instance and Groq client
func RegisterAcceptableValidator(validate *validator.Validate, groqClient *groq.Client) error {
	return validate.RegisterValidation("acceptable", AcceptableMessageValidator(groqClient))
}

