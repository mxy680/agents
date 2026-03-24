"""Tests for API endpoints."""

import pytest
from fastapi.testclient import TestClient

from src.db import init_db, engine
from src.main import app
from src.models import Base


@pytest.fixture(autouse=True)
async def setup_db():
    """Create fresh tables for each test."""
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)
        await conn.run_sync(Base.metadata.create_all)
    yield
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)


@pytest.fixture
def client():
    # Use TestClient for sync tests (simpler)
    with TestClient(app) as c:
        yield c


def test_health(client):
    resp = client.get("/health")
    assert resp.status_code == 200
    assert resp.json() == {"status": "ok"}


def test_add_address(client):
    resp = client.post("/addresses", json={"address": "350 5th Ave", "borough": "Manhattan"})
    assert resp.status_code == 201
    data = resp.json()
    assert data["address"] == "350 5th Ave"
    assert data["borough"] == "Manhattan"
    assert "id" in data


def test_list_addresses(client):
    client.post("/addresses", json={"address": "350 5th Ave", "borough": "Manhattan"})
    client.post("/addresses", json={"address": "1 Vanderbilt Ave", "borough": "Manhattan"})
    resp = client.get("/addresses")
    assert resp.status_code == 200
    data = resp.json()
    assert len(data) == 2


def test_delete_address(client):
    resp = client.post("/addresses", json={"address": "350 5th Ave", "borough": "Manhattan"})
    addr_id = resp.json()["id"]
    resp = client.delete(f"/addresses/{addr_id}")
    assert resp.status_code == 204
    resp = client.get("/addresses")
    assert len(resp.json()) == 0


def test_delete_nonexistent(client):
    resp = client.delete("/addresses/999")
    assert resp.status_code == 404


def test_get_signals_empty(client):
    resp = client.post("/addresses", json={"address": "350 5th Ave", "borough": "Manhattan"})
    addr_id = resp.json()["id"]
    resp = client.get(f"/addresses/{addr_id}/signals")
    assert resp.status_code == 200
    assert resp.json() == {}


def test_get_score_empty(client):
    resp = client.post("/addresses", json={"address": "350 5th Ave", "borough": "Manhattan"})
    addr_id = resp.json()["id"]
    resp = client.get(f"/addresses/{addr_id}/score")
    assert resp.status_code == 200
    data = resp.json()
    assert data["composite_score"] == 0.0


def test_get_alerts_empty(client):
    resp = client.get("/alerts")
    assert resp.status_code == 200
    assert resp.json() == []
