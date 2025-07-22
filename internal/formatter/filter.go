package formatter

import (
	"strings"

	"github.com/annenpolka/cclog/pkg/types"
)

// IsContentfulMessage determines if a message contains meaningful content
func IsContentfulMessage(msg types.Message) bool {
	// Filter out system messages
	if msg.Type == "system" {
		return false
	}

	// Filter out summary messages
	if msg.Type == "summary" {
		return false
	}

	// Filter out meta messages
	if msg.IsMeta {
		return false
	}

	// Extract content and check if it's meaningful
	content := ExtractMessageContent(msg.Message)

	// Filter out empty messages
	if content == "" {
		return false
	}

	// Filter out API errors
	if strings.Contains(content, "API Error") {
		return false
	}

	// Filter out interrupted requests
	if strings.Contains(content, "[Request interrupted") {
		return false
	}

	// Filter out command messages
	if strings.Contains(content, "<command-name>") {
		return false
	}

	// Filter out bash inputs
	if strings.Contains(content, "<bash-input>") {
		return false
	}

	// Filter out command outputs
	if strings.Contains(content, "<local-command-stdout>") {
		return false
	}

	// Filter out system reminders and caveats
	if strings.Contains(content, "Caveat: The messages below were generated") {
		return false
	}

	return true
}

// FilterMessages filters a slice of messages based on content quality
func FilterMessages(messages []types.Message, enableFiltering bool) []types.Message {
	if !enableFiltering {
		return messages
	}

	var filtered []types.Message
	for _, msg := range messages {
		if IsContentfulMessage(msg) {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// FilterConversationLog filters messages in a conversation log
func FilterConversationLog(log *types.ConversationLog, enableFiltering bool) *types.ConversationLog {
	return &types.ConversationLog{
		Messages: FilterMessages(log.Messages, enableFiltering),
		FilePath: log.FilePath,
	}
}
