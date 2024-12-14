# GitHub Readme Stats (Go)

```sh
cp .env.example .env   # add PAT_1, ...
go run ./cmd/server    # listens on :9000
```

**Endpoints:** `GET /api` (stats), `/api/pin` (repo), `/api/top-langs` (languages), `/api/wakatime`, `/api/gist`, `/api/status/up`, `/api/status/pat-info`
Legacy paths `/pin`, `/top-langs`, `/wakatime`, `/gist` also work.

`make build|test|docker|theme-doc|langs-json`

---

## Quick Start

```md
![GitHub Stats](https://YOUR_DOMAIN/api?username=wthrajat)
![Top Languages](https://YOUR_DOMAIN/api/top-langs?username=wthrajat)
![Repo Card](https://YOUR_DOMAIN/api/pin?username=wthrajat&repo=github-readme-stats)
![Gist Card](https://YOUR_DOMAIN/api/gist?id=GIST_ID)
![WakaTime](https://YOUR_DOMAIN/api/wakatime?username=YOUR_WAKATIME_USER)
```

---

## Cards & Key Options

| Card | Endpoint | Key Options |
|------|----------|-------------|
| **Stats** | `/api` | `hide`, `show`, `show_icons`, `hide_rank`, `rank_icon`, `include_all_commits`, `theme`, `custom_title`, `locale`, `number_format`, `lowercase` |
| **Repo** | `/api/pin` | `show_owner`, `description_lines_count` |
| **Gist** | `/api/gist` | `show_owner` |
| **Languages** | `/api/top-langs` | `layout` (normal\|compact\|donut\|donut-vertical\|pie), `langs_count`, `hide`, `hide_progress`, `size_weight`, `count_weight`, `exclude_repo` |
| **WakaTime** | `/api/wakatime` | `layout` (default\|compact), `langs_count`, `hide`, `hide_progress`, `display_format` (time\|percent), `api_domain` |

### Common Options (all cards)
`title_color`, `text_color`, `icon_color`, `bg_color` (hex or `DEG,C1,C2...` gradient), `border_color`, `hide_border`, `theme`, `border_radius`, `cache_seconds`, `disable_animations`

### Themes
81 built-in: `default`, `dark`, `radical`, `gruvbox`, `tokyonight`, `dracula`, `nord`, `github_dark`, `catppuccin_mocha`, `rose_pine`, `transparent`, ...  
Full list & preview: [themes/README.md](themes/README.md) • Config: [internal/themes/themes.go](internal/themes/themes.go)

### Dynamic Themes (GitHub light/dark)
- `theme=transparent` — works on both
- `#gh-dark-mode-only` / `#gh-light-mode-only` URL fragments
- `<picture>` with `prefers-color-scheme` media queries

### Locales
29 supported: `en`, `cn`, `zh-tw`, `de`, `es`, `fr`, `ja`, `kr`, `ru`, `pt-br`, `it`, `pl`, `tr`, `nl`, `hu`, `vi`, `id`, `se`, `uk-ua`, `cs`, `sk`, `np`, `bn`, `ml`, `my`, `el`, `ar`, `uz`  
Pass `locale=CODE` (e.g. `locale=es`).

---

## Rank Algorithm

Based on [Japanese academic grading](https://wikipedia.org/wiki/Academic_grading_in_Japan): S (top 1%), A+, A, A-, B+, B, B-, C+, C.  
Percentile = weighted CDF of commits, PRs, reviews, issues, stars, followers (exponential + log-normal).  
Circle = 100 − percentile.  
Implementation: [internal/rank/rank.go](internal/rank/rank.go)

---

## Language Stats Algorithm

`rank = (bytes ^ size_weight) × (repos ^ count_weight)`  
Defaults: `size_weight=1`, `count_weight=0` (bytes only).  
Recommended: `size_weight=0.5&count_weight=0.5`.

---

## Deploy

### Vercel (native Go)
1. Fork → Import in Vercel
2. Add `PAT_1` (GitHub token with `repo` + `user` scopes) in Project → Settings → Environment Variables
3. Deploy — `vercel.json` sets Go preset automatically

### Docker
```sh
cp .env.example .env
docker build -t github-readme-stats .
docker run -p 9000:9000 --env-file .env github-readme-stats
```

### Go (anywhere)
```sh
cp .env.example .env
go run ./cmd/server
```

### Private Instance
Set `GRS_TOKEN` + optional `ALLOWED_USERNAMES` (comma-separated). Requests need `?token=...` or `x-grs-token` header.

### Env Vars
See [.env.example](.env.example): `PAT_1`, `CACHE_SECONDS`, `GRS_TOKEN`, `ALLOWED_USERNAMES`, `FETCH_MULTI_PAGE_STARS`, `PORT`

---

- Made with Go ❤️
- Idea inspiration from Anurag's [GRS](https://github.com/anuraghazra/github-readme-stats)
- SVG cards served by a single `net/http` binary. No Node, no JavaScript/TypeScript which is the bessssttttt

---
