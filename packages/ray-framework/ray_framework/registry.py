import json
from dataclasses import dataclass, field
from typing import Any


@dataclass(frozen=True)
class PluginRecord:
    name: str
    enabled: bool
    data_types: list[str]
    depends_on: list[str] = field(default_factory=list)
    namespace: str = ""
    image: str = ""
    entrypoint: str = ""
    supports_distributed: bool = False
    worker_group_name: str = ""
    ray_resource_name: str = ""
    ready_replicas: int = 0


def load_registry_from_json(payload: str) -> list[PluginRecord]:
    document = json.loads(payload)
    return [_record_from_json(item) for item in document.get("plugins", [])]


def _record_from_json(item: dict[str, Any]) -> PluginRecord:
    return PluginRecord(
        name=item["name"],
        namespace=item.get("namespace", ""),
        enabled=item.get("enabled", False),
        image=item.get("image", ""),
        entrypoint=item.get("entrypoint", ""),
        data_types=list(item.get("dataTypes", [])),
        depends_on=list(item.get("dependsOn", [])),
        supports_distributed=item.get("supportsDistributed", False),
        worker_group_name=item.get("workerGroupName", ""),
        ray_resource_name=item.get("rayResourceName", ""),
        ready_replicas=item.get("readyReplicas", 0),
    )
