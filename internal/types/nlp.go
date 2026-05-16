package types

// GenerateNLPFiltersRequest is the JSON body for
// POST /v1/projects/{projectID}/events/nlp. Q is the user's
// natural-language query; the server enforces length on its end.
type GenerateNLPFiltersRequest struct {
	Q string `json:"q"`
}

// GenerateNLPFiltersResponse is what the NLP endpoint returns. DSL is
// empty when the model couldn't translate anything; in that case the
// caller should surface Unsupported / Explanation rather than fall
// through to an unfiltered list.
type GenerateNLPFiltersResponse struct {
	DSL            string   `json:"dsl"`
	Unsupported    []string `json:"unsupported,omitempty"`
	Explanation    string   `json:"explanation,omitempty"`
	Model          string   `json:"model,omitempty"`
	InputTokens    int      `json:"input_tokens,omitempty"`
	OutputTokens   int      `json:"output_tokens,omitempty"`
	CacheHitTokens int      `json:"cache_hit_tokens,omitempty"`
}
