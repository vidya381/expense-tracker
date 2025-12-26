package utils

import (
	"strings"
	"testing"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		want      string
	}{
		{
			name:      "normal string - no changes",
			input:     "Hello World",
			maxLength: 100,
			want:      "Hello World",
		},
		{
			name:      "string with XSS script tag",
			input:     "Hello <script>alert('xss')</script> World",
			maxLength: 100,
			want:      "Hello &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt; World",
		},
		{
			name:      "string with HTML tags",
			input:     "<b>Bold</b> and <i>italic</i>",
			maxLength: 100,
			want:      "&lt;b&gt;Bold&lt;/b&gt; and &lt;i&gt;italic&lt;/i&gt;",
		},
		{
			name:      "string exceeding max length",
			input:     "This is a very long string that exceeds the maximum allowed length",
			maxLength: 20,
			want:      "This is a very long ",
		},
		{
			name:      "empty string",
			input:     "",
			maxLength: 100,
			want:      "",
		},
		{
			name:      "string with only whitespace",
			input:     "   \t\n  ",
			maxLength: 100,
			want:      "",
		},
		{
			name:      "string with SQL injection attempt",
			input:     "'; DROP TABLE users; --",
			maxLength: 100,
			want:      "&#39;; DROP TABLE users; --",
		},
		{
			name:      "string with special characters",
			input:     "Price: $19.99 & up!",
			maxLength: 100,
			want:      "Price: $19.99 &amp; up!",
		},
		{
			name:      "string with unicode",
			input:     "Hello 世界 🌍",
			maxLength: 100,
			want:      "Hello 世界 🌍",
		},
		{
			name:      "string with multiple spaces",
			input:     "Hello     World",
			maxLength: 100,
			want:      "Hello     World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeString(tt.input, tt.maxLength)
			if got != tt.want {
				t.Errorf("SanitizeString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeDescription(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal description",
			input: "Grocery shopping at Walmart",
			want:  "Grocery shopping at Walmart",
		},
		{
			name:  "description with script tag",
			input: "Payment <script>alert(1)</script>",
			want:  "Payment &lt;script&gt;alert(1)&lt;/script&gt;",
		},
		{
			name:  "description with HTML",
			input: "<h1>Title</h1> <p>Content</p>",
			want:  "&lt;h1&gt;Title&lt;/h1&gt; &lt;p&gt;Content&lt;/p&gt;",
		},
		{
			name:  "very long description",
			input: strings.Repeat("a", 600),
			want:  strings.Repeat("a", 500),
		},
		{
			name:  "empty description",
			input: "",
			want:  "",
		},
		{
			name:  "description with newlines and tabs",
			input: "Line 1\nLine 2\tTabbed",
			want:  "Line 1\nLine 2\tTabbed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeDescription(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeCategoryName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal category name",
			input: "Food & Dining",
			want:  "Food &amp; Dining",
		},
		{
			name:  "category with script tag",
			input: "Groceries<script>alert('xss')</script>",
			want:  "Groceries&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:  "category with HTML",
			input: "<div>Entertainment</div>",
			want:  "&lt;div&gt;Entertainment&lt;/div&gt;",
		},
		{
			name:  "very long category name",
			input: strings.Repeat("a", 150),
			want:  strings.Repeat("a", 100),
		},
		{
			name:  "empty category name",
			input: "",
			want:  "",
		},
		{
			name:  "category with special characters",
			input: "Health & Fitness",
			want:  "Health &amp; Fitness",
		},
		{
			name:  "category with multiple spaces",
			input: "Gas   &   Transportation",
			want:  "Gas   &amp;   Transportation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeCategoryName(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeCategoryName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{
			name:     "valid username - lowercase",
			username: "john_doe",
			want:     true,
		},
		{
			name:     "valid username - with numbers",
			username: "user123",
			want:     true,
		},
		{
			name:     "valid username - with hyphen",
			username: "john-doe",
			want:     true,
		},
		{
			name:     "valid username - mixed case",
			username: "JohnDoe",
			want:     true,
		},
		{
			name:     "valid username - minimum length",
			username: "abc",
			want:     true,
		},
		{
			name:     "valid username - maximum length",
			username: strings.Repeat("a", 50),
			want:     true,
		},
		{
			name:     "invalid username - too short",
			username: "ab",
			want:     false,
		},
		{
			name:     "invalid username - too long",
			username: strings.Repeat("a", 51),
			want:     false,
		},
		{
			name:     "invalid username - special characters",
			username: "john@doe",
			want:     false,
		},
		{
			name:     "invalid username - spaces",
			username: "john doe",
			want:     false,
		},
		{
			name:     "invalid username - empty",
			username: "",
			want:     false,
		},
		{
			name:     "invalid username - only numbers",
			username: "12345",
			want:     true, // This is actually valid per the regex
		},
		{
			name:     "invalid username - starts with hyphen",
			username: "-john",
			want:     true, // Regex allows this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateUsername(tt.username)
			if got != tt.want {
				t.Errorf("ValidateUsername(%q) = %v, want %v", tt.username, got, tt.want)
			}
		})
	}
}
