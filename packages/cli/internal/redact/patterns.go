package redact

import "regexp"

var builtinPatterns = []pattern{
	{
		name: "private_key_pem",
		re:   regexp.MustCompile(`(?s)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----.*?-----END [A-Z0-9 ]*PRIVATE KEY-----`),
	},
	{
		name: "connection_string",
		re:   regexp.MustCompile(`\b[a-zA-Z][a-zA-Z0-9+.-]{1,15}://[^\s:@/'"]+:[^\s@/'"]+@[^\s'"]+`),
	},
	{
		name: "authorization_header",
		re:   regexp.MustCompile(`(?i)\bAuthorization["']?\s*:\s*["']?[^"'\r\n]+`),
	},
	{
		name: "bearer_token",
		re:   regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9\-._~+/]+=*`),
	},
	{
		name: "cookie_header",
		re:   regexp.MustCompile(`(?i)\b(?:Cookie|Set-Cookie)["']?\s*:\s*["']?[^"'\r\n]+`),
	},
	{
		name: "vendor_api_key",
		re: regexp.MustCompile(`\b(?:sk-ant-[A-Za-z0-9\-_]{20,}|sk-[A-Za-z0-9]{20,}|` +
			`gh[pousr]_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{20,}|` +
			`xox[baprs]-[A-Za-z0-9\-]{10,}|AKIA[0-9A-Z]{16}|` +
			`AIza[A-Za-z0-9\-_]{20,}|glpat-[A-Za-z0-9\-_]{20,})\b`),
	},
	{
		name: "assigned_secret",
		re: regexp.MustCompile(`(?i)\b(?:api[_-]?key|access[_-]?key|secret[_-]?key|client[_-]?secret|` +
			`token|password|passwd|pwd)\b["']?\s*[:=]\s*["']?[A-Za-z0-9\-_./+]{8,}["']?`),
	},
}
