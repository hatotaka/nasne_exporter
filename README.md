# Nasne exporter
[![Docker Repository on Quay](https://quay.io/repository/hatotaka/nasne_exporter/status "Docker Repository on Quay")](https://quay.io/repository/hatotaka/nasne_exporter)

[nasne](https://www.jp.playstation.com/nasne/) 用 Prometheus exporterです｡

## メトリクス

| メトリクス名 | メトリクスタイプ | ラベル | 説明 |
| --- | --- | --- | --- |
| `nasne_info` | Gauge | `hardware_version` `name` `product_name` `software_version` | nasne 情報 |
| `nasne_hdd_size_bytes` | Gauge | `format` `id` `name` `product_id` `vendor_id` | ハードディスクの容量 |
| `nasne_hdd_usage_bytes` | Gauge | `format` `id` `name` `product_id` `vendor_id` | ハードディスクの使用容量 |
| `nasne_dtcpip_clients` | Gauge | `name` | 接続されているDTCP-IPのクライアント数 |
| `nasne_recordings` | Gauge | `name` | 録画中の件数 |
| `nasne_recorded_titles` | Gauge | `name` | 録画されている件数 |
| `nasne_reserved_titles` | Gauge | `name` | 予約されている件数 |
| `nasne_reserved_conflict_titles` | Gauge | `name` | コンフリクトした録画件数 |
| `nasne_reserved_notfound_titles` | Gauge | `name` | 見つからない録画件数 |
| `nasne_last_collect_time` | Gauge | | 最後にメトリクスを収集した時間 |
| `nasne_collect_duration_seconds` | Histogram | `name` | メトリクス収集にかかった時間 |

## ビルドと実行

以下のソフトウェアに依存しています｡

- [Go compiler](https://golang.org/dl/)

以下の手順でビルドできます｡

```
go get github.com/hatotaka/nasne_exporter
cd ${GOPATH-$HOME/go}/src/github.com/hatotaka/nasne_exporter
make
./nasne_exporter <flags>
```

フラグは以下の方法で確認できます｡

```
./nasne_exporter -h
```
