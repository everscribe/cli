package events

import (
	"fmt"
	"io"
	"strings"

	"github.com/everscribe/cli/internal/output"
	"github.com/everscribe/cli/internal/types"
)

// renderEventsList emits the table-or-json-or-yaml form of a list
// response. `tableHeader` controls whether the table prints its
// header row; watch suppresses it after the first batch.
func renderEventsList(w io.Writer, format string, evs []types.Event, tableHeader bool) error {
	f, err := output.ParseFormat(format, true)
	if err != nil {
		return err
	}
	switch f {
	case output.FormatJSON:
		return output.JSON(w, evs)
	case output.FormatYAML:
		return output.YAML(w, evs)
	default:
		if len(evs) == 0 && tableHeader {
			fmt.Fprintln(w, "No events.")
			return nil
		}
		return eventsTable(w, evs, tableHeader, output.NewStylist(w))
	}
}

func eventsTable(w io.Writer, evs []types.Event, header bool, sty output.Stylist) error {
	tbl := output.NewTable(w)
	if header {
		tbl.Header("ID", "TIME", "ACTION", "ACTOR", "TARGET", "TENANT", "RESULT")
	}
	for _, e := range evs {
		tbl.Row(
			e.ID,
			output.Age(e.OccurredAt),
			e.Action,
			formatActor(unmarshalActor(e.Actor)),
			formatTarget(unmarshalTarget(e.Target)),
			tenantOrDash(e.TenantID),
			colorizeResult(unmarshalResult(e.Result), sty),
		)
	}
	return tbl.Flush()
}

func tenantOrDash(t string) string {
	if t == "" {
		return "—"
	}
	return t
}

// colorizeResult returns a stylized status string. Empty status
// renders as a dash; unknown statuses render uncolored so we don't
// hide categories the API may add later.
func colorizeResult(r resultSummary, sty output.Stylist) string {
	status := strings.ToLower(strings.TrimSpace(r.Status))
	if status == "" {
		return "—"
	}
	switch status {
	case "ok", "success":
		return sty.Green(r.Status)
	case "error", "fail", "failure":
		return sty.Red(r.Status)
	case "denied", "deny", "rejected":
		return sty.Yellow(r.Status)
	default:
		return r.Status
	}
}

// describeView is the structured shape rendered by `events describe`.
// Nested raw-JSON fields are decoded to generic any so YAML/JSON output
// shows the actual content rather than a base64-looking byte slice.
type describeView struct {
	ID             string `json:"id" yaml:"id"`
	TenantID       string `json:"tenant_id,omitempty" yaml:"tenant_id,omitempty"`
	OccurredAt     string `json:"occurred_at" yaml:"occurred_at"`
	Action         string `json:"action" yaml:"action"`
	IdempotencyKey string `json:"idempotency_key,omitempty" yaml:"idempotency_key,omitempty"`
	Actor          any    `json:"actor,omitempty" yaml:"actor,omitempty"`
	Target         any    `json:"target,omitempty" yaml:"target,omitempty"`
	Metadata       any    `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Origin         any    `json:"origin,omitempty" yaml:"origin,omitempty"`
	Result         any    `json:"result,omitempty" yaml:"result,omitempty"`
	Change         any    `json:"change,omitempty" yaml:"change,omitempty"`
}

// buildDescribeView decodes the raw JSON columns. Decode errors fall
// through to the raw string — the CLI shouldn't hide an event from
// the user just because one column has unexpected contents.
func buildDescribeView(e types.Event) describeView {
	return describeView{
		ID:             e.ID,
		TenantID:       e.TenantID,
		OccurredAt:     e.OccurredAt.UTC().Format("2006-01-02T15:04:05.000000000Z07:00"),
		Action:         e.Action,
		IdempotencyKey: e.IdempotencyKey,
		Actor:          decodeRawJSON(e.Actor),
		Target:         decodeRawJSON(e.Target),
		Metadata:       decodeRawJSON(e.Metadata),
		Origin:         decodeRawJSON(e.Origin),
		Result:         decodeRawJSON(e.Result),
		Change:         decodeRawJSON(e.Change),
	}
}
