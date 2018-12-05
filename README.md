# Nasne exporter

[nasne](https://www.jp.playstation.com/nasne/) 用 Prometheus exporterです｡

## メトリクス

| メトリクス名 | メトリクスタイプ | ラベル | 説明 |
| --- | --- | --- | --- |
| `nasne_info` | Gauge | `hardware_version` `name` `product_name` `software_version` | nasne の情報 |
| `nasne_hdd_size_bytes` | Gauge |`format` `id` `name` `product_id` `vendor_id` | ハードディスの最大容量 |
| `nasne_hdd_usage_bytes` | Gauge |`format` `id` `name` `product_id` `vendor_id` | ハードディスの使用容量 |
| `nasne_record_total` | Gauge |`name` | 録画中の件数 |
| `nasne_recorded_title_total` | Gauge |`name` | 録画された件数 |
| `nasne_conflict_total` | Gauge |`name` | ?? |
| `nasne_dtcpip_clients_total` | Gauge |`name` | 現在接続されているDTCP-IPのクライアント数 |

## ビルドと実行

以下のソフトウェアに依存しています｡

- (Go compiler)[https://golang.org/dl/]

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
