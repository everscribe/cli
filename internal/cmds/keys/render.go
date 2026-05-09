package keys

import (
	"fmt"
	"io"

	"github.com/everscribe/cli/internal/output"
	"github.com/everscribe/cli/internal/types"
)

// renderKeys writes a slice of API keys. JSON/YAML emit the bare
// array (no envelope); table is kubectl-style.
func renderKeys(w io.Writer, format string, ks []types.APIKey) error {
	f, err := output.ParseFormat(format, true)
	if err != nil {
		return err
	}
	switch f {
	case output.FormatJSON:
		return output.JSON(w, ks)
	case output.FormatYAML:
		return output.YAML(w, ks)
	default:
		if len(ks) == 0 {
			fmt.Fprintln(w, "No API keys.")
			return nil
		}
		return keysTable(w, ks)
	}
}

func keysTable(w io.Writer, ks []types.APIKey) error {
	tbl := output.NewTable(w)
	tbl.Header("ID", "NAME", "PREFIX", "LAST USED", "AGE")
	for _, k := range ks {
		lastUsed := "never"
		if !k.LastUsedAt.IsZero() {
			lastUsed = output.Age(k.LastUsedAt) + " ago"
		}
		tbl.Row(k.ID, k.Name, k.Prefix, lastUsed, output.Age(k.CreatedAt))
	}
	return tbl.Flush()
}

// renderCreatedKey writes the response from POST /keys. For json/yaml
// the full envelope (including plaintext) is dumped. For table, the
// metadata is printed first, then the plaintext is highlighted with a
// banner — it's only returned this once and the user has to copy it.
func renderCreatedKey(w io.Writer, format string, resp *types.CreateAPIKeyResponse) error {
	f, err := output.ParseFormat(format, true)
	if err != nil {
		return err
	}
	switch f {
	case output.FormatJSON:
		return output.JSON(w, resp)
	case output.FormatYAML:
		return output.YAML(w, resp)
	default:
		if err := keysTable(w, []types.APIKey{resp.Key}); err != nil {
			return err
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Save this token now — it will not be shown again:")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  "+resp.Plaintext)
		return nil
	}
}
