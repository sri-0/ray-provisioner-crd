import type { PluginFilter, PluginRecord } from './types';

export function filterPlugins(plugins: PluginRecord[], filter: PluginFilter): PluginRecord[] {
  const query = filter.query.trim().toLowerCase();

  return plugins.filter((plugin) => {
    if (filter.enabled === 'enabled' && !plugin.enabled) {
      return false;
    }
    if (filter.enabled === 'disabled' && plugin.enabled) {
      return false;
    }
    if (filter.dataType !== 'all' && !plugin.dataTypes.includes(filter.dataType)) {
      return false;
    }
    if (!query) {
      return true;
    }

    const searchable = [
      plugin.name,
      plugin.image,
      plugin.entrypoint,
      plugin.workerGroupName,
      ...plugin.dataTypes,
      ...plugin.dependsOn,
    ].join(' ').toLowerCase();
    return searchable.includes(query);
  });
}

export function summarizePlugins(plugins: PluginRecord[]) {
  return plugins.reduce(
    (summary, plugin) => ({
      total: summary.total + 1,
      enabled: summary.enabled + (plugin.enabled ? 1 : 0),
      disabled: summary.disabled + (plugin.enabled ? 0 : 1),
      readyReplicas: summary.readyReplicas + plugin.readyReplicas,
    }),
    { total: 0, enabled: 0, disabled: 0, readyReplicas: 0 },
  );
}

export function collectDataTypes(plugins: PluginRecord[]): string[] {
  const dataTypes = new Set<string>();
  for (const plugin of plugins) {
    for (const dataType of plugin.dataTypes) {
      dataTypes.add(dataType);
    }
  }
  return [...dataTypes].sort();
}
