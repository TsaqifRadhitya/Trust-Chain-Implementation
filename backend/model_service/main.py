from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
import joblib
import numpy as np
import os
import tensorflow as tf

app = FastAPI(
    title="TrustChain AI - Fraud Detection API",
    description="Ensemble Isolation Forest + LSTM",
    version="2.0.0"
)

iso_forest = None
scaler = None
lstm_model = None

WEIGHT_IF = 0.30
WEIGHT_LSTM = 0.70

def rebuild_lstm(n_features: int = 19):
    """Rebuild arsitektur LSTM identik dengan training."""
    model = tf.keras.Sequential([
        tf.keras.layers.LSTM(64, input_shape=(1, n_features), return_sequences=True),
        tf.keras.layers.Dropout(0.3),
        tf.keras.layers.LSTM(32, return_sequences=False),
        tf.keras.layers.Dropout(0.3),
        tf.keras.layers.BatchNormalization(),
        tf.keras.layers.Dense(16, activation='relu'),
        tf.keras.layers.Dropout(0.2),
        tf.keras.layers.Dense(1, activation='sigmoid')
    ])
    model.compile(optimizer='adam', loss='binary_crossentropy')
    return model

@app.on_event("startup")
async def load_ml_models():
    global iso_forest, scaler, lstm_model

    model_dir = "/app/trustchain_models"
    print(f"Loading models from: {model_dir}")
    print(f"TensorFlow version : {tf.__version__}")

    try:
        iso_forest = joblib.load(os.path.join(model_dir, "isolation_forest.pkl"))
        scaler     = joblib.load(os.path.join(model_dir, "scaler.pkl"))
        print("✅ Isolation Forest & Scaler loaded")
    except Exception as e:
        print(f"❌ Gagal load sklearn models: {e}")
        return

    # ── Coba 3 strategi load LSTM ──────────────────────────────

    # Strategi 1: TF SavedModel
    # savedmodel_path = os.path.join(model_dir, "lstm_savedmodel")
    # if os.path.exists(savedmodel_path):
    #     try:
    #         lstm_model = tf.saved_model.load(savedmodel_path)
    #         # Wrap agar bisa dipanggil seperti model biasa
    #         lstm_model._is_savedmodel = True
    #         print("✅ LSTM loaded via TF SavedModel")
    #         return
    #     except Exception as e:
    #         print(f"⚠️  SavedModel gagal: {e}")

    # Strategi 2: Rebuild arsitektur + load weights
    weights_path = os.path.join(model_dir, "lstm_weights.weights.h5")
    if os.path.exists(weights_path):
        try:
            lstm_model = rebuild_lstm(n_features=19)
            # Dummy forward pass untuk initialize weights
            dummy = np.zeros((1, 1, 19), dtype=np.float32)
            lstm_model.predict(dummy, verbose=0)
            lstm_model.load_weights(weights_path)
            lstm_model._is_savedmodel = False
            print("✅ LSTM loaded via weights")
            return
        except Exception as e:
            print(f"⚠️  Load weights gagal: {e}")

    # Strategi 3: .keras dengan custom_object_scope
    keras_path = os.path.join(model_dir, "lstm_model.keras")
    if os.path.exists(keras_path):
        try:
            with tf.keras.utils.custom_object_scope({}):
                lstm_model = tf.keras.models.load_model(
                    keras_path,
                    compile=False,
                    safe_mode=False
                )
            lstm_model._is_savedmodel = False
            print("✅ LSTM loaded via .keras (safe_mode=False)")
            return
        except Exception as e:
            print(f"⚠️  .keras safe_mode=False gagal: {e}")

    print("❌ Semua strategi load LSTM gagal")


def predict_lstm(X_input: np.ndarray) -> float:
    if getattr(lstm_model, '_is_savedmodel', False):
        infer = lstm_model.signatures.get(
            "serve", 
            list(lstm_model.signatures.values())[0]
        )
        input_key = list(infer.structured_input_signature[1].keys())[0]
        tensor_in = tf.constant(X_input, dtype=tf.float32)
        result    = infer(**{input_key: tensor_in})
        output_key = list(result.keys())[0]
        return float(result[output_key].numpy()[0][0])
    else:
        return float(lstm_model.predict(X_input, verbose=0)[0][0])


# ── Schema ─────────────────────────────────────────────────────
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
    flag_reason: str


