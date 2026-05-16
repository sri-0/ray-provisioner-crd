from abc import ABC, abstractmethod
from typing import Any


class BasePlugin(ABC):
    @abstractmethod
    def run(self, s3_bucket: str, s3_key: str, upstream_results: tuple[Any, ...]) -> Any:
        """Run plugin work and return a Ray-serializable result."""
