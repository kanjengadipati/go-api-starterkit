package erroroptimizer

// ErrorMetadata contains error classification info
type ErrorMetadata struct {
	Code                string
	Type                string
	Severity            string
	UserMessage         string
	ShouldExposeDetails bool
}

// Suggestion represents an actionable step
type Suggestion struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Action      string `json:"action"`
	URL         string `json:"url,omitempty"`
	Priority    string `json:"priority"`
}

// OptimizedError is what we return to the client
type OptimizedError struct {
	Code        string       `json:"code"`
	Message     string       `json:"message"`
	Details     string       `json:"details,omitempty"`
	Suggestions []Suggestion `json:"suggestions,omitempty"`
	DocsURL     string       `json:"docs_url,omitempty"`
	SupportURL  string       `json:"support_url,omitempty"`
}

// UserContext provides information about the user
type UserContext struct {
	UserID         *uint
	Device         string
	Language       string
	PreviousErrors []string
	IsNewUser      bool
}
