import os

# ── RabbitMQ ──────────────────────────────────────────────────
RABBITMQ_URL      = os.getenv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
PREDICT_REQ_QUEUE = "predict_requests"

# ── Model ─────────────────────────────────────────────────────
MODEL_DIR   = os.getenv("MODEL_DIR", "/app/trustchain_models")
WEIGHT_IF   = 0.30   # bobot Isolation Forest
WEIGHT_LSTM = 0.70   # bobot LSTM

# ── Category mappings (digunakan di predictor) ─────────────────
CAT_MAP = {
    "vendor_category": {
        "Logistics": 0, "Manufacturing": 1, "Energy": 2,
        "Chemicals": 3, "Construction": 4, "Electronics": 5,
        "Raw Materials": 6, "Engineering": 7, "Trading": 8,
    },
    "department": {
        "Finance": 0, "Procurement": 1, "Operations": 2,
        "Engineering": 3, "Logistics": 4, "HR": 5,
    },
    "transaction_type": {
        "Invoice Payment": 0, "Advance Payment": 1,
        "Reimbursement": 2, "Purchase Order": 3, "Contract Payment": 4,
    },
    "payment_method": {
        "Bank Transfer": 0, "RTGS": 1, "Virtual Account": 2,
        "SWIFT": 3, "Cash": 4,
    },
    "approval_level": {"L1": 0, "L2": 1, "L3": 2, "L4": 3},
}
