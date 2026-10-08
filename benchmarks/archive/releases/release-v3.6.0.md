# easySFTP benchmark: release v3.6.0

| Field | Value |
|---|---|
| Kind | release (official reference) |
| Version | `v3.6.0` |
| Recorded | 2026-08-21T12:29:08Z |
| Commit | `3bb56035efaa5b9389bf2316034c07e2d8c2eb81` |
| Workflow run | https://github.com/eiserv/easySFTP/actions/runs/32482071049 |
| Raw data | [release-v3.6.0.json](release-v3.6.0.json) |
| Flat export | [release-v3.6.0.csv](release-v3.6.0.csv) |

## easySFTP benchmark

| Setting | Value |
|---|---|
| Candidate | `v3.6.0 (3bb5603)` |
| Baseline | `none` |
| Repeats per scenario | 3 |
| Runner | Linux 7.0.0-29-generic, 10 cpu |
| Link profiles | the real line |
| Settings | easySFTP defaults (no advanced.* overrides): concurrency 4, request_concurrency 16, retries 2, timeout 30s, mode overlay |

### The link

| Profile | When | RTT p50 | RTT p90 | Handshake | Control 1 stream | Control N streams | Host load |
|---|---|---|---|---|---|---|---|
| baseline | start | 13.27 ms | 14.17 ms | 420.82 ms | 0.37 MiB/s | 0.97 MiB/s | n/a |
| baseline | end | 13.23 ms | 13.61 ms | 358.47 ms | 0.36 MiB/s | 0.87 MiB/s | n/a |

No link shaping was requested: every profile here is the real line.

The control measurement uses `x/crypto/ssh` and `pkg/sftp` directly, never easySFTP's uploader. It separates "the line is slow" from "easySFTP is slow", and a single-stream control close to a scenario's own MiB/s means the run was network bound, where a code delta says nothing.

### Throughput

| Scenario | Build | Profile | Files | Size | Median | Min | Max | MAD | MiB/s | files/s | Retries | Errors | Failed runs | Delta |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| small | candidate | baseline | 300 | 1.2 MiB | 5459 ms | 5199 ms | 5553 ms | 94 ms | 0.21 | 54.96 | 0 | 0 | 0 | - |
| mixed | candidate | baseline | 56 | 11.6 MiB | 18170 ms | 17286 ms | 18738 ms | 568 ms | 0.64 | 3.08 | 0 | 0 | 0 | - |
| large | candidate | baseline | 2 | 32 MiB | 52043 ms | 50054 ms | 56085 ms | 1989 ms | 0.61 | 0.04 | 0 | 0 | 0 | - |

Delta compares each build's median against the `candidate` build **on the same link profile**; negative is faster. MAD is the median absolute deviation of the repeats: a delta smaller than it is inside this host's own noise.

### Resources (median per run)

| Scenario | Build | Profile | User CPU | Sys CPU | CPU % | Peak RSS | Go allocs | GCs | GC pause | Peak goroutines | Net sent |
|---|---|---|---|---|---|---|---|---|---|---|---|
| small | candidate | baseline | 198.76 ms | 241.26 ms | 8.21% | 12.6 MiB | 11.9 MiB | 6 | 2.209854 ms | 141 | 1.4 MiB |
| mixed | candidate | baseline | 207.2 ms | 235.94 ms | 2.44% | 11.3 MiB | 4 MiB | 2 | 1.040325 ms | 110 | 11.7 MiB |
| large | candidate | baseline | 449.04 ms | 604.58 ms | 2.02% | 7.7 MiB | 1.2 MiB | 0 | 0 ms | 26 | 32.1 MiB |

### Where the time goes

Phases are wall clock and add up to roughly the run's duration. Operation totals are **cumulative across parallel workers** and are normally larger than the phase they belong to; read them for their share and their per-call cost, never as wall clock.

