import { describe, expect, it } from 'vitest';
import { filterPlugins, summarizePlugins } from './pluginFilters';
import type { PluginRecord } from './types';

const plugins: PluginRecord[] = [
  {
    name: 'csv-analyzer',
    namespace: 'ray-system',
    enabled: true,
    image: 'registry/plugins/csv:v1',
    entrypoint: 'plugins.csv.Plugin',
    dataTypes: ['text/csv'],
    dependsOn: [],
    supportsDistributed: true,
    workerGroupName: 'plugin-csv-analyzer',
    rayResourceName: 'plugin-csv-analyzer',
    readyReplicas: 2,
  },
  {
    name: 'pdf-parser',
    namespace: 'ray-system',
    enabled: false,
    image: 'registry/plugins/pdf:v1',
    entrypoint: 'plugins.pdf.Plugin',
    dataTypes: ['application/pdf'],
    dependsOn: ['ocr-normalizer'],
    supportsDistributed: false,
    workerGroupName: 'plugin-pdf-parser',
    rayResourceName: 'plugin-pdf-parser',
    readyReplicas: 0,
  },
];

describe('filterPlugins', () => {
  it('filters by enabled state, query, and data type', () => {
    const result = filterPlugins(plugins, {
      query: 'csv',
      enabled: 'enabled',
      dataType: 'text/csv',
    });

    expect(result.map((plugin) => plugin.name)).toEqual(['csv-analyzer']);
  });
});

describe('summarizePlugins', () => {
  it('summarizes total, enabled, disabled, and ready replicas', () => {
    expect(summarizePlugins(plugins)).toEqual({
      total: 2,
      enabled: 1,
      disabled: 1,
      readyReplicas: 2,
    });
  });
});
