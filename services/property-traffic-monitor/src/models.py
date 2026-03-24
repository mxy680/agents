from __future__ import annotations

import datetime
from typing import Any

from pydantic import BaseModel
from sqlalchemy import Column, DateTime, Float, Integer, String, Text, func
from sqlalchemy.orm import DeclarativeBase


# ── SQLAlchemy ORM ──────────────────────────────────────────────


class Base(DeclarativeBase):
    pass


class AddressRow(Base):
    __tablename__ = "addresses"

    id = Column(Integer, primary_key=True, autoincrement=True)
    address = Column(String(500), nullable=False)
    borough = Column(String(100), nullable=False, default="Manhattan")
    created_at = Column(DateTime, server_default=func.now())


class SignalRow(Base):
    __tablename__ = "signals"

    id = Column(Integer, primary_key=True, autoincrement=True)
    address_id = Column(Integer, nullable=False, index=True)
    monitor_name = Column(String(100), nullable=False)
    value = Column(Float, nullable=False, default=0.0)
    source = Column(String(50), nullable=False, default="live")  # live | error | cached
    raw_data = Column(Text, nullable=True)  # JSON blob
    scanned_at = Column(DateTime, server_default=func.now())


class AlertRow(Base):
    __tablename__ = "alerts"

    id = Column(Integer, primary_key=True, autoincrement=True)
    address_id = Column(Integer, nullable=False, index=True)
    composite_score = Column(Float, nullable=False)
    triggered_at = Column(DateTime, server_default=func.now())
    details = Column(Text, nullable=True)


# ── Pydantic schemas ───────────────────────────────────────────


class AddressCreate(BaseModel):
    address: str
    borough: str = "Manhattan"


class AddressOut(BaseModel):
    id: int
    address: str
    borough: str
    created_at: datetime.datetime | None = None

    model_config = {"from_attributes": True}


class SignalOut(BaseModel):
    monitor_name: str
    value: float
    source: str
    raw_data: Any | None = None
    scanned_at: datetime.datetime | None = None

    model_config = {"from_attributes": True}


class ScoreOut(BaseModel):
    address_id: int
    composite_score: float
    signals: dict[str, float]


class AlertOut(BaseModel):
    id: int
    address_id: int
    composite_score: float
    triggered_at: datetime.datetime | None = None
    details: str | None = None

    model_config = {"from_attributes": True}
