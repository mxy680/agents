"""Scheduler for periodic monitoring scans."""

from __future__ import annotations

import json
import logging
from datetime import datetime, timezone

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from src.config import settings
from src.models import AddressRow, AlertRow, SignalRow
from src.monitors.base import BaseMonitor, MonitorResult
from src.monitors.crux import CruxMonitor
from src.monitors.google_autocomplete import GoogleAutocompleteMonitor
from src.monitors.google_cache import GoogleCacheMonitor
from src.monitors.google_trends import GoogleTrendsMonitor
from src.monitors.page_change import PageChangeMonitor
from src.monitors.social_signals import SocialSignalsMonitor
from src.monitors.wayback import WaybackMonitor
from src.signals.aggregator import compute_composite_score

logger = logging.getLogger(__name__)

# All available monitors
MONITORS: list[BaseMonitor] = [
    WaybackMonitor(),
    GoogleAutocompleteMonitor(),
    GoogleTrendsMonitor(),
    CruxMonitor(),
    PageChangeMonitor(),
    SocialSignalsMonitor(),
    GoogleCacheMonitor(),
]


async def scan_address(address_row: AddressRow, db: AsyncSession) -> dict[str, float]:
    """Run all monitors for a single address and store results."""
    signals: dict[str, float] = {}

    for monitor in MONITORS:
        try:
            logger.info(f"Running {monitor.name} for {address_row.address}")
            result = await monitor.check(address_row.address, address_row.borough)

            # Store signal in DB
            signal_row = SignalRow(
                address_id=address_row.id,
                monitor_name=result.monitor_name,
                value=result.value,
                source=result.source,
                raw_data=json.dumps(result.raw_data) if result.raw_data else None,
                scanned_at=datetime.now(timezone.utc),
            )
            db.add(signal_row)
            signals[result.monitor_name] = result.value

            logger.info(
                f"  {monitor.name}: value={result.value:.3f} source={result.source}"
            )

        except Exception as e:
            logger.error(f"  {monitor.name} failed: {e}")
            signal_row = SignalRow(
                address_id=address_row.id,
                monitor_name=monitor.name,
                value=0.0,
                source="error",
                raw_data=json.dumps({"error": str(e)}),
                scanned_at=datetime.now(timezone.utc),
            )
            db.add(signal_row)
            signals[monitor.name] = 0.0

    # Compute composite score
    composite = compute_composite_score(signals)

    # Create alert if above threshold
    if composite > settings.alert_threshold:
        alert = AlertRow(
            address_id=address_row.id,
            composite_score=composite,
            triggered_at=datetime.now(timezone.utc),
            details=json.dumps(signals),
        )
        db.add(alert)
        logger.warning(
            f"ALERT: {address_row.address} composite={composite:.3f} > {settings.alert_threshold}"
        )

    await db.commit()
    return signals


async def run_full_scan(db: AsyncSession) -> int:
    """Scan all monitored addresses. Returns count of addresses scanned."""
    result = await db.execute(select(AddressRow))
    addresses = result.scalars().all()

    count = 0
    for addr in addresses:
        try:
            await scan_address(addr, db)
            count += 1
        except Exception as e:
            logger.error(f"Failed to scan {addr.address}: {e}")

    return count
