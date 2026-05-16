from ray_framework import BasePlugin


class Plugin(BasePlugin):
    def run(self, s3_bucket: str, s3_key: str, upstream_results: tuple[object, ...]) -> dict[str, object]:
        return {
            "plugin": "csv-analyzer",
            "bucket": s3_bucket,
            "key": s3_key,
            "upstreamResults": len(upstream_results),
        }
