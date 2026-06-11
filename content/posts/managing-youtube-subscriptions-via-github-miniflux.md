---
title: "Managing YouTube subscriptions via GitHub and Miniflux"
description: "How I moved my YouTube subscriptions to Miniflux, a self-hosted RSS reader, with feeds managed in a YAML file and synced via a GitHub Action."
date: 2024-07-21
updated: 2025-01-31
cover: /images/youtube-miniflux-hero.jpg
ogImage: /images/covers/youtube-miniflux.jpg
---

**I watch too much YouTube.** I'm not alone; 62% of US internet users visit the platform daily, and 92% weekly. It scratches an itch differently to Twitter/X or Instagram because I subscribe to channels that meaningfully impact my life. However, I often start well but become entranced by the “YouTube rabbit hole”.

The platform’s algorithm threads the needle by providing a balanced, never-ending stream of content that feels productive. It’s a completely different experience to Instagram Reels, for example, which feels like a gluttonous meal of content.

**This post outlines how I migrated my use of YouTube to [Miniflux](https://github.com/miniflux/v2)**, a self-hosted RSS reader, which I manage via a [GitHub repo](https://github.com/revett/feeds).

## Overview

The goals of the system are:

- Host a Miniflux instance
- Manage the subscriptions for Miniflux via a YAML file
- Sync subscription changes automatically via a GitHub Action
- Create a single feed of content

## Why?

[Miniflux](https://github.com/miniflux/v2) provides a single feed containing the news, videos, podcasts, and blog posts I subscribe to. **It has no algorithm. It has an end.** **I completely manage its content.** Miniflux consolidates everything I want to consume in one place.

**I read it once a day, knowing the content is valuable.** If I then go directly on a platform, e.g. Instagram, I know I'm enjoying a distraction. I’m not kidding myself that in some way it is good for me.

Automating the feeds that I subscribe to via GitHub is overengineering, however I prefer to manage configuration this way, see [revett/dotfiles](https://github.com/revett/dotfiles) as another example.

## Implementation

I initially used [DigitalOcean](https://www.digitalocean.com/) to run my Miniflux instance. Last year, I migrated it to [Railway](https://railway.app) as an excuse to try out the platform. I’ve been consistently impressed by the quality and pace of product development from the team — **have a read of their [changelogs](https://railway.app/changelog).**

I run the Miniflux instance in a Docker container (see [docs](https://miniflux.app/docs/docker.html)), alongside Postgres. It costs around $3 per month to run.

Once you have the Miniflux instance up and running, head to `/keys` and create an API key.

Create a new GitHub repo with this file structure — see [revett/feeds](https://github.com/revett/feeds) as an example.

```text
repo:
	- feeds.yml # List of subscriptions
  - .github:
		- workflows:
			- sync.yml # GitHub Action
```

The `feeds.yml` file lists the RSS feeds you want to subscribe to, grouped by category:

```yaml
Blog:
  - https://waitbutwhy.com/feed
HN:
  - https://hnrss.org/frontpage?points=200
Podcast:
  - https://anchor.fm/s/f88f5324/podcast/rss # The Kevin Rose Show
  - https://feeds.megaphone.fm/WWO6655869236 # The Prof G Pod with Scott Galloway
  - https://feeds.megaphone.fm/GLT9190936013 # The Rest Is Politics
  - https://lexfridman.com/feed/podcast
YouTube:
  - https://www.youtube.com/feeds/videos.xml?channel_id=UCB6s-V1Ls4vc_mXEF-4Lz_Q # Arvid Kahl
  - https://www.youtube.com/feeds/videos.xml?channel_id=UCm325cMiw9B15xl22_gr6Dw # Beau Miles
  - https://www.youtube.com/feeds/videos.xml?channel_id=UCtUId5WFnN82GdDy7DgaQ7w # Better Ideas
  - https://www.youtube.com/feeds/videos.xml?channel_id=UCIRjdHLsHq6xdtAHAmoueqg # Bouldering Bobat
  - https://www.youtube.com/feeds/videos.xml?channel_id=UCg97Ni73ozwdoSytGSqTRyA # Framelines
  - https://www.youtube.com/feeds/videos.xml?channel_id=UCDsElQQt_gCZ9LgnW-7v-cQ # Kirsten Dirining
  - https://www.youtube.com/feeds/videos.xml?channel_id=UCcGXEidw0qjNdq7Gii8gHgg # Project Kamp
  - https://www.youtube.com/feeds/videos.xml?channel_id=UC5mPJA4y5G8Z6aNkY6AxgAw # Van Neistat
```

Set up a GitHub Action to sync your subscriptions to the Miniflux instance:

```yaml
name: Sync feeds via revett/miniflux-sync

on: [push]

jobs:
  Run:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout repository
        uses: actions/checkout@v2

      - name: Download and read version
        id: get_version
        run: |
          VERSION=$(curl -s https://raw.githubusercontent.com/revett/miniflux-sync/main/VERSION)
          echo "VERSION=$VERSION" >> $GITHUB_ENV

      - name: Download and extract latest release
        run: |
          curl -L https://github.com/revett/miniflux-sync/releases/download/${{ env.VERSION }}/miniflux-sync_Linux_x86_64.tar.gz | tar -xz

      - name: Run CLI with arguments
        env:
          MINIFLUX_SYNC_ENDPOINT: ${{ secrets.MINIFLUX_SYNC_ENDPOINT }}
          MINIFLUX_SYNC_API_KEY: ${{ secrets.MINIFLUX_SYNC_API_KEY }}
        run: |
          if [ "${{ github.ref_name }}" == "main" ]; then
            ./miniflux-sync sync --path feeds.yml
          else
            ./miniflux-sync sync --dry-run --path feeds.yml
          fi
```

This GitHub Action uses the [revett/miniflux-sync](https://github.com/revett/miniflux-sync) CLI to sync the Miniflux instance with the feeds in `feeds.yml`.

Head to `/feeds/settings/secrets/actions` in your GitHub repo and set the following repository secrets:

- `MINIFLUX_SYNC_ENDPOINT` — e.g. `https://your-domain.com/v1/`
- `MINIFLUX_SYNC_API_KEY` — The API key that you generated earlier

On commits to `main`, the GitHub Action will sync changes. On all other branches, the [revett/miniflux-sync](https://github.com/revett/miniflux-sync) CLI uses `--dry-run` to only output what would be changed, if you wish to check changes before syncing.

## Screenshots

![image](/images/youtube-miniflux-screenshot-1.png)

![image](/images/youtube-miniflux-screenshot-2.png)

## Links

1. [Frequency of YouTube use in the United States as of 3rd quarter (2020)](https://www.statista.com/statistics/256896/frequency-with-which-us-internet-users-visit-youtube/)
2. [Understanding the “YouTube rabbit hole” (2019)](https://medium.com/swlh/understanding-the-youtube-rabbit-hole-4d98e921eabe)
3. [revett/feeds — My RSS feeds, managed in Miniflux](https://github.com/revett/feeds)
4. [revett/miniflux-sync — Manage and sync your Miniflux feeds with YAML](https://github.com/revett/miniflux-sync)

## Credits

Cover image created by [@viglomir](https://x.com/viglomir/status/1621915985606320131).
