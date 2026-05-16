import type { PluginRecord, PluginRegistryResponse } from './types';

const API_BASE = import.meta.env.VITE_API_BASE ?? '';

export async function fetchPlugins(): Promise<PluginRecord[]> {
  const response = await fetch(`${API_BASE}/api/v1/plugins`);
  if (!response.ok) {
    throw new Error(`Failed to load plugins: ${response.status}`);
  }
  const payload = (await response.json()) as PluginRegistryResponse;
  return payload.plugins;
}

export async function patchPlugin(name: string, patch: { enabled?: boolean }): Promise<void> {
  const response = await fetch(`${API_BASE}/api/v1/plugins/${encodeURIComponent(name)}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(patch),
  });
  if (!response.ok) {
    throw new Error(`Failed to update plugin ${name}: ${response.status}`);
  }
}