<details><summary><code>small</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 4707.35 ms |
| candidate | baseline | connect | 412.82 ms |
| candidate | baseline | cleanup | 114.54 ms |
| candidate | baseline | create_dirs | 60.62 ms |
| candidate | baseline | sweep_stale_temps | 54.32 ms |
| candidate | baseline | local_scan | 5.2 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 300 | 282299.2 ms | 941 ms | 881.28 ms | 1897.42 ms | 2684.86 ms | 2835.79 ms |
| candidate | baseline | sftp_write | 300 | 61499.65 ms | 205 ms | 119.91 ms | 432.54 ms | 856.34 ms | 1003.18 ms |
| candidate | baseline | sftp_open | 300 | 25025.94 ms | 83.42 ms | 24.38 ms | 244.96 ms | 784.71 ms | 831.15 ms |
| candidate | baseline | sftp_chmod | 300 | 21226.75 ms | 70.76 ms | 52.31 ms | 114.23 ms | 828.42 ms | 836.03 ms |
| candidate | baseline | sftp_rename | 300 | 14430 ms | 48.1 ms | 30.65 ms | 73.92 ms | 363.28 ms | 753.16 ms |
| candidate | baseline | sftp_readdir | 9 | 484.22 ms | 53.8 ms | 54.05 ms | 54.08 ms | 54.08 ms | 54.08 ms |
| candidate | baseline | sftp_mkdirall | 8 | 402.95 ms | 50.37 ms | 51.18 ms | 60.35 ms | 60.35 ms | 60.35 ms |
| candidate | baseline | ssh_connect | 1 | 373.46 ms | 373.46 ms | 373.46 ms | 373.46 ms | 373.46 ms | 373.46 ms |
| candidate | baseline | sftp_realpath | 3 | 40.08 ms | 13.36 ms | 13.14 ms | 13.62 ms | 13.62 ms | 13.62 ms |

</details>

<details><summary><code>mixed</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 17556.4 ms |
| candidate | baseline | connect | 406.5 ms |
| candidate | baseline | cleanup | 81.76 ms |
| candidate | baseline | create_dirs | 56.57 ms |
| candidate | baseline | sweep_stale_temps | 55.31 ms |
| candidate | baseline | local_scan | 1.24 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 56 | 241733.58 ms | 4316.67 ms | 3297.53 ms | 7775.94 ms | 17555.13 ms | 17555.13 ms |
| candidate | baseline | sftp_write | 56 | 112389.73 ms | 2006.96 ms | 832.5 ms | 5490.85 ms | 15433.83 ms | 15433.83 ms |
| candidate | baseline | sftp_chmod | 56 | 11869.7 ms | 211.96 ms | 124.99 ms | 446.98 ms | 866.56 ms | 866.56 ms |
| candidate | baseline | sftp_rename | 56 | 9923.86 ms | 177.21 ms | 131.46 ms | 455.71 ms | 719.4 ms | 719.4 ms |
| candidate | baseline | sftp_open | 56 | 1127.06 ms | 20.13 ms | 18.02 ms | 25.64 ms | 45.23 ms | 45.23 ms |
| candidate | baseline | sftp_readdir | 9 | 483.98 ms | 53.78 ms | 53.63 ms | 55.01 ms | 55.01 ms | 55.01 ms |
| candidate | baseline | sftp_mkdirall | 8 | 399.61 ms | 49.95 ms | 50.67 ms | 56.31 ms | 56.31 ms | 56.31 ms |
| candidate | baseline | ssh_connect | 1 | 366.83 ms | 366.83 ms | 366.83 ms | 366.83 ms | 366.83 ms | 366.83 ms |
| candidate | baseline | sftp_realpath | 3 | 39.65 ms | 13.22 ms | 13.01 ms | 13.64 ms | 13.64 ms | 13.64 ms |

</details>

