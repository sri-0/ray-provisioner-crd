export type PluginRecord = {
  name: string;
  namespace: string;
  enabled: boolean;
  image: string;
  entrypoint: string;
  dataTypes: string[];
  dependsOn: string[];
  supportsDistributed: boolean;
  workerGroupName: string;
  rayResourceName: string;
  readyReplicas: number;
};

export type PluginRegistryResponse = {
  plugins: PluginRecord[];
};

export type PluginFilter = {
  query: string;
  enabled: 'all' | 'enabled' | 'disabled';
  dataType: string;
};
