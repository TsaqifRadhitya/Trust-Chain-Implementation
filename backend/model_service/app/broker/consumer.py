"""
AMQP consumer — berjalan sebagai asyncio background task.
Consume dari predict_requests, lakukan prediksi, lalu publish hasil ke predict_responses.
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
            print(f"[AMQP Consumer] Terhubung ke RabbitMQ (percobaan {attempt})")
            return conn
        except Exception as exc:
            wait = delay * attempt
            print(f"[AMQP Consumer] Gagal connect percobaan {attempt}/{retries}: {exc}. Retry dalam {wait:.0f}s...")
            await asyncio.sleep(wait)
    raise RuntimeError(f"Tidak dapat terhubung ke RabbitMQ setelah {retries} percobaan")


async def _process_message(
    message: aio_pika.abc.AbstractIncomingMessage,
    channel: aio_pika.abc.AbstractChannel,
) -> None:
    """Proses satu pesan: prediksi → publish hasil ke predict_responses dan reply_to (jika ada)."""
    async with message.process(requeue=False):
        tx_hash = None
        try:
            body = json.loads(message.body.decode())
            tx_hash = body.pop("tx_hash", None)
            params = body.pop("_params", {})

            print(f"[AMQP Consumer] Memulai proses prediksi untuk TxHash: {tx_hash} (Vendor: {body.get('vendor_name')})")

            result = run_prediction(
                body,
                volume_sensitivity=params.get("volume_sensitivity", 50),
                geo_threshold     =params.get("geo_threshold", 50),
                velocity_limit    =params.get("velocity_limit", 50),
            )

            result["tx_hash"] = tx_hash
            print(
                f"[AMQP Consumer] ✓ Prediksi selesai untuk TxHash {tx_hash} — "
                f"is_fraud={result['is_fraud']}, risk={result['risk_score']}, verdict={result['verdict']}"
            )
        except PredictionError as exc:
            print(f"[AMQP Consumer] ⚠️  PredictionError untuk TxHash {tx_hash}: {exc}")
            result = {"error": str(exc), "tx_hash": tx_hash}
        except Exception as exc:
            print(f"[AMQP Consumer] ❌ Unexpected error untuk TxHash {tx_hash}: {exc}")
            result = {"error": f"Internal error: {exc}", "tx_hash": tx_hash}

        # 1. Balas ke reply_to queue jika ada (RPC pattern backward compatibility)
        if message.reply_to:
            print(f"[AMQP Consumer] Mengirim balasan RPC ke routing key '{message.reply_to}'")
            await channel.default_exchange.publish(
                aio_pika.Message(
                    body=json.dumps(result).encode(),
                    correlation_id=message.correlation_id,
                    content_type="application/json",
                ),
                routing_key=message.reply_to,
            )

        # 2. Publish hasil ke predict_responses queue untuk ditangkap consumer asinkron explorer_service
        response_queue_name = "predict_responses"
        await channel.declare_queue(response_queue_name, durable=True)
        print(f"[AMQP Consumer] Publishing hasil prediksi ke queue '{response_queue_name}' untuk TxHash: {tx_hash}")
        await channel.default_exchange.publish(
            aio_pika.Message(
                body=json.dumps(result).encode(),
                content_type="application/json",
            ),
            routing_key=response_queue_name,
        )
        print(f"[AMQP Consumer] ✓ Sukses publish hasil untuk TxHash: {tx_hash}")


async def start_consumer() -> None:
    """
    Entry point untuk background task AMQP consumer.
    Dipanggil dari lifespan FastAPI — berjalan selamanya sampai dibatalkan.
    """
    connection = await _connect_with_retry(RABBITMQ_URL)
    channel    = await connection.channel()
    await channel.set_qos(prefetch_count=1)

    queue = await channel.declare_queue(PREDICT_REQ_QUEUE, durable=True)
    print(f"[AMQP Consumer] Consumer aktif, mendengarkan queue '{PREDICT_REQ_QUEUE}'")

    try:
        async with queue.iterator() as q_iter:
            async for message in q_iter:
                await _process_message(message, channel)
    except asyncio.CancelledError:
        print("[AMQP Consumer] Consumer dihentikan (CancelledError)")
    finally:
        await connection.close()
        print("[AMQP Consumer] Koneksi RabbitMQ ditutup")
