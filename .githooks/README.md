# Git hook kuralları

Bu repo'da yorum yasak (`rules/pre-commit/no-comment`): kod kendini isimleriyle açıklar.

## Kurallar

`rules/<hook>/<kural-adı>`: çalıştırılabilir bir bash script. Sözleşme:

- Geçerse exit 0.
- İhlalde stderr'e `[<kural-adı>]` önekiyle ne yanlış, nerede ve nasıl düzeltileceğini yazar, exit 1.

| Hook | Kuralın aldığı girdi |
|---|---|
| `pre-commit` | yok; `git diff --cached` okunur |
| `commit-msg` | `$1` = mesaj dosyası |
| `pre-push` | `$1` = remote; stdin satırları `<local-ref> <local-sha> <remote-ref> <remote-sha>` |

## Script'ler

- `run <hook> [git'in hook'a verdiği argümanlar...]`: `rules/<hook>/` altındaki tüm çalıştırılabilir kuralları çalıştırır. `pre-commit`, `commit-msg` ve `pre-push` hook'ları sadece bunu çağırır.
- `install`: `core.hooksPath`'i `.githooks` yapar ve hook'ların etkin olduğunu doğrular. Idempotent; değer zaten doğruysa config'e yazmaz, hook'lar etkin değilse exit 1. Her AFK agent setup'ında, agent başlamadan çalıştır:
  - Sandcastle: `host.onWorktreeReady` (sandbox içinde `.git/config` salt okunur), doğrulama için `sandbox.onSandboxReady`.
  - Claude Code: `.claude/settings.json` SessionStart (zaten bağlı).
  - Diğer cloud agent'lar: environment setup script'i.
- `check-range <önceki-sha> <son-sha>`: aralıktaki her commit'in mesajını `commit-msg` kurallarından, diff'ini `pre-commit` kurallarından geçirir. `önceki-sha` sıfırlardan oluşuyorsa (yeni branch) sadece `son-sha` kontrol edilir. CI'da, lokal hook'ları atlatmış commit'leri yakalamak için kullanılır. Kontrol sırasında HEAD'i geçici olarak her commit'in parent'ına taşır ve ayrı bir index kullanır; çıkarken HEAD'i geri koyar, çalışma dizinine ve index'e dokunmaz.
- `tests/run`: kuralları geçici repolarda gerçek `git commit` / `git push` ile uçtan uca test eder.

## CI

Lokal hook'lar atlatılabilir (`--no-verify`, hook'ları kurmamış bir clone, başka bir AI aracı). `.github/workflows/rules.yml` her push'ta testleri ve aynı kuralları tekrar çalıştırır; ihlal main'de kırmızı build olarak görünür.
