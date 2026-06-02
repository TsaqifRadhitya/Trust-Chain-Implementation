from pydantic import BaseModel, Field


class TransactionInput(BaseModel):
    vendor_name:               str   = Field(..., example="Neo Supply International")
    amount_idr:                float = Field(..., example=8_500_000_000)
    hour_of_day:               int   = Field(..., example=2)
    day_of_week:               int   = Field(..., example=6)
    is_weekend:                int   = Field(..., example=1)
    vendor_age_days:           int   = Field(..., example=45)
    vendor_tx_count_30d:       int   = Field(..., example=15)
    amount_vs_vendor_avg:      float = Field(..., example=8.5)
    geographic_deviation:      float = Field(..., example=0.9)
    tx_velocity_1h:            int   = Field(..., example=12)
    tx_velocity_24h:           int   = Field(..., example=20)
    is_round_number:           int   = Field(..., example=0)
    days_since_last_tx_vendor: int   = Field(..., example=0)
    ip_country_match:          int   = Field(..., example=0)
    duplicate_score:           float = Field(..., example=0.6)
    vendor_category:           str   = Field(..., example="Trading")
    department:                str   = Field(..., example="Finance")
    transaction_type:          str   = Field(..., example="Advance Payment")
    payment_method:            str   = Field(..., example="SWIFT")
    approval_level:            str   = Field(..., example="L4")


class PredictionOutput(BaseModel):
    vendor_name:    str
    amount_idr:     float
    risk_score:     int
    if_score:       float
    lstm_prob:      float
    ensemble_score: float
    is_fraud:       bool
    verdict:        str
    flag_reason:    str


class ModelParams(BaseModel):
    """Parameter tuning yang dikirim dari explorer_service."""
    volume_sensitivity: int = 50
    geo_threshold:      int = 50
    velocity_limit:     int = 50
