# easySFTP benchmark: release v3.5.0

| Field | Value |
|---|---|
| Kind | release (official reference) |
| Version | `v3.5.0` |
| Recorded | 2026-08-17T22:36:13Z |
| Commit | `c057a1d08c398224b1f54519b775a5c639b1fbf7` |
| Workflow run | https://github.com/eiserv/easySFTP/actions/runs/32076711732 |
| Raw data | [release-v3.5.0.json](release-v3.5.0.json) |
| Flat export | [release-v3.5.0.csv](release-v3.5.0.csv) |

## easySFTP benchmark

| Setting | Value |
|---|---|
| Candidate | `v3.5.0 (c057a1d)` |
| Baseline | `none` |
| Repeats per scenario | 3 |
| Runner | Linux 7.0.0-29-generic, 10 cpu |
| Link profiles | the real line |
| Settings | easySFTP defaults (no advanced.* overrides): concurrency 4, request_concurrency 16, retries 2, timeout 30s, mode overlay |

### The link

| Profile | When | RTT p50 | RTT p90 | Handshake | Control 1 stream | Control N streams | Host load |
|---|---|---|---|---|---|---|---|
| baseline | start | 13.13 ms | 13.29 ms | 344.47 ms | 0.39 MiB/s | 1.1 MiB/s | n/a |
| baseline | end | 13.19 ms | 13.49 ms | 341.62 ms | 0.39 MiB/s | 0.99 MiB/s | n/a |

No link shaping was requested: every profile here is the real line.

The control measurement uses `x/crypto/ssh` and `pkg/sftp` directly, never easySFTP's uploader. It separates "the line is slow" from "easySFTP is slow", and a single-stream control close to a scenario's own MiB/s means the run was network bound, where a code delta says nothing.

### Throughput

| Scenario | Build | Profile | Files | Size | Median | Min | Max | MAD | MiB/s | files/s | Retries | Errors | Failed runs | Delta |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| small | candidate | baseline | 300 | 1.2 MiB | 14009 ms | 13055 ms | 14599 ms | 590 ms | 0.08 | 21.41 | 0 | 0 | 0 | - |
| mixed | candidate | baseline | 56 | 11.6 MiB | 32406 ms | 29209 ms | 32856 ms | 450 ms | 0.36 | 1.73 | 0 | 0 | 0 | - |
| large | candidate | baseline | 2 | 32 MiB | 86488 ms | 76020 ms | 89316 ms | 2828 ms | 0.37 | 0.02 | 0 | 0 | 0 | - |

Delta compares each build's median against the `candidate` build **on the same link profile**; negative is faster. MAD is the median absolute deviation of the repeats: a delta smaller than it is inside this host's own noise.

### Resources (median per run)

| Scenario | Build | Profile | User CPU | Sys CPU | CPU % | Peak RSS | Go allocs | GCs | GC pause | Peak goroutines | Net sent |
|---|---|---|---|---|---|---|---|---|---|---|---|
| small | candidate | baseline | 226.52 ms | 359.2 ms | 4.41% | 10.8 MiB | 11.3 MiB | 3 | 0.488913 ms | 18 | 1.4 MiB |
| mixed | candidate | baseline | 201.44 ms | 242.64 ms | 1.44% | 8.7 MiB | 2.8 MiB | 1 | 0.240906 ms | 18 | 11.7 MiB |
| large | candidate | baseline | 510.12 ms | 499.1 ms | 1.17% | 7.7 MiB | 1 MiB | 0 | 0 ms | 16 | 32.1 MiB |

### Where the time goes

Phases are wall clock and add up to roughly the run's duration. Operation totals are **cumulative across parallel workers** and are normally larger than the phase they belong to; read them for their share and their per-call cost, never as wall clock.

