# GitHub Readme Stats (Go)

Turn your GH profile into clean SVG cards that you can drop into any README.

![Go](https://img.shields.io/badge/Go-1.27-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-blue)
![Tests](https://img.shields.io/badge/Tests-passing-brightgreen)
![Vercel](https://img.shields.io/badge/Deployed%20on-Vercel-black?logo=vercel)
![Themes](https://img.shields.io/badge/Themes-81-orange)
![Languages](https://img.shields.io/badge/Locales-29-purple)

Stats, languages, pinned repos, gists, and coding time. One small server, no Node, no JavaScript.

## Usage

Deploy your own instance first, then add these to your README. Replace `YOUR_INSTANCE` with your deployment URL, for example `https://your-app.vercel.app`.

```md
![GitHub Stats](https://YOUR_INSTANCE/api?username=YOUR_USERNAME)

![Top Languages](https://YOUR_INSTANCE/api/top-langs?username=YOUR_USERNAME)

![Repo Card](https://YOUR_INSTANCE/api/pin?username=YOUR_USERNAME&repo=github-readme-stats)

![Gist Card](https://YOUR_INSTANCE/api/gist?id=YOUR_GIST_ID)

![WakaTime](https://YOUR_INSTANCE/api/wakatime?username=YOUR_WAKATIME_USER)
```

## Cards

| Card | Endpoint |
| ---- | -------- |
| Stats | `/api?username=` |
| Top Languages | `/api/top-langs?username=` |
| Pinned Repo | `/api/pin?username=&repo=` |
| Gist | `/api/gist?id=` |
| WakaTime | `/api/wakatime?username=` |

## Options

Every card accepts these.

| Option | What it does |
| ------ | ------------ |
| `theme` | Picks one of the [81 themes](themes/README.md) |
| `title_color` | Color of the title |
| `text_color` | Color of the text |
| `bg_color` | Background color, gradients work too |
| `hide_border` | Removes the card border |
| `locale` | Card language, try `locale=es` |
| `cache_seconds` | How long the card is cached |

Cards also have their own extras. Stats supports `hide`, `show_icons`, `hide_rank`, and `include_all_commits`. Languages supports `layout` with `compact`, `donut`, and `pie`, plus `langs_count` and `hide`. WakaTime supports `layout=compact` and `hide_progress`.

## Deploy on Vercel

1. Fork this repo
2. Import it on [Vercel](https://vercel.com/new)
3. Add your GitHub token as `PAT_1` in the environment variables
4. Deploy

## Run it yourself

```sh
cp .env.example .env   # add your token
docker build -t grs .
docker run -p 9000:9000 --env-file .env grs
```

Or with Go installed:

```sh
cp .env.example .env   # add your token
go run ./cmd/server
```
- Idea inspired by [github-readme-stats](https://github.com/anuraghazra/github-readme-stats).
