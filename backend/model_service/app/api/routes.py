from fastapi import APIRouter, HTTPException, Query
import tensorflow as tf

from app.schemas import TransactionInput, PredictionOutput
from app.ml.loader import registry
from app.ml.predictor import run_prediction, PredictionError

router = APIRouter(tags=["Prediction"])


@router.post("/predict", response_model=PredictionOutput, summary="Deteksi fraud transaksi")
async def predict_fraud(
    tx: TransactionInput,
    volume_sensitivity: int = Query(50, ge=0, le=100, description="Sensitivitas anomali volume"),
    geo_threshold:      int = Query(50, ge=0, le=100, description="Sensitivitas deviasi geografis"),
    velocity_limit:     int = Query(50, ge=0, le=100, description="Sensitivitas kecepatan transaksi"),
):
    """
    Jalankan prediksi fraud ensemble (Isolation Forest + LSTM).
    Dapat dipanggil langsung via HTTP maupun melalui AMQP consumer.
    """
    try:
        result = run_prediction(
            tx.model_dump(),
            volume_sensitivity=volume_sensitivity,
            geo_threshold=geo_threshold,
            velocity_limit=velocity_limit,
        )
        return PredictionOutput(**result)
    except PredictionError as exc:
        raise HTTPException(status_code=503, detail=str(exc))
    except Exception as exc:
        raise HTTPException(status_code=400, detail=str(exc))


@router.get("/health", summary="Status model ML")
async def health():
    return {
        "status":      "ready" if registry.is_ready else "loading",
        "iso_forest":  registry.iso_forest is not None,
        "scaler":      registry.scaler is not None,
        "lstm_model":  registry.lstm_model is not None,
        "tf_version":  tf.__version__,
    }


@router.get("/", summary="Info service")
async def root():
    return {
        "service":  "TrustChain AI - Fraud Detection",
        "version":  "3.0.0",
        "status":   "online",
        "tf_version": tf.__version__,
    }
