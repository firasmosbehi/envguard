package validator

import (
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

func TestValidateNewFormats(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		value     string
		wantError bool
	}{
		// datetime (RFC3339)
		{"datetime valid", "datetime", "2024-01-15T10:30:00Z", false},
		{"datetime valid offset", "datetime", "2024-01-15T10:30:00+02:00", false},
		{"datetime invalid", "datetime", "not-a-date", true},

		// date
		{"date valid", "date", "2024-01-15", false},
		{"date invalid format", "date", "15-01-2024", true},
		{"date invalid", "date", "not-a-date", true},

		// time
		{"time valid", "time", "14:30:00", false},
		{"time invalid", "time", "25:00:00", true},
		{"time not a time", "time", "not-a-time", true},

		// timezone
		{"timezone valid UTC", "timezone", "UTC", false},
		{"timezone valid US", "timezone", "America/New_York", false},
		{"timezone valid Europe", "timezone", "Europe/London", false},
		{"timezone invalid", "timezone", "Mars/Phobos", true},

		// color
		{"color hex short", "color", "#fff", false},
		{"color hex long", "color", "#ff00aa", false},
		{"color rgb", "color", "rgb(255, 0, 128)", false},
		{"color rgba", "color", "rgba(255, 0, 128, 0.5)", false},
		{"color hsl", "color", "hsl(120, 50%, 50%)", false},
		{"color invalid", "color", "not-a-color", true},

		// slug
		{"slug valid", "slug", "my-slug-name", false},
		{"slug single word", "slug", "slug", false},
		{"slug invalid uppercase", "slug", "My-Slug", true},
		{"slug invalid space", "slug", "my slug", true},

		// filepath
		{"filepath valid", "filepath", "/path/to/file.txt", false},
		{"filepath relative", "filepath", "./file.txt", false},

		// directory
		{"directory valid", "directory", "/path/to/dir", false},
		{"directory relative", "directory", "./dir", false},

		// locale
		{"locale valid simple", "locale", "en", false},
		{"locale valid region", "locale", "en-US", false},
		{"locale valid script", "locale", "zh-Hans-CN", false},
		{"locale invalid", "locale", "not_a_locale", true},

		// jwt
		{"jwt valid", "jwt", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U", false},
		{"jwt invalid chars", "jwt", "bad jwt.with spaces", true},
		{"jwt too few parts", "jwt", "only-two.parts", true},

		// mongodb-uri
		{"mongodb valid", "mongodb-uri", "mongodb://localhost:27017/mydb", false},
		{"mongodb srv valid", "mongodb-uri", "mongodb+srv://user:pass@cluster.mongodb.net/mydb", false},
		{"mongodb invalid", "mongodb-uri", "http://localhost:27017", true},

		// redis-uri
		{"redis valid", "redis-uri", "redis://localhost:6379/0", false},
		{"redis ssl valid", "redis-uri", "rediss://localhost:6380/0", false},
		{"redis invalid", "redis-uri", "http://localhost:6379", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Version: "1.0",
				Env: map[string]*schema.Variable{
					"VAR": {
						Type:   schema.TypeString,
						Format: tt.format,
					},
				},
			}

			envVars := map[string]string{"VAR": tt.value}
			result := Validate(s, envVars, false, "")

			if tt.wantError && result.Valid {
				t.Errorf("expected error for format=%s value=%q, but validation passed", tt.format, tt.value)
			}
			if !tt.wantError && !result.Valid {
				t.Errorf("expected no error for format=%s value=%q, but got errors: %v", tt.format, tt.value, result.Errors)
			}
		})
	}
}
