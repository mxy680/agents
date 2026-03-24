"""Property Traffic Monitor — FastAPI entrypoint."""

from __future__ import annotations

import logging
from contextlib import asynccontextmanager

from apscheduler.schedulers.asyncio import AsyncIOScheduler
from fastapi import FastAPI

from src.api.routes import router as api_router
from src.api.websocket import router as ws_router
from src.config import settings
from src.db import async_session, init_db
from src.scheduler import run_full_scan

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(name)s: %(message)s",
)
logger = logging.getLogger(__name__)

scheduler = AsyncIOScheduler()


async def _scheduled_scan():
    """Background job that runs periodic scans."""
    async with async_session() as db:
        count = await run_full_scan(db)
        logger.info(f"Scheduled scan complete: {count} addresses scanned")


@asynccontextmanager
async def lifespan(app: FastAPI):
    await init_db()
    logger.info("Database initialized")

    scheduler.add_job(
        _scheduled_scan,
        "interval",
        minutes=settings.scan_interval_minutes,
        id="full_scan",
        replace_existing=True,
    )
    scheduler.start()
    logger.info(f"Scheduler started (interval={settings.scan_interval_minutes}m)")

    yield

    scheduler.shutdown()
    logger.info("Scheduler stopped")


app = FastAPI(
    title="Property Traffic Monitor",
    description="Detects research activity on NYC property addresses",
    version="0.1.0",
    lifespan=lifespan,
)

app.include_router(api_router)
app.include_router(ws_router)


@app.get("/health")
async def health():
    return {"status": "ok"}


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("src.main:app", host=settings.host, port=settings.port, reload=True)
