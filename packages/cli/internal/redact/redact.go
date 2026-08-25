package redact

import "regexp"

const mask = "[redacted]"

type pattern struct {
	name string
	re   *regexp.Regexp
}

type Redactor struct {
	patterns []pattern
}

func New() *Redactor {
	return &Redactor{patterns: builtinPatterns}
}

func (r *Redactor) Text(s string) (string, int) {
	count := 0

	for _, p := range r.patterns {
		s = p.re.ReplaceAllStringFunc(s, func(match string) string {
			count++
			return mask
		})
	}

	return s, count
}

func (r *Redactor) Value(v any) (any, int) {
	switch val := v.(type) {
	case string:
		return r.Text(val)
	case map[string]any:
		out := make(map[string]any, len(val))
		total := 0

		for k, child := range val {
			redacted, count := r.Value(child)
			out[k] = redacted
			total += count
		}

		return out, total
	case []any:
		out := make([]any, len(val))
		total := 0

		for i, child := range val {
			redacted, count := r.Value(child)
			out[i] = redacted
			total += count
		}

		return out, total
	default:
		return val, 0
	}
}
