# easySFTP benchmark: release v3.8.5

| Field | Value |
|---|---|
| Kind | release (official reference) |
| Version | `v3.8.5` |
| Recorded | 2026-10-11T15:40:24Z |
| Commit | `16de9d6dd4604340014ce930aecdf43d9bf8d1c8` |
| Workflow run | https://github.com/eiserv/easySFTP/actions/runs/38152225753 |
| Raw data | [release-v3.8.5.json](release-v3.8.5.json) |
| Flat export | [release-v3.8.5.csv](release-v3.8.5.csv) |

## easySFTP benchmark

| Setting | Value |
|---|---|
| Candidate | `v3.8.5 (16de9d6)` |
| Baseline | `none` |
| Repeats per scenario | 3 |
| Runner | Linux 7.0.0-30-generic, 10 cpu |
| Link profiles | the real line |
| Settings | easySFTP defaults (no advanced.* overrides): connections auto, concurrency auto, request_concurrency auto, retries 2, timeout 30s, mode overlay |

### The link

| Profile | When | RTT p50 | RTT p90 | Handshake | Control 1 stream | Control N streams | Host load |
|---|---|---|---|---|---|---|---|
| baseline | start | 13.04 ms | 13.65 ms | 458.13 ms | 0.34 MiB/s | 1.03 MiB/s | n/a |
| baseline | end | 13.37 ms | 13.96 ms | 391.65 ms | 0.39 MiB/s | 1.08 MiB/s | n/a |

No link shaping was requested: every profile here is the real line.

The control measurement uses `x/crypto/ssh` and `pkg/sftp` directly, never easySFTP's uploader. It separates "the line is slow" from "easySFTP is slow", and a single-stream control close to a scenario's own MiB/s means the run was network bound, where a code delta says nothing.

### Throughput

| Scenario | Build | Profile | Files | Size | Median | Min | Max | MAD | MiB/s | files/s | Retries | Errors | Failed runs | Delta |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| small | candidate | baseline | 300 | 1.2 MiB | 4046 ms | 3319 ms | 4311 ms | 265 ms | 0.29 | 74.15 | 0 | 0 | 0 | - |
| mixed | candidate | baseline | 56 | 11.6 MiB | 12454 ms | 11597 ms | 13450 ms | 857 ms | 0.93 | 4.5 | 0 | 0 | 0 | - |
| large | candidate | baseline | 2 | 32 MiB | 51149 ms | 48601 ms | 51710 ms | 561 ms | 0.63 | 0.04 | 0 | 0 | 0 | - |

Delta compares each build's median against the `candidate` build **on the same link profile**; negative is faster. MAD is the median absolute deviation of the repeats: a delta smaller than it is inside this host's own noise.

### Resources (median per run)

| Scenario | Build | Profile | User CPU | Sys CPU | CPU % | Peak RSS | Go allocs | GCs | GC pause | Peak goroutines | Net sent |
|---|---|---|---|---|---|---|---|---|---|---|---|
| small | candidate | baseline | 196.34 ms | 255.38 ms | 11.28% | 14 MiB | 12.3 MiB | 5 | 2.886569 ms | 184 | 1.4 MiB |
| mixed | candidate | baseline | 267.77 ms | 256.04 ms | 3.99% | 14.2 MiB | 5.5 MiB | 3 | 1.595436 ms | 1325 | 11.7 MiB |
| large | candidate | baseline | 450.87 ms | 503.66 ms | 1.87% | 9 MiB | 1.3 MiB | 0 | 0 ms | 159 | 32.1 MiB |

### Where the time goes

Phases are wall clock and add up to roughly the run's duration. Operation totals are **cumulative across parallel workers** and are normally larger than the phase they belong to; read them for their share and their per-call cost, never as wall clock.

