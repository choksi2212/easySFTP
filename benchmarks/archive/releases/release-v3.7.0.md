# easySFTP benchmark: release v3.7.0

| Field | Value |
|---|---|
| Kind | release (official reference) |
| Version | `v3.7.0` |
| Recorded | 2026-08-25T20:27:44Z |
| Commit | `1d80caab46d9ae6281da2901ae4e79b479508f63` |
| Workflow run | https://github.com/eiserv/easySFTP/actions/runs/32895421674 |
| Raw data | [release-v3.7.0.json](release-v3.7.0.json) |
| Flat export | [release-v3.7.0.csv](release-v3.7.0.csv) |

## easySFTP benchmark

| Setting | Value |
|---|---|
| Candidate | `v3.7.0 (1d80caa)` |
| Baseline | `none` |
| Repeats per scenario | 3 |
| Runner | Linux 7.0.0-30-generic, 10 cpu |
| Link profiles | the real line |
| Settings | easySFTP defaults (no advanced.* overrides): concurrency 4, request_concurrency 16, retries 2, timeout 30s, mode overlay |

### The link

| Profile | When | RTT p50 | RTT p90 | Handshake | Control 1 stream | Control N streams | Host load |
|---|---|---|---|---|---|---|---|
| baseline | start | 13.33 ms | 14.41 ms | 432.51 ms | 0.39 MiB/s | 0.93 MiB/s | n/a |
| baseline | end | 13.26 ms | 13.44 ms | 400.51 ms | 0.36 MiB/s | 1.02 MiB/s | n/a |

No link shaping was requested: every profile here is the real line.

The control measurement uses `x/crypto/ssh` and `pkg/sftp` directly, never easySFTP's uploader. It separates "the line is slow" from "easySFTP is slow", and a single-stream control close to a scenario's own MiB/s means the run was network bound, where a code delta says nothing.

### Throughput

| Scenario | Build | Profile | Files | Size | Median | Min | Max | MAD | MiB/s | files/s | Retries | Errors | Failed runs | Delta |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| small | candidate | baseline | 300 | 1.2 MiB | 4334 ms | 4020 ms | 4834 ms | 314 ms | 0.27 | 69.22 | 0 | 0 | 0 | - |
| mixed | candidate | baseline | 56 | 11.6 MiB | 17219 ms | 15630 ms | 17444 ms | 225 ms | 0.68 | 3.25 | 0 | 0 | 0 | - |
| large | candidate | baseline | 2 | 32 MiB | 52390 ms | 51199 ms | 54418 ms | 1191 ms | 0.61 | 0.04 | 0 | 0 | 0 | - |

Delta compares each build's median against the `candidate` build **on the same link profile**; negative is faster. MAD is the median absolute deviation of the repeats: a delta smaller than it is inside this host's own noise.

### Resources (median per run)

| Scenario | Build | Profile | User CPU | Sys CPU | CPU % | Peak RSS | Go allocs | GCs | GC pause | Peak goroutines | Net sent |
|---|---|---|---|---|---|---|---|---|---|---|---|
| small | candidate | baseline | 195.47 ms | 267.6 ms | 9.74% | 11.8 MiB | 11.6 MiB | 5 | 0.43151 ms | 106 | 1.4 MiB |
| mixed | candidate | baseline | 211.78 ms | 234.88 ms | 2.43% | 11 MiB | 4 MiB | 1 | 0.215005 ms | 108 | 11.7 MiB |
| large | candidate | baseline | 411.71 ms | 542.29 ms | 1.8% | 8 MiB | 1.2 MiB | 0 | 0 ms | 25 | 32.1 MiB |

### Where the time goes

Phases are wall clock and add up to roughly the run's duration. Operation totals are **cumulative across parallel workers** and are normally larger than the phase they belong to; read them for their share and their per-call cost, never as wall clock.

