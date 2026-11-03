QA Hub: TestRail/Zephyr benzeri basit bir test yönetim uygulaması. Projenin amacı, ürün tarifi ve açık sorular: `docs/concept.md`.

## Agent skills

### Issue tracker

Issue'lar bu reponun GitHub Issues'ında, `gh` CLI ile tutulur. Bkz. `docs/agents/issue-tracker.md`.

### Triage label'ları

Standart adlar, değiştirilmedi. Bkz. `docs/agents/triage-labels.md`.

### Domain dokümanları

Single-context. Bkz. `docs/agents/domain.md`.

### Commit ve push kuralları

Trunk-based: herkes doğrudan `main`'e push eder, versiyonlar `vX.Y.Z` tag'leriyle tutulur (bkz. `docs/adr/0001-trunk-based-ve-versiyon-tagleri.md`). Kurallar `.githooks/rules/` altında commit'i ya da push'u engelleyen script'lerdir. Bir hook seni engellediğinde ihlali düzelt. Yeni bir kuralı zorunlu yapmak için `/encode-rule` kullan.

Bu repoda PR yok. pstack playbook'ları bir PR açmayı, PR'a push etmeyi ya da merge etmeyi söylediğinde bunun yerine doğrulanmış işi Conventional Commit'lerle doğrudan `main`'e push et. Agent'lar lokalde worktree ya da kısa ömürlü branch kullanabilir, remote'a sadece `main` gider.
