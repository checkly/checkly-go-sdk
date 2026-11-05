package checkly

import "encoding/json"

// CheckIntent describes what a check protects. Automatic AI repair of a
// check uses it to decide which changes keep the check meaningful.
//
// The API requires Goal, of 1 to 2000 characters, and accepts at
// most 20 statements, of 1 to 1000 characters each, for each of
// RequiredOutcomes and MustPreserve.
type CheckIntent struct {
	// Goal is the user or system outcome the check protects.
	Goal string `json:"goal"`
	// RequiredOutcomes are observable outcomes the check must continue to
	// prove.
	RequiredOutcomes []string `json:"requiredOutcomes"`
	// MustPreserve are guardrails an AI repair must not weaken or bypass.
	MustPreserve []string `json:"mustPreserve"`
}

// MarshalJSON sends nil statement lists as empty lists, because the API
// rejects null for them.
func (i CheckIntent) MarshalJSON() ([]byte, error) {
	type alias CheckIntent
	i.RequiredOutcomes = emptyIfNil(i.RequiredOutcomes)
	i.MustPreserve = emptyIfNil(i.MustPreserve)
	return json.Marshal(alias(i))
}
