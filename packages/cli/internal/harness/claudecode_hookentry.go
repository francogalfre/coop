package harness

import (
	"encoding/json"
	"fmt"
)

const coopSource = "coop"

// hookEntry keeps every field it didn't recognize in extra, so a foreign
// tool's entry (e.g. a command hook's "command" field) round-trips
// untouched when coop rewrites the settings file.
type hookEntry struct {
	Type    string
	URL     string
	Timeout int
	Source  string
	extra   map[string]json.RawMessage
}

func (e hookEntry) MarshalJSON() ([]byte, error) {
	m := make(map[string]json.RawMessage, len(e.extra)+4)
	for k, v := range e.extra {
		m[k] = v
	}

	if err := setHookField(m, "type", e.Type); err != nil {
		return nil, err
	}
	if e.URL != "" {
		if err := setHookField(m, "url", e.URL); err != nil {
			return nil, err
		}
	}
	if e.Timeout != 0 {
		if err := setHookField(m, "timeout", e.Timeout); err != nil {
			return nil, err
		}
	}
	if e.Source != "" {
		if err := setHookField(m, "source", e.Source); err != nil {
			return nil, err
		}
	}

	return json.Marshal(m)
}

func (e *hookEntry) UnmarshalJSON(b []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	takeHookString(m, "type", &e.Type)
	takeHookString(m, "url", &e.URL)
	takeHookInt(m, "timeout", &e.Timeout)
	takeHookString(m, "source", &e.Source)

	e.extra = m

	return nil
}

// hookGroup keeps unrecognized top-level fields (e.g. "matcher") in extra
// so they round-trip untouched.
type hookGroup struct {
	Hooks []hookEntry
	extra map[string]json.RawMessage
}

func (g hookGroup) MarshalJSON() ([]byte, error) {
	m := make(map[string]json.RawMessage, len(g.extra)+1)
	for k, v := range g.extra {
		m[k] = v
	}

	if err := setHookField(m, "hooks", g.Hooks); err != nil {
		return nil, err
	}

	return json.Marshal(m)
}

func (g *hookGroup) UnmarshalJSON(b []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	if raw, ok := m["hooks"]; ok {
		if err := json.Unmarshal(raw, &g.Hooks); err != nil {
			return err
		}
		delete(m, "hooks")
	}

	g.extra = m

	return nil
}

func setHookField(m map[string]json.RawMessage, key string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("harness: marshal %s: %w", key, err)
	}

	m[key] = b

	return nil
}

func takeHookString(m map[string]json.RawMessage, key string, dst *string) {
	if raw, ok := m[key]; ok {
		_ = json.Unmarshal(raw, dst)
		delete(m, key)
	}
}

func takeHookInt(m map[string]json.RawMessage, key string, dst *int) {
	if raw, ok := m[key]; ok {
		_ = json.Unmarshal(raw, dst)
		delete(m, key)
	}
}

func isCoopEntry(e hookEntry) bool {
	return e.Type == "http" && e.Source == coopSource
}

func withCoopEntry(groups []hookGroup, event, baseURL string) []hookGroup {
	kept := withoutCoopEntries(groups)

	kept = append(kept, hookGroup{Hooks: []hookEntry{{
		Type:    "http",
		URL:     fmt.Sprintf("%s/hook/%s/%s", baseURL, claudeCodeName, event),
		Timeout: hookTimeoutSecs,
		Source:  coopSource,
	}}})

	return kept
}

func withoutCoopEntries(groups []hookGroup) []hookGroup {
	kept := make([]hookGroup, 0, len(groups))

	for _, g := range groups {
		entries := make([]hookEntry, 0, len(g.Hooks))

		for _, e := range g.Hooks {
			if !isCoopEntry(e) {
				entries = append(entries, e)
			}
		}

		if len(entries) > 0 {
			kept = append(kept, hookGroup{Hooks: entries, extra: g.extra})
		}
	}

	return kept
}
