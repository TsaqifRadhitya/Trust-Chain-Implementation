"""
Logika prediksi fraud — dipakai oleh HTTP endpoint maupun AMQP consumer.
Tidak ada side effects, tidak tahu tentang FastAPI atau RabbitMQ.
"""
import numpy as np

from app.config import CAT_MAP, WEIGHT_IF, WEIGHT_LSTM
from app.ml.loader import registry


class PredictionError(Exception):
    """Raised jika input tidak valid atau model belum siap."""


def run_prediction(
    payload: dict,
    volume_sensitivity: int = 50,
    geo_threshold: int = 50,
    velocity_limit: int = 50,
) -> dict:
    """
    Jalankan ensemble Isolation Forest + LSTM untuk deteksi fraud.

    Args:
        payload:            Dict field transaksi (sesuai TransactionInput).
        volume_sensitivity: Sensitivitas terhadap anomali volume (0–100).
        geo_threshold:      Sensitivitas terhadap deviasi geografis (0–100).
        velocity_limit:     Sensitivitas terhadap kecepatan transaksi (0–100).

    Returns:
        Dict hasil prediksi (sesuai PredictionOutput).

    Raises:
        PredictionError: Jika model belum siap atau payload tidak valid.
    """
    if not registry.is_ready:
        raise PredictionError("Model ML belum siap, coba lagi beberapa saat")

    try:
        feature_row = _build_feature_row(payload)
    except KeyError as exc:
        raise PredictionError(f"Field wajib tidak ditemukan di payload: {exc}") from exc

    # ── Scoring ──────────────────────────────────────────────────
    X_scaled     = registry.scaler.transform([feature_row])
    iso_score    = -registry.iso_forest.score_samples(X_scaled)[0]
    iso_norm     = float(np.clip((iso_score - 0.3) / 0.4, 0, 1))

    X_lstm       = X_scaled.reshape(1, 1, -1).astype(np.float32)
    lstm_prob    = registry.predict_lstm(X_lstm)

    ensemble     = WEIGHT_IF * iso_norm + WEIGHT_LSTM * lstm_prob

    # ── Dynamic penalties berdasarkan threshold user ──────────────
    vol_penalty  = _volume_penalty(payload, volume_sensitivity)
    geo_penalty  = _geo_penalty(payload, geo_threshold)
    vel_penalty  = _velocity_penalty(payload, velocity_limit)
    ensemble     = min(1.0, ensemble + vol_penalty + geo_penalty + vel_penalty)

    risk_score    = int(ensemble * 100)
    is_fraud      = bool(ensemble >= 0.5)
    flag_reason   = _build_flag_reason(
        is_fraud, payload, vol_penalty, geo_penalty, vel_penalty
    )

    return {
        "vendor_name":    payload.get("vendor_name", ""),
        "amount_idr":     payload["amount_idr"],
        "risk_score":     risk_score,
        "if_score":       round(iso_norm, 3),
        "lstm_prob":      round(lstm_prob, 3),
        "ensemble_score": round(ensemble, 3),
        "is_fraud":       is_fraud,
        "verdict":        "🚨 FRAUD" if is_fraud else "✅ NORMAL",
        "flag_reason":    flag_reason,
    }


# ── Private helpers ───────────────────────────────────────────

def _build_feature_row(p: dict) -> list:
    return [
        p["amount_idr"],
        p["hour_of_day"],
        p["day_of_week"],
        p["is_weekend"],
        p["vendor_age_days"],
        p["vendor_tx_count_30d"],
        p["amount_vs_vendor_avg"],
        p["geographic_deviation"],
        p["tx_velocity_1h"],
        p["tx_velocity_24h"],
        p["is_round_number"],
        p["days_since_last_tx_vendor"],
        p["ip_country_match"],
        p["duplicate_score"],
        CAT_MAP["vendor_category"].get(p.get("vendor_category", ""), 0),
        CAT_MAP["department"].get(p.get("department", ""), 0),
        CAT_MAP["transaction_type"].get(p.get("transaction_type", ""), 0),
        CAT_MAP["payment_method"].get(p.get("payment_method", ""), 0),
        CAT_MAP["approval_level"].get(p.get("approval_level", ""), 0),
    ]


def _volume_penalty(p: dict, sensitivity: int) -> float:
    threshold = max(1.0, (100 - sensitivity) / 10.0)
    if p.get("amount_vs_vendor_avg", 0) > threshold:
        return (sensitivity / 100.0) * 0.20
    return 0.0


def _geo_penalty(p: dict, threshold_cfg: int) -> float:
    threshold = max(0.1, (100 - threshold_cfg) / 100.0)
    if p.get("geographic_deviation", 0) > threshold:
        return (threshold_cfg / 100.0) * 0.20
    return 0.0


def _velocity_penalty(p: dict, limit: int) -> float:
    threshold = max(1.0, (100 - limit) / 5.0)
    if p.get("tx_velocity_1h", 0) > threshold:
        return (limit / 100.0) * 0.20
    return 0.0


def _build_flag_reason(
    is_fraud: bool,
    p: dict,
    vol: float,
    geo: float,
    vel: float,
) -> str:
    if not is_fraud:
        return "Normal"

    reasons: list[str] = []
    if vol > 0 and vol >= geo and vol >= vel:
        reasons.append("Volume Anomaly")
    if geo > 0 and geo >= vol and geo >= vel:
        reasons.append("Geographic Mismatch")
    if vel > 0 and vel >= vol and vel >= geo:
        reasons.append("Velocity Check")

    if not reasons:
        if p.get("duplicate_score", 0) > 0.5:
            reasons.append("Duplicate Detection")
        elif p.get("tx_velocity_1h", 0) > 5:
            reasons.append("Velocity Check")
        elif p.get("amount_vs_vendor_avg", 0) > 2:
            reasons.append("Volume Anomaly")
        elif p.get("geographic_deviation", 0) > 0.5:
            reasons.append("Geographic Mismatch")
        else:
            reasons.append("AI Pattern Anomaly")

    return " & ".join(reasons)