<details><summary><code>large</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 51498.61 ms |
| candidate | baseline | connect | 420.43 ms |
| candidate | baseline | sweep_stale_temps | 54.41 ms |
| candidate | baseline | create_dirs | 43.16 ms |
| candidate | baseline | cleanup | 27.19 ms |
| candidate | baseline | local_scan | 0.34 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 2 | 102467.93 ms | 51233.97 ms | 51498.44 ms | 51498.44 ms | 51498.44 ms | 51498.44 ms |
| candidate | baseline | sftp_write | 2 | 101631.8 ms | 50815.9 ms | 51056.05 ms | 51056.05 ms | 51056.05 ms | 51056.05 ms |
| candidate | baseline | ssh_connect | 1 | 380.95 ms | 380.95 ms | 380.95 ms | 380.95 ms | 380.95 ms | 380.95 ms |
| candidate | baseline | sftp_readdir | 3 | 162.94 ms | 54.31 ms | 54.31 ms | 54.36 ms | 54.36 ms | 54.36 ms |
| candidate | baseline | sftp_mkdirall | 2 | 84.91 ms | 42.46 ms | 43.05 ms | 43.05 ms | 43.05 ms | 43.05 ms |
| candidate | baseline | sftp_realpath | 3 | 39.78 ms | 13.26 ms | 13.24 ms | 13.38 ms | 13.38 ms | 13.38 ms |
| candidate | baseline | sftp_open | 2 | 29.82 ms | 14.91 ms | 14.99 ms | 14.99 ms | 14.99 ms | 14.99 ms |
| candidate | baseline | sftp_rename | 2 | 29.73 ms | 14.87 ms | 15.08 ms | 15.08 ms | 15.08 ms | 15.08 ms |
| candidate | baseline | sftp_chmod | 2 | 26.83 ms | 13.42 ms | 13.62 ms | 13.62 ms | 13.62 ms | 13.62 ms |

</details>

### Delete sweeps

The pre-clean before every measured run wipes the tree the previous repeat left behind, which makes it a pure delete sweep. It costs no extra time (it has always run) and its numbers never enter the upload tables above. Sweeps that found an empty directory are not counted.

| Scenario | Build | Profile | Sweeps | Files deleted | Median | files/s | remote_scan | delete_sweep |
|---|---|---|---|---|---|---|---|---|
| large | candidate | baseline | 2 | 2 | 563 ms | 3.55 | 105.47 ms | 32.59 ms |
| mixed | candidate | baseline | 2 | 56 | 770 ms | 72.73 | 112.25 ms | 190.46 ms |
| small | candidate | baseline | 2 | 300 | 1175 ms | 255.32 | 121.58 ms | 583.26 ms |

| Scenario | Build | Profile | Operation | Count | Cumulative | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| large | candidate | baseline | ssh_connect | 1 | 356.77 ms | 356.77 ms | 356.77 ms | 356.77 ms | 356.77 ms |
| large | candidate | baseline | sftp_readdir | 4 | 210.61 ms | 52.67 ms | 52.91 ms | 52.91 ms | 52.91 ms |
| large | candidate | baseline | sftp_remove | 2 | 30.94 ms | 16.16 ms | 16.16 ms | 16.16 ms | 16.16 ms |
| large | candidate | baseline | sftp_rmdir | 2 | 30.76 ms | 16.05 ms | 16.05 ms | 16.05 ms | 16.05 ms |
| mixed | candidate | baseline | sftp_remove | 56 | 5020.94 ms | 92.12 ms | 145.38 ms | 155 ms | 155 ms |
| mixed | candidate | baseline | sftp_readdir | 10 | 556.34 ms | 57.42 ms | 57.47 ms | 57.47 ms | 57.47 ms |
| mixed | candidate | baseline | ssh_connect | 1 | 381.77 ms | 381.77 ms | 381.77 ms | 381.77 ms | 381.77 ms |
| mixed | candidate | baseline | sftp_rmdir | 8 | 194.82 ms | 25.1 ms | 34.2 ms | 34.2 ms | 34.2 ms |
| small | candidate | baseline | sftp_remove | 300 | 31526.24 ms | 104.56 ms | 130.5 ms | 152.53 ms | 155 ms |
| small | candidate | baseline | sftp_readdir | 10 | 613.6 ms | 61.83 ms | 72.39 ms | 72.39 ms | 72.39 ms |
| small | candidate | baseline | ssh_connect | 1 | 369.52 ms | 369.52 ms | 369.52 ms | 369.52 ms | 369.52 ms |
| small | candidate | baseline | sftp_rmdir | 8 | 183.4 ms | 24.38 ms | 29.48 ms | 29.48 ms | 29.48 ms |

Data only: these numbers set no threshold and fail no build. Collected to evaluate the single-connection ceiling discussed in issue #158 and to show where a run spends its time.
