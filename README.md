# Nasne exporter

[nasne](https://www.jp.playstation.com/nasne/) 用 Prometheus exporterです｡

## メトリクス

| メトリクス名 | メトリクスタイプ | ラベル | 説明 |
| --- | --- | --- | --- |
| `nasne_info` | Gauge | `hardware_version` `name` `product_name` `software_version` | nasne の情報 |
| `nasne_hdd_bytes` | Gauge |`format` `id` `name` `product_id` `vendor_id` | ハードディスの最大容量 |
| `nasne_hdd_usage_bytes` | Gauge |`format` `id` `name` `product_id` `vendor_id` | ハードディスの使用容量 |
| `nasne_record_total` | Gauge |`name` | 録画中の件数 |
| `nasne_recorded_title_total` | Gauge |`name` | 録画された件数 |
| `nasne_conflict_total` | Gauge |`name` | ?? |
| `nasne_dtcpip_client_total` | Gauge |`name` | 現在接続されているDTCP-IPのクライアント数 |