# ── Endpoint ───────────────────────────────────────────────────
@app.post("/predict", response_model=PredictionOutput)
async def predict_fraud(
    tx: TransactionInput,
    volume_sensitivity: int = 50,
    geo_threshold: int = 50,
    velocity_limit: int = 50
):
    if iso_forest is None or scaler is None or lstm_model is None:
        raise HTTPException(status_code=500, detail="Model belum siap.")

    cat_map = {
        'vendor_category': {'Logistics':0,'Manufacturing':1,'Energy':2,
                            'Chemicals':3,'Construction':4,'Electronics':5,
                            'Raw Materials':6,'Engineering':7,'Trading':8},
        'department':      {'Finance':0,'Procurement':1,'Operations':2,
                            'Engineering':3,'Logistics':4,'HR':5},
        'transaction_type':{'Invoice Payment':0,'Advance Payment':1,
                            'Reimbursement':2,'Purchase Order':3,'Contract Payment':4},
        'payment_method':  {'Bank Transfer':0,'RTGS':1,'Virtual Account':2,
                            'SWIFT':3,'Cash':4},
        'approval_level':  {'L1':0,'L2':1,'L3':2,'L4':3},
    }

    try:
        row = [
            tx.amount_idr, tx.hour_of_day, tx.day_of_week, tx.is_weekend,
            tx.vendor_age_days, tx.vendor_tx_count_30d, tx.amount_vs_vendor_avg,
            tx.geographic_deviation, tx.tx_velocity_1h, tx.tx_velocity_24h,
            tx.is_round_number, tx.days_since_last_tx_vendor, tx.ip_country_match,
            tx.duplicate_score,
            cat_map['vendor_category'].get(tx.vendor_category, 0),
            cat_map['department'].get(tx.department, 0),
            cat_map['transaction_type'].get(tx.transaction_type, 0),
            cat_map['payment_method'].get(tx.payment_method, 0),
            cat_map['approval_level'].get(tx.approval_level, 0),
        ]

        X_new        = scaler.transform([row])
        iso_s        = -iso_forest.score_samples(X_new)[0]
        iso_norm_val = float(np.clip((iso_s - 0.3) / 0.4, 0, 1))

        X_lstm   = X_new.reshape(1, 1, -1).astype(np.float32)
        lstm_p   = predict_lstm(X_lstm)

        ensemble_s    = WEIGHT_IF * iso_norm_val + WEIGHT_LSTM * lstm_p

        # Dynamic adjustments based on user configurations
        volume_penalty = 0.0
        if tx.amount_vs_vendor_avg > max(1.0, (100 - volume_sensitivity) / 10.0):
            volume_penalty = (volume_sensitivity / 100.0) * 0.20
            
        geo_penalty = 0.0
        if tx.geographic_deviation > max(0.1, (100 - geo_threshold) / 100.0):
            geo_penalty = (geo_threshold / 100.0) * 0.20
            
        velocity_penalty = 0.0
        if tx.tx_velocity_1h > max(1.0, (100 - velocity_limit) / 5.0):
            velocity_penalty = (velocity_limit / 100.0) * 0.20
            
        ensemble_s = min(1.0, ensemble_s + volume_penalty + geo_penalty + velocity_penalty)

        risk_score    = int(ensemble_s * 100)
        is_fraud_pred = bool(ensemble_s >= 0.5)

        flag_reason = "Normal"
        if is_fraud_pred:
            reasons = []
            if volume_penalty >= geo_penalty and volume_penalty >= velocity_penalty and volume_penalty > 0:
                reasons.append("Volume Anomaly")
            if geo_penalty >= volume_penalty and geo_penalty >= velocity_penalty and geo_penalty > 0:
                reasons.append("Geographic Mismatch")
            if velocity_penalty >= volume_penalty and velocity_penalty >= geo_penalty and velocity_penalty > 0:
                reasons.append("Velocity Check")
                
            if not reasons:
                if tx.duplicate_score > 0.5:
                    reasons.append("Duplicate Detection")
                elif tx.tx_velocity_1h > 5:
                    reasons.append("Velocity Check")
                elif tx.amount_vs_vendor_avg > 2:
                    reasons.append("Volume Anomaly")
                elif tx.geographic_deviation > 0.5:
                    reasons.append("Geographic Mismatch")
                else:
                    reasons.append("AI Pattern Anomaly")
                    
            flag_reason = " & ".join(reasons)

        return PredictionOutput(
            vendor_name=tx.vendor_name,
            amount_idr=tx.amount_idr,
            risk_score=risk_score,
            if_score=round(iso_norm_val, 3),
            lstm_prob=round(lstm_p, 3),
            ensemble_score=round(ensemble_s, 3),
            is_fraud=is_fraud_pred,
            verdict="🚨 FRAUD" if is_fraud_pred else "✅ NORMAL",
            flag_reason=flag_reason
        )

    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/")
async def root():
    return {
        "status": "online",
        "message": "TrustChain AI API is running!",
        "tf_version": tf.__version__
    }

@app.get("/health")
async def health():
    return {
        "iso_forest": iso_forest is not None,
        "scaler": scaler is not None,
        "lstm_model": lstm_model is not None,
    }