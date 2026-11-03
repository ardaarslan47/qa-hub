# QA Hub: Konsept

Bu doküman projenin niyetini ve açık sorularını tutar. Kesinleşen terimler `GLOSSARY.md`'ye, kararlar `docs/adr/`'ye yazılır.

## Neden var?

QA Hub, AFK agent workflow'larını denemek için kurulmuş bir deney alanı. Amaç, AI ile çalışma akışını uçtan uca anlamak ve bu akışla neler inşa edilebileceğini görmek. Örneğin bir issue `ready-for-agent` etiketi aldığında agent'ın onu kendi başına alıp bitirmesi. Repo bozulabilir, bozulması önemli değil.

Ürün ikinci planda. Asıl işlevi agent'lara üzerinde çalışabilecekleri somut bir hedef vermek. Sonunda düzgün bir şey çıkarsa gerçekten kullanılabilir.

## Ürün

QA Hub, TestRail ve Zephyr'den esinlenen basit bir test yönetim uygulaması. Test case'ler yazılır, bir koşu (run) içinde çalıştırılır ve sonuçlar kaydedilir.

Temel fikir: Case bir **tanım**. Bir koşudaki her çalıştırma o tanımdan doğan ayrı bir **kayıt** ve sonuçlar bu kayda yazılır.

## Kısıtlar

- Bağımsız bir proje. Başka bir repoya ya da ortama bağlı değil.
- Public GitHub reposu. CI GitHub Actions'ta koşar.
- Web uygulaması. Gereken altyapıyı (veritabanı vb.) agent'lar kurar.

## Açık sorular

- **Tech stack:** Henüz seçilmedi. İlk adımda seçilecek.
- **AFK agent aracı:** Issue'ları otomatik işleyecek döngünün nasıl kurulacağı belli değil. Sandcastle bir örnek.
- **Deney alanı duruşu:** Agent'lara geriye dönük uyumluluk ve cesur refactor konusunda ne söyleneceği, QA Hub'daki ilk oturumda konuşulacak.
