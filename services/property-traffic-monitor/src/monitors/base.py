from __future__ import annotations

import abc
from dataclasses import dataclass, field
from typing import Any


@dataclass
class MonitorResult:
    """Result from a single monitor check."""

    monitor_name: str
    value: float  # 0.0 - 1.0, normalized signal strength
    source: str = "live"  # live | error | cached
    raw_data: dict[str, Any] = field(default_factory=dict)


class BaseMonitor(abc.ABC):
    """Abstract base class for all traffic monitors."""

    name: str = "base"

    @abc.abstractmethod
    async def check(self, address: str, borough: str) -> MonitorResult:
        """Run the monitor check for a given address.

        Returns a MonitorResult with value 0.0-1.0 and raw source data.
        """
        ...