<details><summary><code>small</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 12836.84 ms |
| candidate | baseline | sweep_stale_temps | 479.6 ms |
| candidate | baseline | create_dirs | 347.46 ms |
| candidate | baseline | connect | 340.1 ms |
| candidate | baseline | cleanup | 13.33 ms |
| candidate | baseline | local_scan | 4.79 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 300 | 51190.76 ms | 170.64 ms | 162.43 ms | 223.51 ms | 297.57 ms | 303.61 ms |
| candidate | baseline | sftp_write | 300 | 22128.74 ms | 73.76 ms | 72.2 ms | 122.47 ms | 143.47 ms | 144.54 ms |
| candidate | baseline | sftp_chmod | 300 | 10964.64 ms | 36.55 ms | 28.36 ms | 67.06 ms | 73.37 ms | 77.46 ms |
| candidate | baseline | sftp_open | 300 | 9536.09 ms | 31.79 ms | 14.79 ms | 66.88 ms | 75.69 ms | 79.93 ms |
| candidate | baseline | sftp_rename | 300 | 8209.1 ms | 27.36 ms | 16.02 ms | 60.73 ms | 78.96 ms | 120.47 ms |
| candidate | baseline | sftp_readdir | 9 | 479.28 ms | 53.25 ms | 53.28 ms | 54.07 ms | 54.07 ms | 54.07 ms |
| candidate | baseline | sftp_mkdirall | 8 | 347.38 ms | 43.42 ms | 42.16 ms | 53.02 ms | 53.02 ms | 53.02 ms |
| candidate | baseline | ssh_connect | 1 | 340.1 ms | 340.1 ms | 340.1 ms | 340.1 ms | 340.1 ms | 340.1 ms |

</details>

<details><summary><code>mixed</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 31272.16 ms |
| candidate | baseline | sweep_stale_temps | 470.67 ms |
| candidate | baseline | create_dirs | 325.27 ms |
| candidate | baseline | connect | 323.44 ms |
| candidate | baseline | cleanup | 13.73 ms |
| candidate | baseline | local_scan | 1.51 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 56 | 109107.72 ms | 1948.35 ms | 881.71 ms | 2882.38 ms | 17344.87 ms | 17344.87 ms |
| candidate | baseline | sftp_write | 56 | 91094.06 ms | 1626.68 ms | 444.09 ms | 2316.94 ms | 16840.6 ms | 16840.6 ms |
| candidate | baseline | sftp_open | 56 | 7196.65 ms | 128.51 ms | 134.82 ms | 238.38 ms | 335.43 ms | 335.43 ms |
| candidate | baseline | sftp_chmod | 56 | 5437.71 ms | 97.1 ms | 91.72 ms | 194.63 ms | 261.36 ms | 261.36 ms |
| candidate | baseline | sftp_rename | 56 | 5070.41 ms | 90.54 ms | 79.66 ms | 184.46 ms | 276.74 ms | 276.74 ms |
| candidate | baseline | sftp_readdir | 9 | 470.59 ms | 52.29 ms | 52.19 ms | 53.16 ms | 53.16 ms | 53.16 ms |
| candidate | baseline | sftp_mkdirall | 8 | 325.22 ms | 40.65 ms | 40.54 ms | 41.84 ms | 41.84 ms | 41.84 ms |
| candidate | baseline | ssh_connect | 1 | 323.42 ms | 323.42 ms | 323.42 ms | 323.42 ms | 323.42 ms | 323.42 ms |

</details>

