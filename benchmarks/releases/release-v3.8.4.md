# easySFTP benchmark: release v3.8.4

| Field | Value |
|---|---|
| Kind | release (official reference) |
| Version | `v3.8.4` |
| Recorded | 2026-10-08T07:04:44Z |
| Commit | `4a4034c7782d2ef940e77a70a3dfd36db765d577` |
| Workflow run | https://github.com/eiserv/easySFTP/actions/runs/37741199491 |
| Raw data | [release-v3.8.4.json](release-v3.8.4.json) |
| Flat export | [release-v3.8.4.csv](release-v3.8.4.csv) |

## easySFTP benchmark

| Setting | Value |
|---|---|
| Candidate | `v3.8.4 (4a4034c)` |
| Baseline | `none` |
| Repeats per scenario | 3 |
| Runner | Linux 7.0.0-30-generic, 10 cpu |
| Link profiles | the real line |
| Settings | easySFTP defaults (no advanced.* overrides): connections auto, concurrency auto, request_concurrency auto, retries 2, timeout 30s, mode overlay |

### The link

| Profile | When | RTT p50 | RTT p90 | Handshake | Control 1 stream | Control N streams | Host load |
|---|---|---|---|---|---|---|---|
| baseline | start | 12.93 ms | 13.48 ms | 388.62 ms | 0.39 MiB/s | 0.95 MiB/s | n/a |
| baseline | end | 13.15 ms | 13.4 ms | 381.22 ms | 0.35 MiB/s | 0.95 MiB/s | n/a |

No link shaping was requested: every profile here is the real line.

The control measurement uses `x/crypto/ssh` and `pkg/sftp` directly, never easySFTP's uploader. It separates "the line is slow" from "easySFTP is slow", and a single-stream control close to a scenario's own MiB/s means the run was network bound, where a code delta says nothing.

### Throughput

| Scenario | Build | Profile | Files | Size | Median | Min | Max | MAD | MiB/s | files/s | Retries | Errors | Failed runs | Delta |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| small | candidate | baseline | 300 | 1.2 MiB | 3703 ms | 3648 ms | 3854 ms | 55 ms | 0.32 | 81.02 | 0 | 0 | 0 | - |
| mixed | candidate | baseline | 56 | 11.6 MiB | 11854 ms | 11187 ms | 12279 ms | 425 ms | 0.98 | 4.72 | 0 | 0 | 0 | - |
| large | candidate | baseline | 2 | 32 MiB | 50261 ms | 47587 ms | 50483 ms | 222 ms | 0.64 | 0.04 | 0 | 0 | 0 | - |

Delta compares each build's median against the `candidate` build **on the same link profile**; negative is faster. MAD is the median absolute deviation of the repeats: a delta smaller than it is inside this host's own noise.

### Resources (median per run)

| Scenario | Build | Profile | User CPU | Sys CPU | CPU % | Peak RSS | Go allocs | GCs | GC pause | Peak goroutines | Net sent |
|---|---|---|---|---|---|---|---|---|---|---|---|
| small | candidate | baseline | 218.28 ms | 245.42 ms | 12.51% | 14 MiB | 12.4 MiB | 5 | 2.436353 ms | 578 | 1.4 MiB |
| mixed | candidate | baseline | 246.07 ms | 235.95 ms | 3.99% | 14 MiB | 5.4 MiB | 3 | 1.994044 ms | 1131 | 11.7 MiB |
| large | candidate | baseline | 454.37 ms | 453.92 ms | 1.83% | 8.8 MiB | 1.3 MiB | 0 | 0 ms | 159 | 32.1 MiB |

### Where the time goes

Phases are wall clock and add up to roughly the run's duration. Operation totals are **cumulative across parallel workers** and are normally larger than the phase they belong to; read them for their share and their per-call cost, never as wall clock.

