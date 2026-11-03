# QA Hub

Test case'lerin yazıldığı, run'lar içinde çalıştırıldığı ve status'larının kaydedildiği basit bir test yönetim uygulaması. Projenin amacı `docs/concept.md`'de, terimler `GLOSSARY.md`'de.

## Gereksinimler

- Go (sürüm `go.mod`'da)
- Node 24 ya da üstü ve npm
- `make`

Go'yu mise ile kuruyorsan ve global sürüm seçmediysen komutların başına `mise exec go@1.26.10 --` koy.

## Kontroller

```
make check
```

CI'ın `main`'e her push'ta koştuğu kontrollerin aynısı, sırayla:

1. `templ generate`
2. `go vet ./...`
3. `go test ./...`
4. Uygulamayı `bin/qa-hub`'a derler.
5. Playwright senaryoları. Her senaryo kendi uygulama process'ini ve temiz bir geçici SQLite dosyasını ayağa kaldırır. Chromium'u ilk koşuda indirir.

Tek tek koşmak için `make generate`, `make vet`, `make test`, `make build`, `make e2e`.

templ'in ürettiği `*_templ.go` dosyaları commit edilmez (ADR 0002). Temiz bir clone'da `go build`, `go vet` ve `go test` önce `make generate` ister. `make` hedefleri bunu kendisi yapar.

## Uygulamayı çalıştırma

```
make build
./bin/qa-hub -addr 127.0.0.1:8080 -db qa-hub.db
```

Veritabanı dosyası yoksa oluşturulur, migration'lar açılışta uygulanır. Sonra http://127.0.0.1:8080 adresini aç.
