---
layout: post
title: "How to Generate a GitHub Personal Access Token for API Access"
description: "Step-by-step guide to creating a GitHub personal access token (classic and fine-grained) for use with CLI tools and API automation."
date: 2026-03-25
author: OnlyCLI Team
tags:
  - guide
  - github
  - authentication
---

Command-line tools and automation scripts that call the **GitHub REST API** need credentials. For machines and local development, a **personal access token (PAT)** is the usual choice. This post explains **why** tokens exist, how **classic** tokens differ from **fine-grained** tokens, and how to use one safely with an **OnlyCLI-generated** GitHub CLI.

Whether you are scripting **repos** and **issues** or letting an **LLM agent** run the same commands you would type, the contract is identical: GitHub’s REST API expects credentials on each request, and a PAT is the low-ceremony way to satisfy that for local tools.

## Why You Need a Token

Unauthenticated requests to `https://api.github.com` are **heavily rate-limited** and cannot access private repositories or most account-specific data. GitHub expects either:

- **OAuth** flows (great for web apps), or  
- **PATs** and **GitHub Apps** for scripts, CLIs, and servers.

For a generated CLI that speaks plain HTTP with a bearer header, a PAT is the straightforward option: you export a secret once per session (or store it in a profile), and every request includes `Authorization: Bearer <token>`.

## Classic Token vs Fine-Grained Token

GitHub now offers two PAT styles:

| Aspect | Classic PAT | Fine-grained PAT |
|--------|-------------|------------------|
| **Scope model** | Broad OAuth-style scopes (`repo`, `read:org`, …) | Repository- and organization-specific permissions |
| **Expiration** | Optional but strongly recommended | Required (max ~1 year) |
| **Enterprise policy** | Often restricted by admins | Often preferred by security teams |
| **Migration** | Older integrations assume classic scopes | New projects should evaluate fine-grained first |

**Classic** tokens are simple: one token, many repos, scope checklist. **Fine-grained** tokens limit blast radius: you pick which repos and which actions (contents read, issues write, etc.). If your org mandates least privilege, fine-grained is usually the right default.

## Step-by-Step: Creating a Classic Token

These steps match the GitHub web UI as of early 2026; if menus move slightly, search the settings bar for **“Personal access tokens”**.

1. Sign in to GitHub in the browser.
2. Open **Settings** (your user avatar → **Settings**).
3. In the left sidebar, scroll to **Developer settings** (at the bottom).
4. Click **Personal access tokens**.
5. Select **Tokens (classic)**.
6. Click **Generate new token** → **Generate new token (classic)**.
7. Give the token a **note** (e.g. `onlycli-local`).
8. Choose an **expiration** (shorter is better for laptops; use automation users for bots).
9. Enable **scopes** your CLI needs (see the next section).
10. Click **Generate token** and **copy the value immediately**—GitHub will not show it again.

Store the token in a password manager or secret store until you export it in a shell.

## Step-by-Step: Creating a Fine-Grained Token

1. Go to **Settings** → **Developer settings** → **Personal access tokens** → **Fine-grained tokens**.
2. Click **Generate new token**.
3. Set a **name** and **expiration**.
4. Under **Resource owner**, choose your user or an org (if allowed).
5. Under **Repository access**, choose **All repositories** or **Only select repositories** (prefer select when possible).
6. Under **Permissions**, grant the minimum **Repository** and **Account** permissions for your workflow (e.g. **Contents: Read-only** for public repo browsing; **Issues: Read and write** to triage tickets).
7. Generate the token and **copy it once**.

Fine-grained tokens often start with a prefix that differs from classic `ghp_` tokens; GitHub’s UI labels the prefix when you create the token.

## Recommended Scopes for Common Operations

**Classic PAT** (illustrative—tighten to what you actually need):

- **Public read-only**: often `public_repo` is enough for public API exploration; many read endpoints work with no scope but hit rate limits fast.
- **Private repos**: `repo` (full control of private repositories).
- **Org and team visibility**: `read:org` (and sometimes `read:user`).

**Fine-grained**: map each CLI command you plan to run to a permission (e.g. **Pull requests: Read-only** for listing PRs). When in doubt, generate a token, run the CLI with `--verbose`, inspect a 403 response, and add one permission at a time.

## Using the Token with an OnlyCLI-Generated GitHub CLI

OnlyCLI can generate a GitHub-oriented CLI from the official REST OpenAPI description. The generated runtime reads **`GITHUB_TOKEN`** from the environment for bearer authentication, and can also load credentials from **config profiles** so you are not re-exporting secrets in every new terminal.

### Session-only: environment variable

Export the token for the current shell session:

```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
./github repos list-for-user --username octocat
```

Replace `./github` with the path to your built binary. To see the resolved request without sending it, many generated CLIs support **`--dry-run`** on the root command; add **`--verbose`** once if you need headers and URL details on stderr while keeping stdout clean.

If you use **GitHub Enterprise Server**, configure the API base URL the way your generated project documents (commonly `GITHUB_BASE_URL` or `config set` on `base_url`). The public `api.github.com` host applies only to GitHub.com.

Never paste real tokens into chat logs, tickets, or screen shares.

### When requests fail with 401 or 403

- **401 Unauthorized** usually means a missing token, a typo, or a revoked credential. Regenerate and update your environment or profile.
- **403 Forbidden** often signals **insufficient scopes** (classic PAT) or **missing repository permissions** (fine-grained PAT). Expand permissions one step at a time until the call succeeds.
- **Aggressive rate limits** on unauthenticated traffic improve dramatically once a valid token is attached; watch `X-RateLimit-Remaining` when you bulk-fetch data.

## Security Best Practices

- **Expiration**: Always set one. Rotate before expiry.
- **Minimal scope**: Prefer fine-grained tokens with only the repos and permissions required.
- **Never commit**: Add `*.env` and local scripts to `.gitignore`; use `git-secrets` or similar in CI.
- **CI/CD**: Use **OIDC**, **GitHub Actions** `GITHUB_TOKEN`, or **encrypted secrets**—not a developer’s PAT checked into the repo.
- **Revocation**: If a token leaks, revoke it immediately under the same Developer settings pages.

## Token Storage: Config Profiles

Typing `export` every session is fine for quick tests. For daily use, generated GitHub CLIs ship a **`config`** command that stores values under your user config directory (the exact path is printed if saving fails).

The **`profile.key`** form targets a named profile—in this case **`default`**:

```bash
./github config set default.token ghp_xxxxxxxxxxxx
```

Switch the active default profile when your build supports it:

```bash
./github config use-profile default
```

For a second account, write `work.token` (or any profile name) and call `config use-profile work` when you need that context. Prefer **profile files with restrictive permissions** (`0600`) over world-readable dotfiles, and never commit the config file to a repository.

## Conclusion

GitHub PATs power automation where OAuth redirects are awkward. **Classic** tokens are quick; **fine-grained** tokens align with least privilege. Pair either with an **OnlyCLI-generated CLI** via **`GITHUB_TOKEN`** or **config profiles**, and treat every token like a password: short-lived, narrowly scoped, and never committed.

---

*For generating the GitHub CLI from OpenAPI, see the [OnlyCLI README](https://github.com/onlycli/onlycli) and [Getting Started with OnlyCLI]({{ '/blog/getting-started-with-onlycli/' | relative_url }}).*
