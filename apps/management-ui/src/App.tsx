import { useEffect, useState } from 'react';
import { fetchPlugins, patchPlugin } from './api';
import { collectDataTypes, filterPlugins, summarizePlugins } from './pluginFilters';
import type { PluginFilter, PluginRecord } from './types';
import './styles.css';

const initialFilter: PluginFilter = {
  query: '',
  enabled: 'all',
  dataType: 'all',
};

export function App() {
  const [plugins, setPlugins] = useState<PluginRecord[]>([]);
  const [filter, setFilter] = useState<PluginFilter>(initialFilter);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [updating, setUpdating] = useState<string | null>(null);

  async function loadPlugins() {
    setLoading(true);
    setError(null);
    try {
      setPlugins(await fetchPlugins());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load plugins');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadPlugins();
  }, []);

  const visiblePlugins = filterPlugins(plugins, filter);
  const summary = summarizePlugins(plugins);
  const dataTypes = collectDataTypes(plugins);

  async function togglePlugin(plugin: PluginRecord) {
    setUpdating(plugin.name);
    setError(null);
    const nextEnabled = !plugin.enabled;
    setPlugins((current) => current.map((item) => (item.name === plugin.name ? { ...item, enabled: nextEnabled } : item)));
    try {
      await patchPlugin(plugin.name, { enabled: nextEnabled });
      await loadPlugins();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update plugin');
      setPlugins((current) => current.map((item) => (item.name === plugin.name ? plugin : item)));
    } finally {
      setUpdating(null);
    }
  }

  return (
    <main className="shell">
      <section className="hero" aria-labelledby="page-title">
        <div>
          <p className="eyebrow">Ray Plugin Provisioner</p>
          <h1 id="page-title">Plugin fleet control</h1>
          <p className="lede">
            Track enabled plugins, dependency edges, worker group routing, and rollout readiness without touching the Ray cluster by hand.
          </p>
        </div>
        <button className="refresh" onClick={() => void loadPlugins()} disabled={loading}>
          {loading ? 'Refreshing' : 'Refresh registry'}
        </button>
      </section>

      <section className="metrics" aria-label="Plugin summary">
        <Metric label="Total plugins" value={summary.total} />
        <Metric label="Enabled" value={summary.enabled} tone="good" />
        <Metric label="Disabled" value={summary.disabled} tone="muted" />
        <Metric label="Ready replicas" value={summary.readyReplicas} tone="accent" />
      </section>

      <section className="panel" aria-label="Plugin filters">
        <label>
          <span>Search</span>
          <input
            value={filter.query}
            placeholder="plugin, image, dependency"
            onChange={(event) => setFilter({ ...filter, query: event.target.value })}
          />
        </label>
        <label>
          <span>Status</span>
          <select value={filter.enabled} onChange={(event) => setFilter({ ...filter, enabled: event.target.value as PluginFilter['enabled'] })}>
            <option value="all">All</option>
            <option value="enabled">Enabled</option>
            <option value="disabled">Disabled</option>
          </select>
        </label>
        <label>
          <span>Data type</span>
          <select value={filter.dataType} onChange={(event) => setFilter({ ...filter, dataType: event.target.value })}>
            <option value="all">All types</option>
            {dataTypes.map((dataType) => (
              <option key={dataType} value={dataType}>
                {dataType}
              </option>
            ))}
          </select>
        </label>
      </section>

      {error ? <div className="error">{error}</div> : null}

      <section className="plugin-grid" aria-live="polite">
        {visiblePlugins.map((plugin) => (
          <article className="plugin-card" key={plugin.name}>
            <div className="plugin-card__header">
              <div>
                <p className="resource">{plugin.workerGroupName || 'worker group pending'}</p>
                <h2>{plugin.name}</h2>
              </div>
              <span className={plugin.enabled ? 'status status--enabled' : 'status status--disabled'}>
                {plugin.enabled ? 'Enabled' : 'Disabled'}
              </span>
            </div>

            <dl className="facts">
              <div>
                <dt>Ready replicas</dt>
                <dd>{plugin.readyReplicas}</dd>
              </div>
              <div>
                <dt>Distributed</dt>
                <dd>{plugin.supportsDistributed ? 'Yes' : 'No'}</dd>
              </div>
            </dl>

            <div className="tags" aria-label={`${plugin.name} data types`}>
              {plugin.dataTypes.length ? plugin.dataTypes.map((dataType) => <span key={dataType}>{dataType}</span>) : <span>No data types</span>}
            </div>

            <p className="image" title={plugin.image}>{plugin.image}</p>
            <p className="entrypoint">{plugin.entrypoint}</p>

            <div className="dependencies">
              <span>Depends on</span>
              <strong>{plugin.dependsOn.length ? plugin.dependsOn.join(', ') : 'nothing'}</strong>
            </div>

            <button className="toggle" disabled={updating === plugin.name} onClick={() => void togglePlugin(plugin)}>
              {plugin.enabled ? 'Disable plugin' : 'Enable plugin'}
            </button>
          </article>
        ))}
      </section>

      {!loading && visiblePlugins.length === 0 ? <div className="empty">No plugins match the current filters.</div> : null}
    </main>
  );
}

function Metric({ label, value, tone = 'default' }: { label: string; value: number; tone?: 'default' | 'good' | 'muted' | 'accent' }) {
  return (
    <div className={`metric metric--${tone}`}>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
