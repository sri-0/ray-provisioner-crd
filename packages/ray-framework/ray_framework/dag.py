from ray_framework.registry import PluginRecord


class MissingDependencyError(ValueError):
    pass


class DependencyCycleError(ValueError):
    pass


def build_execution_layers(plugins: list[PluginRecord]) -> list[list[PluginRecord]]:
    enabled_plugins = {plugin.name: plugin for plugin in plugins if plugin.enabled}

    for plugin in enabled_plugins.values():
        missing = [dependency for dependency in plugin.depends_on if dependency not in enabled_plugins]
        if missing:
            raise MissingDependencyError(f"plugin {plugin.name!r} depends on missing plugins: {', '.join(missing)}")

    remaining = set(enabled_plugins)
    completed: set[str] = set()
    layers: list[list[PluginRecord]] = []

    while remaining:
        ready_names = sorted(
            name
            for name in remaining
            if all(dependency in completed for dependency in enabled_plugins[name].depends_on)
        )
        if not ready_names:
            raise DependencyCycleError("plugin dependency graph contains a cycle")

        layers.append([enabled_plugins[name] for name in ready_names])
        completed.update(ready_names)
        remaining.difference_update(ready_names)

    return layers
