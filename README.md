# shrinkr

A CLI that re-compresses photos and videos exported from Google Photos via Takeout. Meant as the middle step of a pipeline that archives originals to S3 / Backblaze B2, then re-uploads compressed versions to Google Photos to reclaim quota.

Ships as a single Go binary plus a couple of external CLIs (`ffmpeg`, `exiftool`, ...).

## Pipeline

```bash
# 1. Export the album via Google Takeout as a tgz.
#    The Google Photos Library API scope change (2025-03) removed third-party
#    read access to the user's library, so downloads have to be done by hand.

tar xzf ~/Downloads/takeout-YYYYMMDD.tgz -C ~/photos-raw

# 2. Archive the originals to Backblaze B2 (rclone remote must be pre-configured).
rclone copy ~/photos-raw b2:my-photos-archive --transfers 8

# 3. Compress.
shrinkr run ~/photos-raw ~/photos-slim --preset balanced --report ~/photos-slim/report.json

# 4. Delete the original album from Google Photos in the web UI.

# 5. Re-upload the compressed copies.
gphotos-uploader-cli push ~/photos-slim
```

## Install

```bash
go install github.com/terakoya76/shrinkr@latest
```

### External dependencies

Required:

- `ffmpeg` (`ffprobe` ships with it)
- `exiftool`

Optional (shrinkr still runs without them, but per-format features degrade):

- `cjpeg` (mozjpeg or libjpeg-turbo) — smaller JPEGs than the pure-Go fallback (mozjpeg: 15-30%, libjpeg-turbo: 5-15%)
- `cwebp` — without it, WebP inputs are skipped
- `heif-convert` (libheif) — without it, HEIC inputs are skipped
- `oxipng` — without it, PNGs are copied through unchanged

Ubuntu:

```bash
# libjpeg-turbo-progs provides a compatible `cjpeg` that shrinkr uses in place
# of mozjpeg (JPEGs end up a few percent larger than with true mozjpeg).
sudo apt install ffmpeg libimage-exiftool-perl libjpeg-turbo-progs webp libheif-examples

# oxipng is not packaged for default apt repository, install it via cargo instead.
cargo install oxipng

# Actual mozjpeg (~15-30% smaller JPEGs than libjpeg-turbo's cjpeg) is not
# packaged for Ubuntu. Build from source if you want it:
#   https://github.com/mozilla/mozjpeg  (needs cmake and nasm)
```

macOS:

```bash
brew install ffmpeg exiftool mozjpeg webp libheif oxipng
```

## Usage

```bash
shrinkr doctor            # Report which dependencies are available.
shrinkr presets           # List embedded presets.
shrinkr run <src> <dst>   # Compress.
```

### Main `run` flags

| flag                                | default    | description                                                                                    |
| ----------------------------------- | ---------- | ---------------------------------------------------------------------------------------------- |
| `--preset`                          | `balanced` | Choose from `aggressive` / `balanced` / `conservative`.                                        |
| `--config <yaml>`                   | -          | Overlay a YAML file on top of the preset.                                                      |
| `--workers`                         | `NumCPU`   | Concurrent jobs.                                                                               |
| `--dry-run`                         | off        | Plan only; write nothing.                                                                      |
| `--min-savings 0.10`                | preset     | If the compressed output is not at least this much smaller, copy the original instead.         |
| `--image-max-edge`                  | preset     | Long-edge cap for images.                                                                      |
| `--video-max-height`                | preset     | Height cap for videos.                                                                         |
| `--jpeg-quality`                    | preset     | Around 82.                                                                                     |
| `--webp-quality`                    | preset     | Around 80.                                                                                     |
| `--video-crf`                       | preset     | H.264 CRF (lower = higher quality). HEVC sources add +5 internally to keep visual quality parity. |
| `--report <path>`                   | -          | Write a JSON summary to this path.                                                             |
| `--overwrite`                       | off        | Overwrite outputs even if they look current (default: skip via mtime check).                   |
| `--include-glob` / `--exclude-glob` | -          | Filename filters.                                                                              |

### Presets (embedded)

| preset       | img_edge | jpeg_q | webp_q | vid_h | vid_crf | min_save |
| ------------ | -------- | ------ | ------ | ----- | ------- | -------- |
| aggressive   | 2560     | 75     | 72     | 720   | 30      | 0.05     |
| balanced     | 4096     | 82     | 80     | 1080  | 26      | 0.10     |
| conservative | 6000     | 88     | 85     | 1440  | 22      | 0.15     |

## Behavior details

- **Input formats**: `.jpg` / `.png` / `.heic` / `.heif` / `.webp` / `.mp4` / `.mov` / `.mkv` / `.avi` / `.3gp` / `.m4v` / `.webm`. Everything else is ignored (Takeout sidecar `.json` files are also skipped automatically).
- **Output formats**: HEIC becomes JPEG. Video containers are always MP4 (AAC audio). Video codec matches the source — HEVC (H.265) input is re-encoded with libx265 (kept HEVC, hvc1-tagged), and everything else with libx264. Transcoding an already-HEVC source to H.264 at the same CRF often makes the file larger, which auto-preservation avoids.
- **Metadata preservation**: EXIF (`DateTimeOriginal`, GPS, orientation, ...) is copied to the destination with `exiftool -TagsFromFile`. Video container metadata (`creation_time`, ...) is preserved via `ffmpeg -map_metadata 0`. File mtime is aligned to the source.
- **Idempotent**: If the destination already exists and is not older than the source, the job is skipped. Use `--overwrite` to force.
- **min-savings**: If the compressed output is larger than `(1 - min-savings) * source_size`, shrinkr writes the original bytes instead. Avoids re-encoding when the result would not actually save space.
- **Error handling**: Not fail-fast. A failing job is skipped and the run continues; the per-file error message only lives in the `--report` JSON, so pass `--report path.json` whenever you might want to debug failures. Exit code is 1 only if every job failed.

## Development

```bash
go test ./...
go vet ./...
go build ./...
./shrinkr doctor
```

External dependencies come from system packages. CI installs them via `apt-get install -y ffmpeg libimage-exiftool-perl` before running `go test` (see `.github/workflows/ci.yml`).