<details><summary><code>small</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 3108.1 ms |
| candidate | baseline | connect | 418.76 ms |
| candidate | baseline | create_dirs | 59.82 ms |
| candidate | baseline | sweep_stale_temps | 54.92 ms |
| candidate | baseline | cleanup | 53.67 ms |
| candidate | baseline | local_scan | 5.29 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 300 | 182525.84 ms | 608.42 ms | 299.79 ms | 1669.71 ms | 2526.99 ms | 2530.29 ms |
| candidate | baseline | sftp_write | 300 | 65738.68 ms | 219.13 ms | 102.39 ms | 706.99 ms | 775.78 ms | 776.83 ms |
| candidate | baseline | sftp_open | 300 | 22100.2 ms | 73.67 ms | 43.92 ms | 137.41 ms | 598.69 ms | 660.37 ms |
| candidate | baseline | sftp_rename | 300 | 20846.69 ms | 69.49 ms | 34.82 ms | 182.51 ms | 466.34 ms | 489.61 ms |
| candidate | baseline | sftp_chmod | 300 | 19846.26 ms | 66.15 ms | 52.02 ms | 123.18 ms | 388.34 ms | 466.32 ms |
| candidate | baseline | sftp_readdir | 9 | 484.35 ms | 53.82 ms | 54.32 ms | 54.6 ms | 54.6 ms | 54.6 ms |
| candidate | baseline | sftp_mkdirall | 8 | 422.42 ms | 52.8 ms | 55.14 ms | 59.68 ms | 59.68 ms | 59.68 ms |
| candidate | baseline | ssh_connect | 1 | 379.51 ms | 379.51 ms | 379.51 ms | 379.51 ms | 379.51 ms | 379.51 ms |
| candidate | baseline | sftp_realpath | 3 | 39.15 ms | 13.05 ms | 12.94 ms | 13.46 ms | 13.46 ms | 13.46 ms |

</details>

<details><summary><code>mixed</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 11183.12 ms |
| candidate | baseline | connect | 432.75 ms |
| candidate | baseline | cleanup | 108.96 ms |
| candidate | baseline | create_dirs | 60.77 ms |
| candidate | baseline | sweep_stale_temps | 54.64 ms |
| candidate | baseline | local_scan | 1.55 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 56 | 237220.11 ms | 4236.07 ms | 4303.76 ms | 6212.04 ms | 11181.89 ms | 11181.89 ms |
| candidate | baseline | sftp_write | 56 | 103221.95 ms | 1843.25 ms | 1250.37 ms | 3387.89 ms | 9532.25 ms | 9532.25 ms |
| candidate | baseline | sftp_chmod | 56 | 32012.64 ms | 571.65 ms | 619.37 ms | 1243.02 ms | 1777.23 ms | 1777.23 ms |
| candidate | baseline | sftp_rename | 56 | 24252.62 ms | 433.08 ms | 409.32 ms | 1012.71 ms | 1532.69 ms | 1532.69 ms |
| candidate | baseline | sftp_open | 56 | 1198.25 ms | 21.4 ms | 20.89 ms | 28.76 ms | 34.62 ms | 34.62 ms |
| candidate | baseline | sftp_readdir | 9 | 489.3 ms | 54.37 ms | 54.48 ms | 54.55 ms | 54.55 ms | 54.55 ms |
| candidate | baseline | sftp_mkdirall | 8 | 426.38 ms | 53.3 ms | 54.7 ms | 60.55 ms | 60.55 ms | 60.55 ms |
| candidate | baseline | ssh_connect | 1 | 393.39 ms | 393.39 ms | 393.39 ms | 393.39 ms | 393.39 ms | 393.39 ms |
| candidate | baseline | sftp_realpath | 3 | 39.32 ms | 13.11 ms | 13.11 ms | 13.21 ms | 13.21 ms | 13.21 ms |

</details>

