package idgen

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
)

// Config structs - exported for library users

type TemplateConfig struct {
	Template string           `json:"template"`
	Sequence *SequenceConfig  `json:"sequence,omitempty"`
	Random   *RandomConfig    `json:"random,omitempty"`
}

type SequenceConfig struct {
	Scope   string         `json:"scope"`
	Start   int64          `json:"start"`
	Padding PaddingConfig  `json:"padding"`
}

type PaddingConfig struct {
	Length int    `json:"length"`
	Char   string `json:"char"`
}

type RandomConfig struct {
	Length  int    `json:"length"`
	Charset string `json:"charset"`
}

// IDGenerator represents an ID generator instance
type IDGenerator struct {
	db *sql.DB
}

// New creates a new IDGenerator instance
func New(db *sql.DB) *IDGenerator {
	return &IDGenerator{db: db}
}

// RegisterTemplate registers a template configuration
func (g *IDGenerator) RegisterTemplate(templateID string, config *TemplateConfig) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	_, err = g.db.Exec(
		`INSERT INTO idgen_templates (id, config, created_at) VALUES ($1, $2, $3)
		 ON CONFLICT (id) DO UPDATE SET config = EXCLUDED.config`,
		templateID, configJSON, time.Now().Unix(),
	)
	if err != nil {
		return fmt.Errorf("failed to store template: %w", err)
	}

	return nil
}

// GenerateID generates an ID using the specified template and variables
func (g *IDGenerator) GenerateID(templateID string, variables map[string]string) (string, error) {
	// Fetch template config from DB
	var configRaw json.RawMessage
	err := g.db.QueryRow(`SELECT config FROM idgen_templates WHERE id = $1`, templateID).Scan(&configRaw)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("template not found: %s", templateID)
	} else if err != nil {
		return "", fmt.Errorf("failed to fetch template: %w", err)
	}

	var config TemplateConfig
	if err := json.Unmarshal(configRaw, &config); err != nil {
		return "", fmt.Errorf("invalid template config: %w", err)
	}

	id, err := g.generateIDFromConfig(templateID, &config, variables)
	if err != nil {
		return "", err
	}

	// Store the generated ID for audit/traceability
	variablesJSON, _ := json.Marshal(variables)
	_, err = g.db.Exec(
		`INSERT INTO idgen_generated (template_id, generated_id, variables, created_at) VALUES ($1, $2, $3, $4)`,
		templateID, id, variablesJSON, time.Now().Unix(),
	)
	if err != nil {
		// Log the error but don't fail the request
		fmt.Printf("Warning: Failed to store generated ID: %v\n", err)
	}

	return id, nil
}

// Token regex: {TOKEN} or {TOKEN:format}
var tokenRe = regexp.MustCompile(`\{([A-Z]+)(?::([^}]+))?}`)

// Java date format to Go layout
var javaToGoDate = map[string]string{
	"yyyy": "2006",
	"MM":   "01",
	"dd":   "02",
	"HH":   "15",
	"mm":   "04",
	"ss":   "05",
}

func javaDateToGo(javaFmt string) string {
	// Replace Java date tokens with Go layout
	out := javaFmt
	for j, g := range javaToGoDate {
		out = strings.ReplaceAll(out, j, g)
	}
	return out
}

// Expand charset like "A-Z0-9" to full set
func expandCharset(charset string) string {
	var out strings.Builder
	rangeRe := regexp.MustCompile(`([A-Za-z0-9])-([A-Za-z0-9])`)
	last := 0
	for _, m := range rangeRe.FindAllStringSubmatchIndex(charset, -1) {
		out.WriteString(charset[last:m[0]])
		start, end := charset[m[2]], charset[m[4]]
		for c := start; c <= end; c++ {
			out.WriteByte(byte(c))
		}
		last = m[1]
	}
	out.WriteString(charset[last:])
	return out.String()
}

// Generate random string of given length from charset
func randomString(length int, charset string) (string, error) {
	chars := []rune(expandCharset(charset))
	if len(chars) == 0 {
		return "", fmt.Errorf("empty charset")
	}
	if len(chars) == 0 {
		return "", fmt.Errorf("empty charset")
	}
	var out strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		out.WriteRune(chars[n.Int64()])
	}
	return out.String(), nil
}

// Get or create sequence for template and date
func (g *IDGenerator) getNextSequence(templateId string, date string, seqCfg *SequenceConfig) (int64, error) {
	seqName := fmt.Sprintf("seq_%s_%s", templateId, date)
	// Create sequence if not exists
	createSQL := fmt.Sprintf(`CREATE SEQUENCE IF NOT EXISTS %s START WITH %d INCREMENT BY 1 MINVALUE 1 CACHE 1;`, seqName, seqCfg.Start)
	if _, err := g.db.Exec(createSQL); err != nil {
		return 0, fmt.Errorf("failed to create sequence: %w", err)
	}
	// Get nextval
	var next int64
	q := fmt.Sprintf(`SELECT nextval('%s')`, seqName)
	if err := g.db.QueryRow(q).Scan(&next); err != nil {
		return 0, fmt.Errorf("failed to get nextval: %w", err)
	}
	return next, nil
}

// Pad sequence number   
func padSeq(n int64, padLen int, padChar string) string {
	s := fmt.Sprintf("%d", n)
	if len(s) >= padLen {
		return s
	}
	return strings.Repeat(padChar, padLen-len(s)) + s
}

// generateIDFromConfig generates an ID from the template config, using the provided variables map.
func (g *IDGenerator) generateIDFromConfig(templateId string, config *TemplateConfig, variables map[string]string) (string, error) {
	if config == nil {
		return "", fmt.Errorf("template config is nil")
	}
	result := tokenRe.ReplaceAllStringFunc(config.Template, func(token string) string {
		matches := tokenRe.FindStringSubmatch(token)
		if len(matches) < 2 {
			return token // Should not happen
		}
		typeName := matches[1]
		arg := ""
		if len(matches) > 2 {
			arg = matches[2]
		}
		switch typeName {
		case "DATE":
			layout := "2006-01-02"
			if arg != "" {
				layout = javaDateToGo(arg)
			}
			return time.Now().Format(layout)
		case "SEQ":
			if config.Sequence == nil {
				return "SEQ" // or error
			}
			// For daily scope, use current date
			var dateStr string
			if config.Sequence.Scope == "daily" {
				dateStr = time.Now().Format("20060102")
			} else {
				dateStr = "global"
			}
			next, err := g.getNextSequence(templateId, dateStr, config.Sequence)
			if err != nil {
				return "SEQERR"
			}
			return padSeq(next, config.Sequence.Padding.Length, config.Sequence.Padding.Char)
		case "RAND":
			if config.Random == nil {
				return "RAND"
			}
			rand, err := randomString(config.Random.Length, config.Random.Charset)
			if err != nil {
				return "RANDERR"
			}
			return rand
		default:
			// Variable replacement
			if val, ok := variables[typeName]; ok {
				return val
			}
			return token // Unknown token, leave as is
		}
	})
	return result, nil
}