<details><summary><code>large</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 85918.73 ms |
| candidate | baseline | connect | 316.57 ms |
| candidate | baseline | sweep_stale_temps | 157.06 ms |
| candidate | baseline | create_dirs | 82.08 ms |
| candidate | baseline | cleanup | 13.18 ms |
| candidate | baseline | local_scan | 0.28 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 2 | 171823.07 ms | 85911.54 ms | 85918.59 ms | 85918.59 ms | 85918.59 ms | 85918.59 ms |
| candidate | baseline | sftp_write | 2 | 171734.99 ms | 85867.5 ms | 85873.09 ms | 85873.09 ms | 85873.09 ms | 85873.09 ms |
| candidate | baseline | ssh_connect | 1 | 316.56 ms | 316.56 ms | 316.56 ms | 316.56 ms | 316.56 ms | 316.56 ms |
| candidate | baseline | sftp_readdir | 3 | 157 ms | 52.33 ms | 52.2 ms | 52.86 ms | 52.86 ms | 52.86 ms |
| candidate | baseline | sftp_mkdirall | 2 | 82.07 ms | 41.04 ms | 41.53 ms | 41.53 ms | 41.53 ms | 41.53 ms |
| candidate | baseline | sftp_open | 2 | 32.49 ms | 16.25 ms | 16.91 ms | 16.91 ms | 16.91 ms | 16.91 ms |
| candidate | baseline | sftp_rename | 2 | 29.48 ms | 14.74 ms | 15.27 ms | 15.27 ms | 15.27 ms | 15.27 ms |
| candidate | baseline | sftp_chmod | 2 | 26.87 ms | 13.44 ms | 13.48 ms | 13.48 ms | 13.48 ms | 13.48 ms |

</details>

### Delete sweeps

The pre-clean before every measured run wipes the tree the previous repeat left behind, which makes it a pure delete sweep. It costs no extra time (it has always run) and its numbers never enter the upload tables above. Sweeps that found an empty directory are not counted.

| Scenario | Build | Profile | Sweeps | Files deleted | Median | files/s | remote_scan | delete_sweep |
|---|---|---|---|---|---|---|---|---|
| large | candidate | baseline | 2 | 2 | 652 ms | 3.07 | 161.3 ms | 58.99 ms |
| mixed | candidate | baseline | 2 | 56 | 1945 ms | 28.79 | 476.97 ms | 1064.01 ms |
| small | candidate | baseline | 2 | 300 | 5876 ms | 51.06 | 488.68 ms | 4920.77 ms |

| Scenario | Build | Profile | Operation | Count | Cumulative | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| large | candidate | baseline | ssh_connect | 1 | 350.23 ms | 350.23 ms | 350.23 ms | 350.23 ms | 350.23 ms |
| large | candidate | baseline | sftp_readdir | 4 | 213.96 ms | 53.99 ms | 54.19 ms | 54.19 ms | 54.19 ms |
| large | candidate | baseline | sftp_rmdir | 2 | 29.92 ms | 15.13 ms | 15.13 ms | 15.13 ms | 15.13 ms |
| large | candidate | baseline | sftp_remove | 2 | 29.05 ms | 14.53 ms | 14.53 ms | 14.53 ms | 14.53 ms |
| mixed | candidate | baseline | sftp_remove | 56 | 865.07 ms | 14.74 ms | 15.98 ms | 44.57 ms | 44.57 ms |
| mixed | candidate | baseline | sftp_readdir | 10 | 530.92 ms | 53.01 ms | 54.06 ms | 54.06 ms | 54.06 ms |
| mixed | candidate | baseline | ssh_connect | 1 | 334.31 ms | 334.31 ms | 334.31 ms | 334.31 ms | 334.31 ms |
| mixed | candidate | baseline | sftp_rmdir | 8 | 121.93 ms | 15.22 ms | 16.27 ms | 16.27 ms | 16.27 ms |
| small | candidate | baseline | sftp_remove | 300 | 4791.03 ms | 14.76 ms | 15.98 ms | 54.05 ms | 66.22 ms |
| small | candidate | baseline | sftp_readdir | 10 | 541.33 ms | 54.18 ms | 54.66 ms | 54.66 ms | 54.66 ms |
| small | candidate | baseline | ssh_connect | 1 | 348.82 ms | 348.82 ms | 348.82 ms | 348.82 ms | 348.82 ms |
| small | candidate | baseline | sftp_rmdir | 8 | 128.53 ms | 15.06 ms | 20.38 ms | 20.38 ms | 20.38 ms |

Data only: these numbers set no threshold and fail no build. Collected to evaluate the single-connection ceiling discussed in issue #158 and to show where a run spends its time.
