from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
import joblib
import numpy as np
import tensorflow as tf
from tensorflow.keras.models import load_model
import os

# ==========================================
# 1. SETUP APLIKASI & LOAD MODEL
# ==========================================
app = FastAPI(
    title="TrustChain AI - Fraud Detection API",
    description="API untuk memprediksi anomali transaksi menggunakan Ensemble (Isolation Forest + LSTM)",
    version="1.0.0"
)

# Global variables untuk model
iso_forest = None
scaler = None
lstm_model = None

# Konfigurasi Bobot Ensemble
WEIGHT_IF = 0.30
WEIGHT_LSTM = 0.70

@app.on_event("startup")
async def load_ml_models():
    global iso_forest, scaler, lstm_model
    try:        
        model_dir = "/app/trustchain_models" 
        
        print(f"Loading models from: {model_dir}")
        iso_forest = joblib.load(os.path.join(model_dir, "isolation_forest.pkl"))
        scaler = joblib.load(os.path.join(model_dir, "scaler.pkl"))
        lstm_model = load_model(os.path.join(model_dir, "lstm_model.keras"))
        
        print("✅ Semua model berhasil dimuat di dalam Docker!")
    except Exception as e:
        print(f"❌ Gagal memuat model di Docker: {e}")

# ==========================================
# 2. DEFINISI SCHEMA INPUT (PYDANTIC)
# ==========================================
class TransactionInput(BaseModel):
    vendor_name: str = Field(..., example="Neo Supply International")
    amount_idr: float = Field(..., example=8500000000)
    hour_of_day: int = Field(..., example=2)
    day_of_week: int = Field(..., example=6)
    is_weekend: int = Field(..., example=1)
    vendor_age_days: int = Field(..., example=45)
    vendor_tx_count_30d: int = Field(..., example=15)
    amount_vs_vendor_avg: float = Field(..., example=8.5)
    geographic_deviation: float = Field(..., example=0.9)
    tx_velocity_1h: int = Field(..., example=12)
    tx_velocity_24h: int = Field(..., example=20)
    is_round_number: int = Field(..., example=0)
    days_since_last_tx_vendor: int = Field(..., example=0)
    ip_country_match: int = Field(..., example=0)
    duplicate_score: float = Field(..., example=0.6)
    vendor_category: str = Field(..., example="Trading")
    department: str = Field(..., example="Finance")
    transaction_type: str = Field(..., example="Advance Payment")
    payment_method: str = Field(..., example="SWIFT")
    approval_level: str = Field(..., example="L4")

class PredictionOutput(BaseModel):
    vendor_name: str
    amount_idr: float
    risk_score: int
    if_score: float
    lstm_prob: float
    ensemble_score: float
    is_fraud: bool
    verdict: str

# ==========================================
# 3. ENDPOINT PREDIKSI
# ==========================================
@app.post("/predict", response_model=PredictionOutput)
async def predict_fraud(tx: TransactionInput):
    if iso_forest is None or scaler is None or lstm_model is None:
        raise HTTPException(status_code=500, detail="Model belum siap/gagal dimuat dari server.")

    # Mapping Categorical sesuai Cell 11
    cat_map = {
        'vendor_category': {'Logistics':0,'Manufacturing':1,'Energy':2, 'Chemicals':3,'Construction':4,'Electronics':5, 'Raw Materials':6,'Engineering':7,'Trading':8},
        'department':       {'Finance':0,'Procurement':1,'Operations':2, 'Engineering':3,'Logistics':4,'HR':5},
        'transaction_type': {'Invoice Payment':0,'Advance Payment':1, 'Reimbursement':2,'Purchase Order':3,'Contract Payment':4},
        'payment_method':   {'Bank Transfer':0,'RTGS':1,'Virtual Account':2, 'SWIFT':3,'Cash':4},
        'approval_level':   {'L1':0,'L2':1,'L3':2,'L4':3},
    }

    try:
        # Susun feature array (HARUS SESUAI URUTAN TRAINING)
        row = [
            tx.amount_idr,
            tx.hour_of_day,
            tx.day_of_week,
            tx.is_weekend,
            tx.vendor_age_days,
            tx.vendor_tx_count_30d,
            tx.amount_vs_vendor_avg,
            tx.geographic_deviation,
            tx.tx_velocity_1h,
            tx.tx_velocity_24h,
            tx.is_round_number,
            tx.days_since_last_tx_vendor,
            tx.ip_country_match,
            tx.duplicate_score,
            cat_map['vendor_category'].get(tx.vendor_category, 0),
            cat_map['department'].get(tx.department, 0),
            cat_map['transaction_type'].get(tx.transaction_type, 0),
            cat_map['payment_method'].get(tx.payment_method, 0),
            cat_map['approval_level'].get(tx.approval_level, 0),
        ]

        # Preprocessing
        X_new = scaler.transform([row])
        
        # 1. Isolation Forest Prediction
        iso_s = -iso_forest.score_samples(X_new)[0]
        iso_norm_val = float(np.clip((iso_s - 0.3) / 0.4, 0, 1))

        # 2. LSTM Prediction (reshape to 1 sample, 1 timestep, N features)
        X_lstm = X_new.reshape(1, 1, -1)
        lstm_p = float(lstm_model.predict(X_lstm, verbose=0)[0][0])

        # 3. Ensemble
        ensemble_s = (WEIGHT_IF * iso_norm_val) + (WEIGHT_LSTM * lstm_p)
        risk_score = int(ensemble_s * 100)
        is_fraud_pred = bool(ensemble_s >= 0.5)

        return PredictionOutput(
            vendor_name=tx.vendor_name,
            amount_idr=tx.amount_idr,
            risk_score=risk_score,
            if_score=round(iso_norm_val, 3),
            lstm_prob=round(lstm_p, 3),
            ensemble_score=round(ensemble_s, 3),
            is_fraud=is_fraud_pred,
            verdict="🚨 FRAUD" if is_fraud_pred else "✅ NORMAL"
        )

    except Exception as e:
        raise HTTPException(status_code=400, detail=f"Terjadi kesalahan saat memproses data: {str(e)}")

# ==========================================
# 4. ENDPOINT KESEHATAN (HEALTH CHECK)
# ==========================================
@app.get("/")
async def root():
    return {"status": "online", "message": "TrustChain AI API is running!"}