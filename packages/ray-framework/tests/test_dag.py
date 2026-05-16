import unittest

from ray_framework.dag import DependencyCycleError, MissingDependencyError, build_execution_layers
from ray_framework.registry import PluginRecord


class DagTests(unittest.TestCase):
    def test_builds_parallel_layers_from_dependencies(self):
        plugins = [
            PluginRecord(name="extract", enabled=True, data_types=["text/csv"], depends_on=[]),
            PluginRecord(name="validate", enabled=True, data_types=["text/csv"], depends_on=["extract"]),
            PluginRecord(name="summarize", enabled=True, data_types=["text/csv"], depends_on=["extract"]),
            PluginRecord(name="publish", enabled=True, data_types=["text/csv"], depends_on=["validate", "summarize"]),
        ]

        layers = build_execution_layers(plugins)

        self.assertEqual([[plugin.name for plugin in layer] for layer in layers], [["extract"], ["summarize", "validate"], ["publish"]])

    def test_rejects_missing_dependency(self):
        plugins = [PluginRecord(name="publish", enabled=True, data_types=["text/csv"], depends_on=["missing"])]

        with self.assertRaises(MissingDependencyError):
            build_execution_layers(plugins)

    def test_rejects_cycles(self):
        plugins = [
            PluginRecord(name="a", enabled=True, data_types=["text/csv"], depends_on=["b"]),
            PluginRecord(name="b", enabled=True, data_types=["text/csv"], depends_on=["a"]),
        ]

        with self.assertRaises(DependencyCycleError):
            build_execution_layers(plugins)


if __name__ == "__main__":
    unittest.main()