<details><summary><code>small</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 3760.22 ms |
| candidate | baseline | connect | 428.74 ms |
| candidate | baseline | create_dirs | 55.55 ms |
| candidate | baseline | cleanup | 54.42 ms |
| candidate | baseline | sweep_stale_temps | 54.2 ms |
| candidate | baseline | local_scan | 4.79 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 300 | 231882.61 ms | 772.94 ms | 343.45 ms | 2036.49 ms | 3071.89 ms | 3126.45 ms |
| candidate | baseline | sftp_write | 300 | 76572.35 ms | 255.24 ms | 124.54 ms | 944.47 ms | 1042.08 ms | 1066.79 ms |
| candidate | baseline | sftp_open | 300 | 38675.72 ms | 128.92 ms | 52.12 ms | 422.64 ms | 827.13 ms | 847.26 ms |
| candidate | baseline | sftp_rename | 300 | 23090.15 ms | 76.97 ms | 42.76 ms | 177.88 ms | 889.42 ms | 892.36 ms |
| candidate | baseline | sftp_chmod | 300 | 19611.92 ms | 65.37 ms | 48.48 ms | 138.34 ms | 409.92 ms | 884.18 ms |
| candidate | baseline | sftp_readdir | 9 | 481.52 ms | 53.5 ms | 53.52 ms | 53.8 ms | 53.8 ms | 53.8 ms |
| candidate | baseline | sftp_mkdirall | 8 | 404.76 ms | 50.6 ms | 50.8 ms | 54.74 ms | 54.74 ms | 54.74 ms |
| candidate | baseline | ssh_connect | 1 | 389.03 ms | 389.03 ms | 389.03 ms | 389.03 ms | 389.03 ms | 389.03 ms |
| candidate | baseline | sftp_realpath | 3 | 39.69 ms | 13.23 ms | 13.25 ms | 13.31 ms | 13.31 ms | 13.31 ms |

</details>

<details><summary><code>mixed</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 16607.15 ms |
| candidate | baseline | connect | 421.28 ms |
| candidate | baseline | cleanup | 83.41 ms |
| candidate | baseline | sweep_stale_temps | 54.61 ms |
| candidate | baseline | create_dirs | 53.04 ms |
| candidate | baseline | local_scan | 2.2 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 56 | 234833.77 ms | 4193.46 ms | 3428.75 ms | 6950.69 ms | 16606.55 ms | 16606.55 ms |
| candidate | baseline | sftp_write | 56 | 108341.11 ms | 1934.66 ms | 744.72 ms | 4516.84 ms | 14590.91 ms | 14590.91 ms |
| candidate | baseline | sftp_rename | 56 | 11103.1 ms | 198.27 ms | 99.88 ms | 429.75 ms | 641.04 ms | 641.04 ms |
| candidate | baseline | sftp_chmod | 56 | 9905.71 ms | 176.89 ms | 143.36 ms | 395.1 ms | 741.39 ms | 741.39 ms |
| candidate | baseline | sftp_open | 56 | 1304.8 ms | 23.3 ms | 23.29 ms | 32.6 ms | 35.66 ms | 35.66 ms |
| candidate | baseline | sftp_readdir | 9 | 487.81 ms | 54.2 ms | 54.15 ms | 54.44 ms | 54.44 ms | 54.44 ms |
| candidate | baseline | ssh_connect | 1 | 380.59 ms | 380.59 ms | 380.59 ms | 380.59 ms | 380.59 ms | 380.59 ms |
| candidate | baseline | sftp_mkdirall | 8 | 378.14 ms | 47.27 ms | 48.24 ms | 52.86 ms | 52.86 ms | 52.86 ms |
| candidate | baseline | sftp_realpath | 3 | 39.87 ms | 13.29 ms | 13.28 ms | 13.36 ms | 13.36 ms | 13.36 ms |

</details>

<details><summary><code>large</code> phases and round-trips</summary>

| Build | Profile | Phase | Wall |
|---|---|---|---|
| candidate | baseline | upload | 51819.51 ms |
| candidate | baseline | connect | 423.25 ms |
| candidate | baseline | sweep_stale_temps | 53.57 ms |
| candidate | baseline | create_dirs | 45.34 ms |
| candidate | baseline | cleanup | 27.53 ms |
| candidate | baseline | local_scan | 0.32 ms |

