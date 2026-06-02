"""
ML model loader — singleton pattern.

Semua state ML model disimpan di class MLModelRegistry.
Instansi global `registry` diimpor oleh modul lain yang membutuhkan.
"""
import os
import joblib
import numpy as np
import tensorflow as tf

from app.config import MODEL_DIR


def _rebuild_lstm(n_features: int = 19) -> tf.keras.Model:
    """Rebuild arsitektur LSTM identik dengan saat training."""
    model = tf.keras.Sequential([
        tf.keras.layers.LSTM(64, input_shape=(1, n_features), return_sequences=True),
        tf.keras.layers.Dropout(0.3),
        tf.keras.layers.LSTM(32, return_sequences=False),
        tf.keras.layers.Dropout(0.3),
        tf.keras.layers.BatchNormalization(),
        tf.keras.layers.Dense(16, activation="relu"),
        tf.keras.layers.Dropout(0.2),
        tf.keras.layers.Dense(1, activation="sigmoid"),
    ])
    model.compile(optimizer="adam", loss="binary_crossentropy")
    return model


class MLModelRegistry:
    """Menyimpan instance model ML yang sudah dimuat."""

    def __init__(self):
        self.iso_forest  = None
        self.scaler      = None
        self.lstm_model  = None
        self._lstm_is_savedmodel = False

    @property
    def is_ready(self) -> bool:
        return all([self.iso_forest, self.scaler, self.lstm_model])

    async def load(self) -> None:
        """Muat semua model dari disk. Dipanggil satu kali saat startup."""
        print(f"[MLLoader] Memuat model dari: {MODEL_DIR}")
        print(f"[MLLoader] TensorFlow versi: {tf.__version__}")

        self._load_sklearn_models()
        self._load_lstm()

    # ── private helpers ───────────────────────────────────────────

    def _load_sklearn_models(self) -> None:
        try:
            self.iso_forest = joblib.load(os.path.join(MODEL_DIR, "isolation_forest.pkl"))
            self.scaler     = joblib.load(os.path.join(MODEL_DIR, "scaler.pkl"))
            print("[MLLoader] ✅ Isolation Forest & Scaler berhasil dimuat")
        except Exception as exc:
            print(f"[MLLoader] ❌ Gagal muat sklearn models: {exc}")

    def _load_lstm(self) -> None:
        """Coba tiga strategi load LSTM secara berurutan."""
        if self._try_load_lstm_weights():
            return
        if self._try_load_lstm_keras():
            return
        print("[MLLoader] ❌ Semua strategi load LSTM gagal")

    def _try_load_lstm_weights(self) -> bool:
        weights_path = os.path.join(MODEL_DIR, "lstm_weights.weights.h5")
        if not os.path.exists(weights_path):
            return False
        try:
            model = _rebuild_lstm(n_features=19)
            dummy = np.zeros((1, 1, 19), dtype=np.float32)
            model.predict(dummy, verbose=0)  # initialize weights
            model.load_weights(weights_path)
            self.lstm_model = model
            self._lstm_is_savedmodel = False
            print("[MLLoader] ✅ LSTM dimuat via weights file")
            return True
        except Exception as exc:
            print(f"[MLLoader] ⚠️  Load weights gagal: {exc}")
            return False

    def _try_load_lstm_keras(self) -> bool:
        keras_path = os.path.join(MODEL_DIR, "lstm_model.keras")
        if not os.path.exists(keras_path):
            return False
        try:
            with tf.keras.utils.custom_object_scope({}):
                model = tf.keras.models.load_model(
                    keras_path, compile=False, safe_mode=False
                )
            self.lstm_model = model
            self._lstm_is_savedmodel = False
            print("[MLLoader] ✅ LSTM dimuat via .keras")
            return True
        except Exception as exc:
            print(f"[MLLoader] ⚠️  Load .keras gagal: {exc}")
            return False

    def predict_lstm(self, X_input: np.ndarray) -> float:
        """Jalankan inferensi LSTM. Mendukung SavedModel dan Keras biasa."""
        if self._lstm_is_savedmodel:
            infer      = self.lstm_model.signatures.get(
                "serve", list(self.lstm_model.signatures.values())[0]
            )
            input_key  = list(infer.structured_input_signature[1].keys())[0]
            tensor_in  = tf.constant(X_input, dtype=tf.float32)
            result     = infer(**{input_key: tensor_in})
            output_key = list(result.keys())[0]
            return float(result[output_key].numpy()[0][0])
        return float(self.lstm_model.predict(X_input, verbose=0)[0][0])


# ── Singleton yang dipakai seluruh aplikasi ───────────────────
registry = MLModelRegistry()
