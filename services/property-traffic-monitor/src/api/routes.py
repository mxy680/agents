from __future__ import annotations

import json

from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from src.db import get_session
from src.models import (
    AddressCreate,
    AddressOut,
    AddressRow,
    AlertOut,
    AlertRow,
    ScoreOut,
    SignalOut,
    SignalRow,
)
from src.signals.aggregator import compute_composite_score

router = APIRouter()


@router.post("/addresses", response_model=AddressOut, status_code=201)
async def add_address(body: AddressCreate, db: AsyncSession = Depends(get_session)):
    row = AddressRow(address=body.address, borough=body.borough)
    db.add(row)
    await db.commit()
    await db.refresh(row)
    return row


@router.get("/addresses", response_model=list[AddressOut])
async def list_addresses(db: AsyncSession = Depends(get_session)):
    result = await db.execute(select(AddressRow))
    return result.scalars().all()


@router.delete("/addresses/{address_id}", status_code=204)
async def delete_address(address_id: int, db: AsyncSession = Depends(get_session)):
    row = await db.get(AddressRow, address_id)
    if not row:
        raise HTTPException(status_code=404, detail="Address not found")
    await db.delete(row)
    await db.commit()


@router.get("/addresses/{address_id}/signals")
async def get_signals(address_id: int, db: AsyncSession = Depends(get_session)):
    row = await db.get(AddressRow, address_id)
    if not row:
        raise HTTPException(status_code=404, detail="Address not found")

    result = await db.execute(
        select(SignalRow)
        .where(SignalRow.address_id == address_id)
        .order_by(SignalRow.scanned_at.desc())
    )
    rows = result.scalars().all()

    # Group by monitor_name, keep latest per monitor
    signals: dict[str, dict] = {}
    for r in rows:
        if r.monitor_name not in signals:
            raw = None
            if r.raw_data:
                try:
                    raw = json.loads(r.raw_data)
                except (json.JSONDecodeError, TypeError):
                    raw = r.raw_data
            signals[r.monitor_name] = {
                "value": r.value,
                "source": r.source,
                "raw_data": raw,
                "scanned_at": r.scanned_at.isoformat() if r.scanned_at else None,
            }

    return signals


@router.get("/addresses/{address_id}/score", response_model=ScoreOut)
async def get_score(address_id: int, db: AsyncSession = Depends(get_session)):
    row = await db.get(AddressRow, address_id)
    if not row:
        raise HTTPException(status_code=404, detail="Address not found")

    result = await db.execute(
        select(SignalRow)
        .where(SignalRow.address_id == address_id)
        .order_by(SignalRow.scanned_at.desc())
    )
    rows = result.scalars().all()

    # Latest per monitor
    latest: dict[str, float] = {}
    for r in rows:
        if r.monitor_name not in latest:
            latest[r.monitor_name] = r.value

    composite = compute_composite_score(latest)
    return ScoreOut(address_id=address_id, composite_score=composite, signals=latest)


@router.get("/alerts", response_model=list[AlertOut])
async def get_alerts(db: AsyncSession = Depends(get_session)):
    result = await db.execute(select(AlertRow).order_by(AlertRow.triggered_at.desc()).limit(50))
    return result.scalars().all()


@router.post("/scan")
async def trigger_scan(db: AsyncSession = Depends(get_session)):
    from src.scheduler import run_full_scan

    result = await run_full_scan(db)
    return {"status": "scan_complete", "addresses_scanned": result}
