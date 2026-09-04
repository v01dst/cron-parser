<div align="center">

# ⏰ cron-parser

**Decode cron expressions, list upcoming runs. Never misread a crontab again.**

[![CI](https://github.com/v01dst/cron-parser/actions/workflows/ci.yml/badge.svg)](https://github.com/v01dst/cron-parser/actions/workflows/ci.yml)
![License](https://img.shields.io/badge/license-MIT-8A2BE2)
![Go](https://img.shields.io/badge/go-1.27-00ADD8?logo=go&logoColor=white)
![Deps](https://img.shields.io/badge/dependencies-stdlib%20only-00c853)
![Tests](https://img.shields.io/badge/tests-9%20passing-brightgreen)

`next runs` · `matches check` · `field explain` · `names & ranges` · `vixie OR semantics`

</div>

---

## ✨ Features

- **📅 Full 5-field cron** — `*`, lists, ranges, steps, day/month names (`mon`, `jan`)
- **🔮 Next-N scheduling** — list the next 1–100 upcoming runs from any timestamp
- **✅ Match checks** — scriptable yes/no match against any RFC3339 time
- **🔍 Field explain** — human-readable expansion of every field
- **📐 Vixie semantics** — day-of-month/weekday OR rule when both are restricted
- **🪶 Pure stdlib** — zero dependencies, single binary

## 🚀 Quick Start

```bash
git clone https://github.com/v01dst/cron-parser
cd cron-parser
go build -o cron-parser .
./cron-parser next "*/15 9-17 * * mon-fri" 3
```

Or with Docker:

```bash
docker build -t cron-parser .
docker run --rm cron-parser next "0 0 1 * *" 2
```

## 📖 Usage

```
cron-parser next "<expr>" [count] [--from <RFC3339>]
cron-parser matches "<expr>" <RFC3339>
cron-parser explain "<expr>"
```

### Examples

```bash
$ cron-parser next "*/15 9-17 * * mon-fri" 3
2026-09-04T09:00:00Z
2026-09-04T09:15:00Z
2026-09-04T09:30:00Z

$ cron-parser explain "30 4 1,15 * sat"
expression: 30 4 1,15 * sat
  minute:  30
  hour:    4
  day:     1 15
  month:   1 2 3 4 5 6 7 8 9 10 11 12
  weekday: 6

$ cron-parser matches "0 9 * * mon-fri" 2026-09-07T09:00:00Z
yes

$ cron-parser next "0 0 1 * *" 2        # month boundary handled
2026-10-01T00:00:00Z
2026-11-01T00:00:00Z
```

| Exit code | Meaning            |
|-----------|--------------------|
| `0`       | Success            |
| `1`       | Invalid expression |
| `2`       | Usage error        |

## 🧱 Tech Stack

| Layer     | Tech                  |
|-----------|-----------------------|
| Language  | Go 1.27 (stdlib only) |
| Testing   | go test               |
| Packaging | Docker (multi-stage)  |
| CI        | GitHub Actions        |

---

<div align="center">

Built with ⚡ by **v01dst**

[![GitHub](https://img.shields.io/badge/github-v01dst-181717?logo=github)](https://github.com/v01dst)
[![Discord](https://img.shields.io/badge/discord-9p.1-5865F2?logo=discord&logoColor=white)](https://discord.com/users/9p.1)

*Project 018 / 99 — The Loop*

</div>
