"""
FastAPI application — wiring semua komponen.
"""
import asyncio
from contextlib import asynccontextmanager

from fastapi import FastAPI

from app.ml.loader import registry
from app.api.routes import router
from app.broker.consumer import start_consumer


@asynccontextmanager
async def lifespan(app: FastAPI):
    """
    Lifecycle handler (menggantikan on_event deprecated).
    
    Startup:
      1. Muat model ML dari disk
      2. Mulai AMQP consumer sebagai background task
    
    Shutdown:
      - Cancel background consumer dengan bersih
    """
    # ── Startup ───────────────────────────────────────────────
    await registry.load()

    consumer_task = asyncio.create_task(start_consumer())
    print("[App] Background AMQP consumer dimulai")

    yield  # aplikasi berjalan di sini

    # ── Shutdown ──────────────────────────────────────────────
    consumer_task.cancel()
    try:
        await consumer_task
    except asyncio.CancelledError:
        pass
    print("[App] Shutdown selesai")


app = FastAPI(
    title="TrustChain AI — Fraud Detection",
    description=(
        "Ensemble model Isolation Forest + LSTM untuk deteksi transaksi fraud. "
        "Mendukung HTTP REST dan AMQP RPC."
    ),
    version="3.0.0",
    lifespan=lifespan,
)

app.include_router(router)
