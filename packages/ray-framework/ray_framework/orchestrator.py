from importlib import import_module
from typing import Any

from ray_framework.dag import build_execution_layers
from ray_framework.plugin import BasePlugin
from ray_framework.registry import PluginRecord


def select_plugins_for_data_type(plugins: list[PluginRecord], data_type: str) -> list[PluginRecord]:
    return [plugin for plugin in plugins if plugin.enabled and data_type in plugin.data_types]


def import_plugin(entrypoint: str) -> BasePlugin:
    module_name, _, class_name = entrypoint.rpartition(".")
    if not module_name or not class_name:
        raise ValueError(f"invalid plugin entrypoint: {entrypoint}")
    module = import_module(module_name)
    plugin_cls = getattr(module, class_name)
    plugin = plugin_cls()
    if not isinstance(plugin, BasePlugin):
        raise TypeError(f"{entrypoint} must inherit from BasePlugin")
    return plugin


def plan_execution(plugins: list[PluginRecord], data_type: str) -> list[list[PluginRecord]]:
    return build_execution_layers(select_plugins_for_data_type(plugins, data_type))


def run_local_plugin(entrypoint: str, s3_bucket: str, s3_key: str, upstream_results: tuple[Any, ...]) -> Any:
    return import_plugin(entrypoint).run(s3_bucket, s3_key, upstream_results)
