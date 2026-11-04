# 0003: Project'ten bağımsız tek Library

**Durum:** Kabul edildi, 2026-11-04

## Bağlam

Project ilk tasarımda case'lerin ve run'ların ait olduğu izole bir alandı. Aynı case iki project'te gerekince iki ayrı case olarak yazılıyor, aralarında bağ kalmıyordu (#7).

İncelediğimiz test yönetim tool'larında case deposu genellikle project'e bağlı. TestRail'de repository project içinde duruyor. Bu yüzden project'ten bağımsız bir depo alışılmış yolun dışında ve okuyan birinin "düzeltmek" isteyebileceği bir tercih.

Değerlendirilen seçenekler:

- Her project'in kendi Library'si: tool'ların modeli, ama paylaşım kopyalamaya döner.
- Project'i tamamen kaldırmak: run'ların gruplanacağı yer kalmaz.
- Project'in kendi case listesini tutması: case'i project'e geri bağlar.

## Karar

- Bütün case'ler tek bir Library'de durur. Case hiçbir project'e ait değildir. Project sadece run'ları tutar.
- Library düzdür, Folder ya da Label yoktur.
- Run başladığı anda Library'deki her case için bir execution içerir. Bu kural geçicidir, seçim ertelendi.

## Sonuçlar

- Farklı app'lerin case'leri aynı listede durur ve aynı run'a girer. Library büyüyünce run'a case seçimi ve düzenleme aracı yeniden konuşulur.
- Run'a hangi case'lerin gireceği kararı, Playwright ve benzeri tool'lardan case ve run alma fikriyle birlikte verilecek. Bu bağlantının project ile Library'yi nasıl eşleyeceği açık.
- Case'ler project'ten bağımsız kaldığı için Journey gibi birden fazla case'i sıralayan kavramlar sonradan eklenebilir.
