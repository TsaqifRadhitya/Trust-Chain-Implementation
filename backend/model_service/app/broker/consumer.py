"""
AMQP consumer — berjalan sebagai asyncio background task.
Consume dari predict_requests, publish hasil ke reply_to queue.
"""
import asyncio
import json

import aio_pika
import aio_pika.abc

from app.config import RABBITMQ_URL, PREDICT_REQ_QUEUE
from app.ml.predictor import run_prediction, PredictionError


async def _connect_with_retry(
    url: str,
    retries: int = 15,
    delay: float = 3.0,
) -> aio_pika.abc.AbstractRobustConnection:
    """Coba connect ke RabbitMQ dengan retry eksponensial sederhana."""
    for attempt in range(1, retries + 1):
        try:
            conn = await aio_pika.connect_robust(url)
            print(f"[AMQP] Terhubung ke RabbitMQ (percobaan {attempt})")
            return conn
        except Exception as exc:
            wait = delay * attempt
            print(f"[AMQP] Gagal connect percobaan {attempt}/{retries}: {exc}. Retry dalam {wait:.0f}s...")
            await asyncio.sleep(wait)
    raise RuntimeError(f"Tidak dapat terhubung ke RabbitMQ setelah {retries} percobaan")


async def _process_message(
    message: aio_pika.abc.AbstractIncomingMessage,
    channel: aio_pika.abc.AbstractChannel,
) -> None:
    """Proses satu pesan: prediksi → publish hasil ke reply_to."""
    async with message.process(requeue=False):
        try:
            body   = json.loads(message.body.decode())
            params = body.pop("_params", {})

            result = run_prediction(
                body,
                volume_sensitivity=params.get("volume_sensitivity", 50),
                geo_threshold     =params.get("geo_threshold", 50),
                velocity_limit    =params.get("velocity_limit", 50),
            )
            print(
                f"[AMQP] ✓ Prediksi selesai — "
                f"is_fraud={result['is_fraud']}, risk={result['risk_score']}"
            )
        except PredictionError as exc:
            print(f"[AMQP] ⚠️  PredictionError: {exc}")
            result = {"error": str(exc)}
        except Exception as exc:
            print(f"[AMQP] ❌ Unexpected error: {exc}")
            result = {"error": f"Internal error: {exc}"}

        # Balas ke reply_to queue jika ada (RPC pattern)
        if message.reply_to:
            await channel.default_exchange.publish(
                aio_pika.Message(
                    body=json.dumps(result).encode(),
                    correlation_id=message.correlation_id,
                    content_type="application/json",
                ),
                routing_key=message.reply_to,
            )


async def start_consumer() -> None:
    """
    Entry point untuk background task AMQP consumer.
    Dipanggil dari lifespan FastAPI — berjalan selamanya sampai dibatalkan.
    """
    connection = await _connect_with_retry(RABBITMQ_URL)
    channel    = await connection.channel()
    await channel.set_qos(prefetch_count=1)

    queue = await channel.declare_queue(PREDICT_REQ_QUEUE, durable=True)
    print(f"[AMQP] Consumer aktif, mendengarkan queue '{PREDICT_REQ_QUEUE}'")

    try:
        async with queue.iterator() as q_iter:
            async for message in q_iter:
                await _process_message(message, channel)
    except asyncio.CancelledError:
        print("[AMQP] Consumer dihentikan (CancelledError)")
    finally:
        await connection.close()
        print("[AMQP] Koneksi RabbitMQ ditutup")
