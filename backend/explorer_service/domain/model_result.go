package domain

type ModelResult struct {
	IsFraud    bool   `json:"is_fraud"`
	Verdict    string `json:"verdict"`
	FlagReason string `json:"flag_reason"`
	RiskScore  int    `json:"risk_score"`
	Data       string `json:"data"` // JSON string of original TransactionInput
}