<details><summary><code>small</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 3455.16 ms |
| candidate | baseline | connect | 425.9 ms |
| candidate | baseline | sweep_stale_temps | 59.46 ms |
| candidate | baseline | create_dirs | 57.1 ms |
| candidate | baseline | cleanup | 56.53 ms |
| candidate | baseline | local_scan | 5.1 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 300 | 191103.35 ms | 637.01 ms | 446.27 ms | 1481.53 ms | 2087.79 ms | 2095.02 ms |
| candidate | baseline | sftp_write | 300 | 70094.36 ms | 233.65 ms | 124.6 ms | 538.55 ms | 957.42 ms | 1291.85 ms |
| candidate | baseline | sftp_rename | 300 | 27742.95 ms | 92.48 ms | 38.33 ms | 206.26 ms | 646.76 ms | 648.1 ms |
| candidate | baseline | sftp_open | 300 | 24758.4 ms | 82.53 ms | 43.8 ms | 109.92 ms | 799.55 ms | 923.88 ms |
| candidate | baseline | sftp_chmod | 300 | 22599.1 ms | 75.33 ms | 46.67 ms | 124.58 ms | 495.57 ms | 852.17 ms |
| candidate | baseline | sftp_readdir | 9 | 506.38 ms | 56.26 ms | 56.77 ms | 59.2 ms | 59.2 ms | 59.2 ms |
| candidate | baseline | sftp_mkdirall | 8 | 402.21 ms | 50.28 ms | 51.76 ms | 56.93 ms | 56.93 ms | 56.93 ms |
| candidate | baseline | ssh_connect | 1 | 383.82 ms | 383.82 ms | 383.82 ms | 383.82 ms | 383.82 ms | 383.82 ms |
| candidate | baseline | sftp_realpath | 3 | 39.18 ms | 13.06 ms | 13.07 ms | 13.14 ms | 13.14 ms | 13.14 ms |

</details>

<details><summary><code>mixed</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 11753.26 ms |
| candidate | baseline | connect | 404.34 ms |
| candidate | baseline | cleanup | 110.86 ms |
| candidate | baseline | create_dirs | 67.27 ms |
| candidate | baseline | sweep_stale_temps | 56.16 ms |
| candidate | baseline | local_scan | 1.18 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 56 | 255010.94 ms | 4553.77 ms | 4713.52 ms | 6606.9 ms | 11751.17 ms | 11751.17 ms |
| candidate | baseline | sftp_write | 56 | 110663.9 ms | 1976.14 ms | 1149.21 ms | 3809.14 ms | 9319.54 ms | 9319.54 ms |
| candidate | baseline | sftp_chmod | 56 | 28037.22 ms | 500.66 ms | 509.79 ms | 1041.3 ms | 2091.41 ms | 2091.41 ms |
| candidate | baseline | sftp_rename | 56 | 23924.27 ms | 427.22 ms | 557.92 ms | 847.61 ms | 1138.47 ms | 1138.47 ms |
| candidate | baseline | sftp_open | 56 | 1211.51 ms | 21.63 ms | 20.67 ms | 27.05 ms | 36.67 ms | 36.67 ms |
| candidate | baseline | sftp_readdir | 9 | 490.52 ms | 54.5 ms | 54.67 ms | 55.08 ms | 55.08 ms | 55.08 ms |
| candidate | baseline | sftp_mkdirall | 8 | 480.29 ms | 60.04 ms | 61.83 ms | 66.43 ms | 66.43 ms | 66.43 ms |
| candidate | baseline | ssh_connect | 1 | 365.54 ms | 365.54 ms | 365.54 ms | 365.54 ms | 365.54 ms | 365.54 ms |
| candidate | baseline | sftp_realpath | 3 | 38.84 ms | 12.95 ms | 12.94 ms | 13.01 ms | 13.01 ms | 13.01 ms |

</details>

