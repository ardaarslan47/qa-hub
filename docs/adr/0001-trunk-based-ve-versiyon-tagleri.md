# 0001: Trunk-based geliştirme ve versiyon tag'leri

**Durum:** Kabul edildi, 2026-11-03

## Bağlam

Kullanıcı ve AFK agent'lar PR açmadan doğrudan push etmek istiyor. Biriken iş toplu kontrol ediliyor, sonra bir versiyon çıkıyor. Versiyonu branch adıyla tutmak (`v1`, `v1.1`) GitHub'da sürtünme yaratıyor: default branch, `Closes #N` ve scheduled workflow'lar. Araştırma: `docs/research/branch-modeli-trunk-based.md`.

## Karar

- Tek kalıcı branch `main`. Herkes doğrudan `main`'e push eder.
- Agent'lar lokalde kısa ömürlü branch/worktree kullanabilir. Remote'a sadece `main` gider.
- Versiyon, `main` üzerindeki stabil commit'e atılan annotated `vMAJOR.MINOR.PATCH` tag'i ve GitHub Release ile tutulur.
- Stabilizasyon sırasında agent'ların çalışması sürecekse `main`'den `release/vX.Y` kesilir. Fix'ler `main`'de yapılır, cherry-pick edilir. Tag release branch'inden atılır, sonra branch silinir.
- Commit mesajları Conventional Commits formatında. Bir sonraki versiyon numarası bunlardan türetilir.

## Uygulama

Kurallar `.githooks/rules/` altında commit'i ya da push'u engelleyen script'ler olarak tutulur. `.github/workflows/rules.yml` aynı kuralları her push'ta tekrar koşar. Kural metni bu ADR'de değil, script'lerin başlık yorumlarında yaşar.

## Sonuçlar

- `main` silinmez ve force push edilmez. Hatalı bir versiyon silinmez, yeni bir patch ile düzeltilir.
- Lokal hook'lar atlatılabilir. Sunucu tarafı garanti GitHub ruleset'leriyle sağlanır: `main` silme/force push yasağı ve `v*` tag'lerini sadece repo sahibinin oluşturması.