<details><summary><code>large</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 49755.14 ms |
| candidate | baseline | connect | 382.78 ms |
| candidate | baseline | sweep_stale_temps | 52.69 ms |
| candidate | baseline | create_dirs | 42.78 ms |
| candidate | baseline | cleanup | 26.67 ms |
| candidate | baseline | local_scan | 0.26 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 2 | 92577.05 ms | 46288.53 ms | 49754.99 ms | 49754.99 ms | 49754.99 ms | 49754.99 ms |
| candidate | baseline | sftp_write | 2 | 92116.38 ms | 46058.19 ms | 49712.19 ms | 49712.19 ms | 49712.19 ms | 49712.19 ms |
| candidate | baseline | ssh_connect | 1 | 343.66 ms | 343.66 ms | 343.66 ms | 343.66 ms | 343.66 ms | 343.66 ms |
| candidate | baseline | sftp_readdir | 3 | 157.66 ms | 52.55 ms | 52.61 ms | 52.67 ms | 52.67 ms | 52.67 ms |
| candidate | baseline | sftp_mkdirall | 2 | 83.59 ms | 41.8 ms | 42.66 ms | 42.66 ms | 42.66 ms | 42.66 ms |
| candidate | baseline | sftp_realpath | 3 | 39.08 ms | 13.03 ms | 12.89 ms | 13.21 ms | 13.21 ms | 13.21 ms |
| candidate | baseline | sftp_rename | 2 | 32.17 ms | 16.09 ms | 16.83 ms | 16.83 ms | 16.83 ms | 16.83 ms |
| candidate | baseline | sftp_open | 2 | 29.3 ms | 14.65 ms | 14.92 ms | 14.92 ms | 14.92 ms | 14.92 ms |
| candidate | baseline | sftp_chmod | 2 | 26.46 ms | 13.23 ms | 13.38 ms | 13.38 ms | 13.38 ms | 13.38 ms |

</details>

### Delete sweeps

The pre-clean before every measured run wipes the tree the previous repeat left behind, which makes it a pure delete sweep. It costs no extra time (it has always run) and its numbers never enter the upload tables above. Sweeps that found an empty directory are not counted.

| Scenario | Build | Profile | Sweeps | Files deleted | Median | files/s | remote_scan | delete_sweep |
|---|---|---|---|---|---|---|---|---|
| large | candidate | baseline | 2 | 2 | 584 ms | 3.42 | 106.8 ms | 35.87 ms |
| mixed | candidate | baseline | 2 | 56 | 713 ms | 78.54 | 108.56 ms | 181.84 ms |
| small | candidate | baseline | 2 | 300 | 1514 ms | 198.15 | 123.19 ms | 920.73 ms |

| Scenario | Build | Profile | Operation | Count | Cumulative | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| large | candidate | baseline | ssh_connect | 1 | 375.29 ms | 375.29 ms | 375.29 ms | 375.29 ms | 375.29 ms |
| large | candidate | baseline | sftp_readdir | 4 | 212.44 ms | 53.34 ms | 53.68 ms | 53.68 ms | 53.68 ms |
| large | candidate | baseline | sftp_rmdir | 2 | 33.75 ms | 18.17 ms | 18.17 ms | 18.17 ms | 18.17 ms |
| large | candidate | baseline | sftp_remove | 2 | 32.5 ms | 17.43 ms | 17.43 ms | 17.43 ms | 17.43 ms |
| mixed | candidate | baseline | sftp_remove | 56 | 4567.95 ms | 82.85 ms | 126.4 ms | 138.23 ms | 138.23 ms |
| mixed | candidate | baseline | sftp_readdir | 10 | 542.2 ms | 54.46 ms | 56.55 ms | 56.55 ms | 56.55 ms |
| mixed | candidate | baseline | ssh_connect | 1 | 352.04 ms | 352.04 ms | 352.04 ms | 352.04 ms | 352.04 ms |
| mixed | candidate | baseline | sftp_rmdir | 8 | 191.66 ms | 23.96 ms | 33.71 ms | 33.71 ms | 33.71 ms |
| small | candidate | baseline | sftp_remove | 300 | 51333.91 ms | 152.01 ms | 228.41 ms | 273.29 ms | 273.51 ms |
| small | candidate | baseline | sftp_readdir | 10 | 594.8 ms | 57.51 ms | 69.21 ms | 69.21 ms | 69.21 ms |
| small | candidate | baseline | ssh_connect | 1 | 385.71 ms | 385.71 ms | 385.71 ms | 385.71 ms | 385.71 ms |
| small | candidate | baseline | sftp_rmdir | 8 | 214.94 ms | 26.69 ms | 35.03 ms | 35.03 ms | 35.03 ms |

Data only: these numbers set no threshold and fail no build. Collected to evaluate the single-connection ceiling discussed in issue #158 and to show where a run spends its time.
