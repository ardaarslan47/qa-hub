# 0002: Tech stack olarak Go

**Durum:** Kabul edildi, 2026-11-03

## Bağlam

QA Hub'ın kodunu kullanıcı okumayacak. Kodu AFK agent'lar yazacak, agent'ın hatasını derleyici ve testler yakalayacak. Kullanıcı TypeScript, React, SQLite, Vitest ve Playwright'ı okuyabiliyor. Okuyamadığı bir stack, bu deneyin amacına daha uygun.

Seçimi belirleyen ölçütler:

- Agent bir hatayı ne kadar erken görüyor? Statik tipler ve hızlı derleme bu süreyi kısaltır.
- Farklı agent'ların yazdığı kod birbirine ne kadar benziyor?
- CI'da ve agent worktree'lerinde kurulacak bir servis var mı?
- Stack, `no-comment` hook'uyla ne kadar sürtünme çıkarıyor?

Rust daha çok hatayı derleme anında yakalıyor. Buna karşılık derleme süresi daha uzun. Agent'ların borrow checker'la daha çok tur harcayacağını tahmin ettik, bunu ölçmedik. Elixir, Ruby ve Python'da tip kontrolü yok ya da zayıf. C# ve Kotlin bu boyutta bir uygulama için ağır kalıyor.

## Karar

- Uygulama dili Go. HTTP için standart kütüphanedeki `net/http` kullanılır.
- HTML sunucuda üretilir. Şablonlar templ ile yazılır, etkileşim htmx ile sağlanır. templ şablonları Go'ya çevrildiği için yanlış bir alan adı derleme hatası olur.
- Veritabanı SQLite. Sürücü `modernc.org/sqlite`, saf Go olduğu için cgo gerektirmez. Her agent worktree'si kendi DB dosyasını kullanır.
- Migration'lar düz SQL dosyalarıdır ve küçük bir runner tarafından uygulanır. goose gibi araçların `-- +goose Up` satırları `no-comment` hook'una takılır.
- Birim ve entegrasyon testleri `go test` ile koşar.
- Kabul testleri Playwright ile, TypeScript'te yazılır. Kullanıcı uygulama kodunu okumaz ama bu testleri okuyabilir. Agent'ın işi bu senaryolar üzerinden kabul edilir.
- templ'in ürettiği `*_templ.go` dosyaları commit edilmez. Build ve CI bunları `templ generate` ile üretir.

## Sonuçlar

- `no-comment` hook'u `.go` dosyalarında `//go:` direktiflerini (`//go:embed`, `//go:build`) yorum saymaz. Diğer `//` satırları yasak kalır.
- Repoda iki dil olur. Uygulama Go, kabul testleri TypeScript.
- Çok kullanıcılı bir deploy SQLite'ı aşarsa veritabanı kararı yeni bir ADR ile yeniden açılır.
