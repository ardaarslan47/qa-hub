# QA Hub

Case'lerin yazıldığı, run'lar içinde çalıştırıldığı ve status'larının kaydedildiği basit bir test yönetim uygulaması. Projenin amacı ve açık soruları `docs/concept.md`'de, terimler `GLOSSARY.md`'de.

## Gereksinimler

- Go, `go.mod`'daki sürüm.
- Node.js ve npm, Playwright testleri için.
- `make`.

templ ayrıca kurulmaz. `go.mod`'da tool olarak tanımlı, `go tool templ` ile çalışır.

## Kontrolleri çalıştır

CI'ın `main`'e her push'ta koştuğu kontrollerin aynısını lokalde çalıştırmak için:

```sh
make check
```

`make check` sırasıyla `templ generate`, `go vet ./...` ve `go test ./...` çalıştırır, sonra Playwright senaryolarını Chromium'da koşar. İlk çalıştırmada `e2e/` altındaki npm paketlerini ve Chromium'u kurar.

Playwright her senaryo için uygulamayı ayrı bir process olarak, boş bir SQLite dosyasıyla başlatır. Senaryolar birbirinin verisini görmez, herhangi bir sırayla ya da tek başına koşabilir. Tek bir senaryoyu koşmak için:

```sh
cd e2e && npx playwright test -g "Boş başlık"
```

`npx playwright test` uygulamayı önce `make build` ile derler.

## Uygulamayı başlat

```sh
make run
```

Uygulama `http://127.0.0.1:8080` adresinde açılır ve veriyi çalıştığı dizindeki `qa-hub.db` dosyasına yazar. Adresi ve dosyayı değiştirmek için binary'yi doğrudan çalıştır:

```sh
./bin/qa-hub -addr 127.0.0.1:9000 -db /tmp/qa-hub.db
```

## Kod nerede

- `cmd/qa-hub`: HTTP server'ı başlatan binary.
- `internal/qahub`: domain tipleri, kuralları ve SQLite'a yazan `Store`. Migration'lar `internal/qahub/migrations/` altında, `0001_ad.sql` biçiminde sıralı SQL dosyaları. `Open` uygulanmamış olanları sırayla uygular.
- `internal/web`: HTTP handler'ları ve templ şablonları. `go test` kabul testleri de burada, uygulamayı process içinde HTTP üzerinden sürer.
- `e2e`: Playwright kabul senaryoları, TypeScript.