<details><summary><code>large</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 50608.56 ms |
| candidate | baseline | connect | 417.93 ms |
| candidate | baseline | sweep_stale_temps | 52.38 ms |
| candidate | baseline | create_dirs | 44.13 ms |
| candidate | baseline | cleanup | 27.28 ms |
| candidate | baseline | local_scan | 0.25 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 2 | 96818.63 ms | 48409.32 ms | 50608.36 ms | 50608.36 ms | 50608.36 ms | 50608.36 ms |
| candidate | baseline | sftp_write | 2 | 96279.15 ms | 48139.58 ms | 50164.61 ms | 50164.61 ms | 50164.61 ms | 50164.61 ms |
| candidate | baseline | ssh_connect | 1 | 378.75 ms | 378.75 ms | 378.75 ms | 378.75 ms | 378.75 ms | 378.75 ms |
| candidate | baseline | sftp_readdir | 3 | 156 ms | 52 ms | 51.96 ms | 52.12 ms | 52.12 ms | 52.12 ms |
| candidate | baseline | sftp_mkdirall | 2 | 86.17 ms | 43.09 ms | 44.09 ms | 44.09 ms | 44.09 ms | 44.09 ms |
| candidate | baseline | sftp_chmod | 2 | 75.13 ms | 37.57 ms | 62.22 ms | 62.22 ms | 62.22 ms | 62.22 ms |
| candidate | baseline | sftp_realpath | 3 | 39.15 ms | 13.05 ms | 13 ms | 13.2 ms | 13.2 ms | 13.2 ms |
| candidate | baseline | sftp_rename | 2 | 31.76 ms | 15.88 ms | 16.52 ms | 16.52 ms | 16.52 ms | 16.52 ms |
| candidate | baseline | sftp_open | 2 | 30.57 ms | 15.29 ms | 15.38 ms | 15.38 ms | 15.38 ms | 15.38 ms |

</details>

### Delete sweeps

The pre-clean before every measured run wipes the tree the previous repeat left behind, which makes it a pure delete sweep. It costs no extra time (it has always run) and its numbers never enter the upload tables above. Sweeps that found an empty directory are not counted.

| Scenario | Build | Profile | Sweeps | Files deleted | Median | files/s | remote_scan | delete_sweep |
|---|---|---|---|---|---|---|---|---|
| large | candidate | baseline | 2 | 2 | 590 ms | 3.39 | 104.59 ms | 34.05 ms |
| mixed | candidate | baseline | 2 | 56 | 799 ms | 70.09 | 115.22 ms | 245.77 ms |
| small | candidate | baseline | 2 | 300 | 1507 ms | 199.07 | 124.19 ms | 951.4 ms |

| Scenario | Build | Profile | Operation | Count | Cumulative | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| large | candidate | baseline | ssh_connect | 1 | 365.64 ms | 365.64 ms | 365.64 ms | 365.64 ms | 365.64 ms |
| large | candidate | baseline | sftp_readdir | 4 | 208.28 ms | 51.95 ms | 52.51 ms | 52.51 ms | 52.51 ms |
| large | candidate | baseline | sftp_rmdir | 2 | 32.49 ms | 17.1 ms | 17.1 ms | 17.1 ms | 17.1 ms |
| large | candidate | baseline | sftp_remove | 2 | 31.37 ms | 16.83 ms | 16.83 ms | 16.83 ms | 16.83 ms |
| mixed | candidate | baseline | sftp_remove | 56 | 6747.95 ms | 129.94 ms | 196.9 ms | 209.54 ms | 209.54 ms |
| mixed | candidate | baseline | sftp_readdir | 10 | 546.15 ms | 54.33 ms | 58.1 ms | 58.1 ms | 58.1 ms |
| mixed | candidate | baseline | ssh_connect | 1 | 367.11 ms | 367.11 ms | 367.11 ms | 367.11 ms | 367.11 ms |
| mixed | candidate | baseline | sftp_rmdir | 8 | 173.47 ms | 20.56 ms | 30.15 ms | 30.15 ms | 30.15 ms |
| small | candidate | baseline | sftp_remove | 300 | 51414.91 ms | 138.36 ms | 266.68 ms | 362.21 ms | 362.67 ms |
| small | candidate | baseline | sftp_readdir | 10 | 603.92 ms | 58.62 ms | 69.05 ms | 69.05 ms | 69.05 ms |
| small | candidate | baseline | ssh_connect | 1 | 351.51 ms | 351.51 ms | 351.51 ms | 351.51 ms | 351.51 ms |
| small | candidate | baseline | sftp_rmdir | 8 | 190.98 ms | 25.16 ms | 31.63 ms | 31.63 ms | 31.63 ms |

Data only: these numbers set no threshold and fail no build. Collected to evaluate the single-connection ceiling discussed in issue #158 and to show where a run spends its time.
