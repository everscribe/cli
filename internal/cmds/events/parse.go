package events

import "encoding/json"

// actorSummary / targetSummary / resultSummary mirror the structs the
// Everscribe UI uses to render the event list. Field shapes match the
// JSON the SDK ingests, so we can json-decode the raw event columns
// into them directly.
type actorSummary struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type targetSummary struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type resultSummary struct {
	Status  string `json:"status"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// unmarshalActor returns the zero struct for empty/unparseable input -
// callers fall through to a "-" rendering rather than failing the row.
func unmarshalActor(raw json.RawMessage) actorSummary {
	var a actorSummary
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &a)
	}
	return a
}

func unmarshalTarget(raw json.RawMessage) targetSummary {
	var t targetSummary
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &t)
	}
	return t
}

func unmarshalResult(raw json.RawMessage) resultSummary {
	var r resultSummary
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &r)
	}
	return r
}

// formatActor renders an actorSummary as "<display_name> (<type>)" or
// just "<type>" if there's no display name; falls back to "-".
func formatActor(a actorSummary) string {
	display := a.DisplayName
	if display == "" {
		display = a.ID
	}
	switch {
	case display != "" && a.Type != "":
		return display + " (" + a.Type + ")"
	case display != "":
		return display
	case a.Type != "":
		return a.Type
	default:
		return "-"
	}
}

// formatTarget renders "<type>/<id>" or "<type>" when id is empty;
// falls back to "-" for fully-empty targets.
func formatTarget(t targetSummary) string {
	switch {
	case t.Type != "" && t.ID != "":
		return t.Type + "/" + t.ID
	case t.Type != "":
		return t.Type
	default:
		return "-"
	}
}
