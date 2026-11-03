# Domain dokümanları

Engineering skill'leri kod tabanını keşfederken bu reponun domain dokümanlarını nasıl kullanmalı.

## Keşfe başlamadan önce bunları oku

- Repo kökündeki **`GLOSSARY.md`**, ya da
- Varsa repo kökündeki **`GLOSSARY-MAP.md`**: her context için bir `GLOSSARY.md`'yi gösterir. Konuyla ilgili olanların her birini oku.
- **`docs/adr/`**: üzerinde çalışacağın alana dokunan ADR'leri oku. Multi-context repolarda context'e özel kararlar için `src/<context>/docs/adr/` altına da bak.

Bu dosyalardan biri yoksa **sessizce devam et**. Eksik olduklarını belirtme, önceden oluşturulmalarını önerme. `/domain-modeling` skill'i (`/grill-with-docs` ve `/improve-codebase-architecture` üzerinden çağrılır) bunları, terimler ya da kararlar gerçekten netleştiğinde oluşturur.

## Dosya yapısı

Single-context repo (çoğu repo):

```
/
├── GLOSSARY.md
├── docs/adr/
│   ├── 0001-event-sourced-orders.md
│   └── 0002-postgres-for-write-model.md
└── src/
```

Multi-context repo (kökte `GLOSSARY-MAP.md` varsa):

```
/
├── GLOSSARY-MAP.md
├── docs/adr/                          ← sistem genelindeki kararlar
└── src/
    ├── ordering/
    │   ├── GLOSSARY.md
    │   └── docs/adr/                  ← context'e özel kararlar
    └── billing/
        ├── GLOSSARY.md
        └── docs/adr/
```

## Glossary'nin kelime dağarcığını kullan

Çıktın bir domain kavramına ad verdiğinde (issue başlığı, refactor önerisi, hipotez, test adı), terimi `GLOSSARY.md`'de tanımlandığı haliyle kullan. Glossary'nin özellikle kaçındığı eş anlamlılara kayma.

İhtiyacın olan kavram glossary'de henüz yoksa bu bir sinyaldir: ya projenin kullanmadığı bir dil icat ediyorsundur (yeniden düşün), ya da gerçek bir boşluk vardır (`/domain-modeling` için not al).

## ADR çelişkilerini belirt

Çıktın mevcut bir ADR ile çelişiyorsa, sessizce üzerine yazmak yerine bunu açıkça belirt:

> _ADR-0007 (event-sourced orders) ile çelişiyor, ama yeniden açmaya değer çünkü…_