| Build | Profile | Operation | Count | Cumulative | Avg | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| candidate | baseline | file_upload | 2 | 102241.87 ms | 51120.94 ms | 51819.37 ms | 51819.37 ms | 51819.37 ms | 51819.37 ms |
| candidate | baseline | sftp_write | 2 | 101336.7 ms | 50668.35 ms | 51370.18 ms | 51370.18 ms | 51370.18 ms | 51370.18 ms |
| candidate | baseline | ssh_connect | 1 | 384.04 ms | 384.04 ms | 384.04 ms | 384.04 ms | 384.04 ms | 384.04 ms |
| candidate | baseline | sftp_readdir | 3 | 159.76 ms | 53.25 ms | 53.18 ms | 53.48 ms | 53.48 ms | 53.48 ms |
| candidate | baseline | sftp_mkdirall | 2 | 86.26 ms | 43.13 ms | 45.26 ms | 45.26 ms | 45.26 ms | 45.26 ms |
| candidate | baseline | sftp_chmod | 2 | 84.33 ms | 42.17 ms | 71.17 ms | 71.17 ms | 71.17 ms | 71.17 ms |
| candidate | baseline | sftp_realpath | 3 | 39.96 ms | 13.32 ms | 13.18 ms | 13.75 ms | 13.75 ms | 13.75 ms |
| candidate | baseline | sftp_rename | 2 | 31.09 ms | 15.55 ms | 15.72 ms | 15.72 ms | 15.72 ms | 15.72 ms |
| candidate | baseline | sftp_open | 2 | 29.31 ms | 14.66 ms | 14.75 ms | 14.75 ms | 14.75 ms | 14.75 ms |

</details>

### Delete sweeps

The pre-clean before every measured run wipes the tree the previous repeat left behind, which makes it a pure delete sweep. It costs no extra time (it has always run) and its numbers never enter the upload tables above. Sweeps that found an empty directory are not counted.

| Scenario | Build | Profile | Sweeps | Files deleted | Median | files/s | remote_scan | delete_sweep |
|---|---|---|---|---|---|---|---|---|
| large | candidate | baseline | 2 | 2 | 618 ms | 3.24 | 105.93 ms | 34.88 ms |
| mixed | candidate | baseline | 2 | 56 | 687 ms | 81.51 | 108.54 ms | 132.49 ms |
| small | candidate | baseline | 2 | 300 | 1142 ms | 262.7 | 120.78 ms | 582.19 ms |

| Scenario | Build | Profile | Operation | Count | Cumulative | p50 | p90 | p99 | Max |
|---|---|---|---|---|---|---|---|---|---|
| large | candidate | baseline | ssh_connect | 1 | 407.73 ms | 407.73 ms | 407.73 ms | 407.73 ms | 407.73 ms |
| large | candidate | baseline | sftp_readdir | 4 | 211.33 ms | 52.88 ms | 53.06 ms | 53.06 ms | 53.06 ms |
| large | candidate | baseline | sftp_rmdir | 2 | 32.68 ms | 17.37 ms | 17.37 ms | 17.37 ms | 17.37 ms |
| large | candidate | baseline | sftp_remove | 2 | 32.45 ms | 17.19 ms | 17.19 ms | 17.19 ms | 17.19 ms |
| mixed | candidate | baseline | sftp_remove | 56 | 3399.33 ms | 61.21 ms | 97.14 ms | 104.02 ms | 104.02 ms |
| mixed | candidate | baseline | sftp_readdir | 10 | 534.01 ms | 53.25 ms | 54.21 ms | 54.21 ms | 54.21 ms |
| mixed | candidate | baseline | ssh_connect | 1 | 373.14 ms | 373.14 ms | 373.14 ms | 373.14 ms | 373.14 ms |
| mixed | candidate | baseline | sftp_rmdir | 8 | 169.13 ms | 21.29 ms | 27.82 ms | 27.82 ms | 27.82 ms |
| small | candidate | baseline | sftp_remove | 300 | 31887.21 ms | 110.73 ms | 131.49 ms | 134.44 ms | 135.37 ms |
| small | candidate | baseline | sftp_readdir | 10 | 583.71 ms | 57.9 ms | 67.18 ms | 67.18 ms | 67.18 ms |
| small | candidate | baseline | ssh_connect | 1 | 372.27 ms | 372.27 ms | 372.27 ms | 372.27 ms | 372.27 ms |
| small | candidate | baseline | sftp_rmdir | 8 | 173.46 ms | 22.34 ms | 29.16 ms | 29.16 ms | 29.16 ms |

Data only: these numbers set no threshold and fail no build. Collected to evaluate the single-connection ceiling discussed in issue #158 and to show where a run spends its time.
