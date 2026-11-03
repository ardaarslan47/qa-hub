# Branch modeli: Trunk-based mi, versiyon branch'leri mi?

> Araştırma tarihi: 2026-11-03. Her iddianın yanında kaynağı var. "Doğrulanamadı" yazan yerler, birincil kaynağa ulaşılamayan iddialar.

**İncelenen model (düzeltilmiş hali):** Kalıcı bir `master`/`main` yok. Branch'ler versiyon adı taşıyor (`v0`, `v1`, `v1.1`, `v2`). Mevcut bir versiyondan yeni bir versiyon branch'i açılıyor (örneğin `v1`'den `v1.1`). Kullanıcı ve agent'lar bu branch'e PR olmadan, doğrudan push ediyor. Biriken iş en sonda toplu kontrol ediliyor ve bug'lar toplu fix'leniyor. Sonra iki yol var: (a) bu branch'ten bir sonraki versiyon açılıyor (`v1.1` → `v2`), ya da (b) branch geri merge ediliyor (`v1.1` → `v1`).

## TL;DR

- **Bu model trunk-based değil.** Trunk-based development tek bir kalıcı branch (trunk) üzerinde çalışmayı ve başka uzun ömürlü geliştirme branch'i açmamayı şart koşuyor ([trunkbaseddevelopment.com](https://trunkbaseddevelopment.com/)). Bu modelde kalıcı tek bir trunk yok. Her versiyon branch'i birkaç gün değil, bir versiyon boyunca yaşıyor.
- **En yakın isimli pattern'ler:** (b) yolu, trunkbaseddevelopment.com'un "**Mainline**" (ClearCase tarzı) diye adlandırdığı modele denk geliyor. Kalıcı bir branch'ten iş branch'i açılıyor, iş bitince "büyük bir merge" ile geri dönülüyor ([alternative branching models](https://trunkbaseddevelopment.com/alternative-branching-models/)). (a) yolu, yani her versiyon branch'inin bir sonrakinin tabanı olması için birincil kaynakta isimli bir pattern bulamadım. Doğrulanamadı.
- **Ama kullanıcının asıl niyeti trunk-based ile uyumlu.** "Doğrudan push, PR yok" kısmı trunk-based'in küçük ekipler için tarif ettiği "doğrudan trunk'a commit" stiliyle aynı ([trunkbaseddevelopment.com](https://trunkbaseddevelopment.com/)). Farklı olan tek şey, versiyonun branch adıyla tutulması. Versiyonu branch yerine tag ve GitHub Release ile tutmak bu ihtiyacı karşılıyor ([About releases](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases)).
- **GitHub'da versiyon branch'i modeli sürtünme yaratıyor.** `Closes #N` yalnızca default branch'te çalışıyor, scheduled workflow'lar yalnızca default branch'te koşuyor, `gh pr create` varsayılan olarak default branch'i hedefliyor. Aktif branch her versiyonda değiştiği için default branch'i de her seferinde değiştirmek gerekiyor.
- **Öneri (benim değerlendirmem):** Tek kalıcı trunk `main` olsun. Kullanıcı ve agent orkestratörü `main`'e doğrudan push etsin. Toplu kontrol ve toplu fix `main` üzerinde yapılsın, sonra `v1.1.0` gibi bir tag ve GitHub Release ile versiyonlansın. Agent'lar stabilizasyon sırasında çalışmaya devam edecekse trunk'tan kısa ömürlü bir `release/v1.1` branch'i kesilsin. Agent'lara ayrı bir GitHub kimliği verilsin ve ruleset'lerle tag oluşturma ve `main`'i silme/force push engellensin.

---

## 1. Trunk-based development nedir ve nasıl uygulanır?

### 1.1 Tanım

| Kavram | Kaynağın söylediği | Kaynak |
|---|---|---|
| Tek trunk | Geliştiriciler "trunk" adlı tek bir branch üzerinde ortak çalışır ve başka uzun ömürlü geliştirme branch'leri açma baskısına direnir. | [trunkbaseddevelopment.com](https://trunkbaseddevelopment.com/) |
| Trunk'ın adı | Git topluluğunda 2020'den beri `main`, öncesinde `master`. | [trunkbaseddevelopment.com](https://trunkbaseddevelopment.com/) |
| Küçük batch'ler | Her geliştirici işini küçük batch'lere böler ve trunk'a merge eder. | [DORA](https://dora.dev/capabilities/trunk-based-development/) |
| Branch sayısı | Repoda üç ya da daha az aktif branch. | [DORA](https://dora.dev/capabilities/trunk-based-development/) |
| Merge sıklığı | Branch'ler trunk'a en az günde bir kez merge edilir. | [DORA](https://dora.dev/capabilities/trunk-based-development/) |
| Code freeze | Code freeze ve entegrasyon fazı yok. | [DORA](https://dora.dev/capabilities/trunk-based-development/) |
| CI ile ilişki | Trunk-based development, continuous integration için zorunlu bir pratik. | [DORA](https://dora.dev/capabilities/trunk-based-development/) |
| Fowler'ın özeti | Tüm iş mainline'da ("trunk") yapılır, uzun ömürlü branch'lerden kaçınılır. Küçük ekipler doğrudan mainline'a commit eder. | [Fowler, Branching Patterns](https://martinfowler.com/articles/branching-patterns.html#Trunk-basedDevelopment) |

### 1.2 Doğrudan commit ve kısa ömürlü feature branch

| Stil | Nasıl çalışır | Kaynak |
|---|---|---|
| Doğrudan trunk'a commit | "Çok küçük ekipler doğrudan trunk'a commit edebilir." Küçük ekipler için daha geleneksel seçenek. | [trunkbaseddevelopment.com](https://trunkbaseddevelopment.com/), [short-lived feature branches](https://trunkbaseddevelopment.com/short-lived-feature-branches/) |
| Kısa ömürlü feature branch | Branch en fazla bir iki gün yaşar. Tek bir geliştiriciye (eşli çalışılıyorsa ikiye) aittir. Trunk'a merge yalnızca branch kapatılırken yapılır, sonra branch silinir. | [short-lived feature branches](https://trunkbaseddevelopment.com/short-lived-feature-branches/) |
| AI ağırlıklı ekipler | Agentic işte birim "bir geliştirici ve onun agent'ları". Kısa ömürlü branch kullanılmıyorsa alternatif, trunk'a bir commit queue arkasından doğrudan commit etmek. | [short-lived feature branches](https://trunkbaseddevelopment.com/short-lived-feature-branches/) |
| DORA'nın branch ömrü | Branch'ler genellikle birkaç saatten uzun yaşamaz. | [DORA](https://dora.dev/capabilities/trunk-based-development/) |

Fowler'a göre feature branching ile CI arasındaki fark, branch olup olmaması değil. Fark, mainline ile ne zaman entegre olunduğu ([Fowler](https://martinfowler.com/articles/branching-patterns.html#trunk-based-development)).

### 1.3 Release: "branch for release" ve "release from trunk"

| Yaklaşım | Nasıl çalışır | Kaynak |
|---|---|---|
| Branch for release | Release'ten birkaç gün önce trunk'tan "just in time" release branch'i kesilir. Bug önce trunk'ta test ile fix'lenir, sonra release branch'ine cherry-pick edilir. Release branch'ine doğrudan commit yapılmaz. Branch trunk'a geri merge edilmez. İş bitince silinir, silmeden önce release commit'i tag'lenir. | [branch for release](https://trunkbaseddevelopment.com/branch-for-release/) |
| Release from trunk | Çok sık release yapan ekipler release branch'i kullanmaz, doğrudan trunk'tan release eder. Bug'ı trunk'ta "roll forward" ile, bir feature gibi fix'ler. | [release from trunk](https://trunkbaseddevelopment.com/release-from-trunk/) |

### 1.4 Feature flags ve CI

| Konu | Kaynağın söylediği | Kaynak |
|---|---|---|
| Feature flags | Bitmemiş iş kapalı halde trunk'a ve production'a gidebilir, sonra açılır. | [feature flags](https://trunkbaseddevelopment.com/feature-flags/) |
| CI'ın rolü | "Bireyler trunk-based development uygular, ekipler CI uygular." Üç kişiden büyük ekiplerde trunk'a gelen commit'leri doğrulayan bir CI daemon'u gerekir. En verimli ekipler trunk'ın kırılmasını önleyen CI kurulumuna sahiptir. Sık kırmızı olan build değerini kaybeder. | [continuous integration](https://trunkbaseddevelopment.com/continuous-integration/) |
| Commit öncesi test | Testler commit'ten önce koşmalı, yoksa trunk çalışır halde kalmayabilir. | [DORA](https://dora.dev/capabilities/trunk-based-development/) |
| Healthy Branch | Her commit'te otomatik kontroller (build ve test) koşar. | [Fowler](https://martinfowler.com/articles/branching-patterns.html#healthy-branch) |

---

## 2. Kullanıcının modeli hangi pattern'e denk geliyor?

### 2.1 Aday pattern'ler

| Pattern | Kaynaktaki tanım | Bu modelle ilişkisi (benim değerlendirmem) |
|---|---|---|
| Trunk-based development | Tek trunk, başka uzun ömürlü geliştirme branch'i yok ([trunkbaseddevelopment.com](https://trunkbaseddevelopment.com/)). | **Denk gelmiyor.** Kalıcı tek trunk yok. Versiyon branch'leri uzun ömürlü. Sadece "doğrudan push" kısmı benziyor. |
| TBD.com "Mainline" (ClearCase tarzı) | "Sonsuza kadar yaşayan" bir mainline var. Ekipler ondan iş branch'i açar. İş bitince o branch'ten release yapılabilir ve mainline'a "büyük" bir merge yapılır. Yazarın hükmü: trunk-based'in tam tersi, yapmayın ([alternative branching models](https://trunkbaseddevelopment.com/alternative-branching-models/)). | **(b) yolu buna denk geliyor.** `v1` kalıcı branch, `v1.1` iş branch'i, sonunda `v1.1` → `v1` toplu merge. |
| TBD.com "Cascade" | Her release'in kendi branch'i var ve her release ekibi "upstream" branch'ten her gün merge alır. Hüküm: trunk-based ile uyumsuz ([alternative branching models](https://trunkbaseddevelopment.com/alternative-branching-models/)). | **Kısmen.** "Her versiyonun kendi branch'i" kısmı benziyor. Ama bu modelde aynı anda birden çok versiyon branch'inde paralel geliştirme yok. |
| Fowler "Release Branch" | Yalnızca bir versiyonu release'e hazırlamak için stabilize eden commit'leri kabul eden branch ([Fowler](https://martinfowler.com/articles/branching-patterns.html#release-branch)). Mainline'dan kopyalanır, yeni feature eklenmez. | **Denk gelmiyor.** `v1.1` yeni feature alıyor, sadece stabilizasyon değil. Ama toplu kontrol fazı tek başına bir release branch'i işine benziyor. |
| Fowler "Maturity Branch" | Head'i kod tabanının belirli bir olgunluk seviyesinin son halini gösteren branch. Alternatifi tag şeması. Fowler'a göre güçlü fayda ya da maliyeti yok ([Fowler](https://martinfowler.com/articles/branching-patterns.html#maturity-branch)). | **Kapanmış eski versiyon branch'leri bunun gibi davranıyor.** Ama bu modelde hiçbir branch kalıcı olarak "en son stabil" anlamı taşımıyor. |
| Fowler "Environment Branch" | Ürünü yeni bir ortamda çalıştırmak için commit uygulanan branch. Fowler buna klasik anti-pattern diyor ([Fowler](https://martinfowler.com/articles/branching-patterns.html#environment-branch)). | **Denk gelmiyor.** Branch'ler ortamı değil versiyonu temsil ediyor. |
| Fowler "Feature Branching" | Bir feature'ın tüm işi kendi branch'inde yapılır, bitince mainline'a entegre edilir ([Fowler](https://martinfowler.com/articles/branching-patterns.html#feature-branching)). | **(b) yolunda `v1.1`, bir versiyon boyunca yaşayan çok büyük bir feature branch'i gibi davranıyor.** |
| git-flow | İki ana branch: `master` her zaman production'a hazır, `develop` bir sonraki release'in en güncel hali. Release branch'leri `develop`'tan açılır, `master`'a merge edilip tag'lenir ([nvie.com](https://nvie.com/posts/a-successful-git-branching-model/)). Fowler'a göre git-flow'da `develop` mainline, `master` "Production Maturity Branch" ([Fowler](https://martinfowler.com/articles/branching-patterns.html#Git-flow)). | **Denk gelmiyor.** Versiyon başına branch var ama kalıcı `develop`/`master` ikilisi yok. |

**"Version/maintenance branch" ve "branch per release" adları:** Bu adları trunkbaseddevelopment.com'un alternatif modeller sayfasında ve Fowler'ın yazısında isimli bir pattern olarak bulamadım. Fowler, müşteride eski sürümleri yaşayan yazılımlar için her sürüme açık bir release branch'i tutulduğunu ve fix'lerin release branch'inde yapılıp mainline'a merge edildiğini anlatıyor ([Fowler](https://martinfowler.com/articles/branching-patterns.html#release-branch)). Ama orada da kalıcı bir mainline var. Doğrulanamadı: (a) yolunun (`v1.1` → `v2`, geri merge yok) yerleşik bir adı.

**git-flow yazarının sonradan eklediği not (5 Mart 2020):** Yazar, git-flow'un bir "dogma ya da her derde deva" gibi görülmeye başladığını söylüyor. Continuous delivery yapan ekiplere GitHub flow gibi çok daha basit bir akış öneriyor. Açıkça versiyonlanan ya da sahada birden çok sürümü desteklenen yazılımlar için git-flow'un hâlâ uygun olabileceğini ekliyor ([nvie.com](https://nvie.com/posts/a-successful-git-branching-model/)).

### 2.2 Kalıcı tek bir mainline olmamasının sonuçları

Kaynaklı noktalar:

- Fowler, büyük merge'lerin asıl sorununun iş miktarı değil belirsizlik olduğunu söylüyor. Sık entegrasyon merge sayısını artırır ama karmaşıklığı ve riski azaltır ([Fowler](https://martinfowler.com/articles/branching-patterns.html#integration-frequency)).
- Release branch'leri birbirinden uzaklaştıkça mainline'a geri merge etmek zorlaşıyor ([Fowler](https://martinfowler.com/articles/branching-patterns.html#release-branch)).
- TBD.com, Mainline modelinin geliştirici sayısı arttıkça kötüleştiğini söylüyor ([alternative branching models](https://trunkbaseddevelopment.com/alternative-branching-models/)).

Benim değerlendirmem:

- **"Şu an neredeyiz?" sorusunun cevabı branch adına bağlı.** Agent'ların, CI'ın ve GitHub'ın hangi branch'in aktif olduğunu bilmesi gerekiyor. Her versiyonda bu bilgi değişiyor (bkz. Bölüm 4.1).
- **(a) yolunda branch'ler aslında tag gibi davranıyor.** `v1.1` → `v2` açıldıktan sonra `v1.1`'e kimse dokunmuyorsa, `v1.1` sadece bir commit'i işaret eden bir isim. Bu, tag'in işi. Fowler da Maturity Branch için tag'i alternatif olarak gösteriyor ([Fowler](https://martinfowler.com/articles/branching-patterns.html#maturity-branch)).
- **(b) yolunda merge riski versiyon süresiyle büyüyor.** `v1` beklerken `v1.1` uzun süre ayrı yaşarsa, ama `v1`'e başka bir şey girmiyorsa, merge aslında bir fast-forward'dan ibaret. O zaman `v1.1` ile `v1` ayrımı da tag ile aynı işi yapıyor.

### 2.3 Orijinal anlatım (`master` + `v0`) hakkında kısa not

Kullanıcının ilk anlattığı model (`v0`'a doğrudan push, stabil olunca `master`'a almak), git-flow'daki `develop`/`master` ayrımına ve Fowler'ın "Production Maturity Branch" tanımına daha yakındı ([Fowler](https://martinfowler.com/articles/branching-patterns.html#maturity-branch)). Trunk-based'deki release branch'i ise trunk'tan kesilir, sadece stabilizasyon alır ve trunk'a geri merge edilmez ([branch for release](https://trunkbaseddevelopment.com/branch-for-release/)). Düzeltilmiş modelde `master` olmadığı için bu not sadece bağlam için.

---

## 3. Tek kişilik proje ve AFK agent workflow'u için hangi model uygun?

> Bu bölüm büyük ölçüde benim değerlendirmem. Kaynaklı cümleler link taşıyor.

### 3.1 Sandcastle agent'ları nasıl çalıştırıyor?

Birincil kaynak: [github.com/mattpocock/sandcastle](https://github.com/mattpocock/sandcastle). npm paketi `@ai-hero/sandcastle`.

| Konu | Kaynağın söylediği | Kaynak |
|---|---|---|
| Genel | Agent'ı sandbox'ta çalıştırır, branch stratejisi ayarlanabilir. Branch'lerde yapılan commit'ler geri merge edilir. | [README](https://github.com/mattpocock/sandcastle/blob/main/README.md) |
| `head` stratejisi | Agent doğrudan host'un working directory'sine yazar. Worktree ya da branch açılmaz. Docker gibi bind-mount provider'larda varsayılan. | [branch-strategies.md](https://github.com/mattpocock/sandcastle/blob/main/docs/guide/branch-strategies.md) |
| `merge-to-head` stratejisi | Bir git worktree'de geçici bir branch açılır. Run bitince değişiklikler o an checkout edilmiş branch'e merge edilir. Merge başarısızsa abort edilir, branch'iniz değişmez. İzole provider'larda varsayılan. | [branch-strategies.md](https://github.com/mattpocock/sandcastle/blob/main/docs/guide/branch-strategies.md) |
| `branch` stratejisi | Commit'ler bir worktree'de, açıkça adı verilen bir branch'e düşer. | [branch-strategies.md](https://github.com/mattpocock/sandcastle/blob/main/docs/guide/branch-strategies.md) |
| `simple-loop` şablonu | Issue'ları tek tek alır ve kapatır. `merge-to-head` kullanır. | [init.md](https://github.com/mattpocock/sandcastle/blob/main/docs/guide/init.md), [simple-loop/main.mts](https://github.com/mattpocock/sandcastle/blob/main/src/templates/simple-loop/main.mts) |
| `parallel-planner` şablonu | Paralel yapılabilecek issue'ları planlar. Her agent kendi branch'inde çalışır ("execution sırasında conflict olmasın diye"). Sonra tek bir agent tüm branch'leri **o anki branch'e** `git merge` ile birleştirir, conflict'leri çözer, typecheck ve test koşar, issue'ları kapatır. | [init.md](https://github.com/mattpocock/sandcastle/blob/main/docs/guide/init.md), [parallel-planner/main.mts](https://github.com/mattpocock/sandcastle/blob/main/src/templates/parallel-planner/main.mts), [merge-prompt.md](https://github.com/mattpocock/sandcastle/blob/main/src/templates/parallel-planner/merge-prompt.md) |
| Remote'a push | Branch strategies sayfası push'u anlatmıyor. `simple-loop` ve `parallel-planner` şablon dosyalarında `git push` ya da PR açma adımı bulamadım (dosyaları taradım). | Doğrulanamadı: Sandcastle'ın push'u kendisinin yapıp yapmadığı. |

Çıkarım (benim değerlendirmem): Sandcastle paralel agent'ları ayrı branch/worktree'de çalıştırıyor ve **entegrasyonu lokal olarak, sırayla, tek bir noktada** yapıyor. Remote'a push eden tek aktör orkestratörün çalıştığı makine oluyor. Yani "agent'lar aynı remote branch'e paralel push ediyor" durumu Sandcastle ile doğal olarak oluşmuyor.

### 3.2 Agent'lar aynı branch'e paralel push ederse ne olur?

Git, varsayılan olarak fast-forward olmayan bir push'u reddeder. Diğer tarafın commit'lerini kaybetmemek için önce fetch, sonra merge ya da rebase, sonra tekrar push gerekir ([git-push, "Note about fast-forwards"](https://git-scm.com/docs/git-push#_note_about_fast_forwards)).

Benim değerlendirmem: Bu, branch'in adı `main` da olsa `v1.1` da olsa aynı. Yani paralel push sorunu branch modelinden bağımsız. Çözüm, ya Sandcastle'daki gibi tek bir merge noktası, ya da her agent'ın push'tan önce rebase etmesi. TBD.com'un AI ağırlıklı ekipler için bahsettiği "commit queue arkasından trunk'a doğrudan commit" ([short-lived feature branches](https://trunkbaseddevelopment.com/short-lived-feature-branches/)) da aynı soruna verilen cevap.

### 3.3 "Her değişikliği değil, toplu kontrol" gerekçesi

| | Değerlendirme |
|---|---|
| **Artı** | Her agent değişikliğinde insan review'u darboğaz. TBD.com da kısa ömürlü branch'lerin darboğazı insan review'una taşıdığını, doğrudan commit'in ise test ve build'in hızına dayandığını söylüyor ([short-lived feature branches](https://trunkbaseddevelopment.com/short-lived-feature-branches/)). DORA da ağır ve asenkron code review'u yaygın bir tuzak olarak sayıyor ([DORA](https://dora.dev/capabilities/trunk-based-development/)). Yani "her değişikliğe PR yok" kararı trunk-based ile çelişmiyor. |
| **Artı** | Repo bir deney alanı ve bozulabilir ([concept.md](../concept.md)). Toplu kontrolün maliyeti düşük. |
| **Artı** | Tek kişi olunca "entegrasyon fazı" diğer insanları bekletmiyor. Fowler'ın ve DORA'nın çekincelerinin çoğu ekipler arası bekleşmeyle ilgili. |
| **Eksi** | DORA code freeze ve entegrasyon fazı olmamasını öneriyor ([DORA](https://dora.dev/capabilities/trunk-based-development/)). Toplu kontrol fazı, adı konmamış bir entegrasyon fazı. |
| **Eksi** | Toplu kontrolde bulunan bug'ın hangi agent değişikliğinden geldiğini bulmak zorlaşıyor. Fowler'ın büyük merge'deki "belirsizlik" uyarısı burada da geçerli ([Fowler](https://martinfowler.com/articles/branching-patterns.html#integration-frequency)). |
| **Eksi** | Agent'lar kırık bir branch üzerine iş yığmaya devam edebilir. TBD.com sık kırmızı olan build'in değerini kaybettiğini söylüyor ([continuous integration](https://trunkbaseddevelopment.com/continuous-integration/)). Bu yüzden insan review'u toplu olsa bile **otomatik CI her push'ta koşmalı**. |
| **Nötr** | Toplu kontrolün değeri branch modelinden gelmiyor. Aynı "biriktir, kontrol et, fix'le, versiyonla" döngüsü tek trunk + tag ile de kurulabiliyor (bkz. Bölüm 4.5). |

### 3.4 Tag mi, versiyon branch'i mi?

| Ölçüt | Versiyon branch'leri (`v1`, `v1.1`, `v2`) | Tek trunk + tag (`main` + `v1.1.0`) | Tek trunk + stabilizasyon için release branch |
|---|---|---|---|
| Aktif branch nerede? | Her versiyonda değişiyor | Hep `main` | Hep `main`, stabilizasyon sırasında ek olarak `release/v1.1` |
| GitHub default branch | Her versiyonda değiştirilmeli (Bölüm 4.1) | Bir kez `main` | Bir kez `main` |
| `Closes #N` auto-close | Sadece default branch'e girince | `main`'e girince | `main`'e girince |
| Versiyon işareti | Branch adı (hareket edebilir) | Tag. GitHub Release'leri tag'e dayanır ([About releases](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases)) | Tag, release branch'inden |
| Toplu kontrol sırasında agent'lar | Aynı branch'e push etmeye devam ederse kontrol edilen şey kayar | Agent kuyruğu durdurulmalı ya da kontrol edilen commit sabitlenmeli | Agent'lar `main`'de devam eder, kontrol `release/v1.1`'de yapılır ([branch for release](https://trunkbaseddevelopment.com/branch-for-release/)) |
| Bug fix yeri | Versiyon branch'i | `main` | `main`'de fix, release branch'ine cherry-pick ([branch for release](https://trunkbaseddevelopment.com/branch-for-release/)) |
| Eski versiyona dönüş | Branch'e checkout | Tag'e checkout. Gerekirse tag'den branch açılır | Tag'e checkout |
| Kaynakların hükmü | Mainline/Cascade: "yapmayın" ([TBD.com](https://trunkbaseddevelopment.com/alternative-branching-models/)) | Release from trunk ([TBD.com](https://trunkbaseddevelopment.com/release-from-trunk/)) | Branch for release ([TBD.com](https://trunkbaseddevelopment.com/branch-for-release/)) |

Benim değerlendirmem: Kullanıcının niyetini (PR yok, toplu kontrol, sonra versiyon) en az sürtünmeyle "tek trunk + tag" karşılıyor. Agent'ların stabilizasyon sırasında da çalışması isteniyorsa "tek trunk + kısa ömürlü release branch" eklenebilir. Tek kişilik bir deney reposunda büyük ihtimalle ilki yeterli.

---

## 4. GitHub tarafında konfigürasyon

### 4.1 Default branch

| Konu | GitHub'ın söylediği | Versiyon branch'i modeline etkisi (benim değerlendirmem) |
|---|---|---|
| Default branch'in rolü | Default branch, pull request'lerin ve code commit'lerin taban branch'i ([Changing the default branch](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-branches-in-your-repository/changing-the-default-branch)). | Aktif versiyon default olmazsa repo sayfası eski kodu gösterir. |
| Değiştirme şartı | Repoda birden fazla branch olmalı ve admin yetkisi gerekir ([Changing the default branch](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-branches-in-your-repository/changing-the-default-branch)). | Her yeni versiyonda elle (ya da API ile) değiştirmek gerekir. |
| `Closes #N` | Closing keyword'ler yalnızca default branch'i hedefleyen PR'larda çalışır. Başka branch'i hedefleyen PR'da keyword'ler yok sayılır. Commit mesajındaki keyword, commit default branch'e merge edildiğinde issue'yu kapatır ([Linking a PR to an issue](https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/linking-a-pull-request-to-an-issue)). | Aktif versiyon default değilse issue'lar kapanmaz. (b) yolunda issue'lar ancak `v1.1` → `v1` merge'inde kapanır. Not: Sandcastle'ın merge prompt'u issue'ları ayrıca bir komutla kapatıyor ([merge-prompt.md](https://github.com/mattpocock/sandcastle/blob/main/src/templates/parallel-planner/merge-prompt.md)), yani agent akışı auto-close'a bağımlı olmayabilir. |
| Doğrudan push ile auto-close | Doküman "merge" diyor, doğrudan push'u ayrıca anlatmıyor. | Doğrulanamadı: default branch'e doğrudan push edilen commit'teki `Closes #N`'in issue'yu kapattığı (yaygın olarak böyle çalıştığı biliniyor ama dokümanda açık cümle bulamadım). |
| Actions `schedule` | Scheduled workflow'lar default branch'teki son commit üzerinde, yalnızca default branch'te koşar. Public repoda 60 gün aktivite olmazsa otomatik devre dışı kalır ([Events that trigger workflows](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#schedule)). | Default branch eski versiyonda kalırsa nightly işler eski kodu test eder. |
| `workflow_dispatch` butonu | UI'daki "Run workflow" butonu, workflow dosyası default branch'te varsa görünür ([Events that trigger workflows](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#workflow_dispatch)). | Yeni bir workflow sadece aktif versiyon branch'inde varsa UI'da görünmez. |
| `gh pr create` | `--base` verilmezse önce `gh-merge-base` git config'i, o da yoksa reponun default branch'i kullanılır (`gh pr create --help`, gh 2.102.0). | Agent'lar PR açarsa yanlış branch'i hedefleyebilir. |

### 4.2 Rulesets ve branch protection

| Konu | GitHub'ın söylediği | Kaynak |
|---|---|---|
| Free planda kullanılabilirlik | Rulesets, GitHub Free ile public repolarda kullanılabilir. Protected branches de öyle. | [About rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets), [About protected branches](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches) |
| Branch protection ile birlikte | İkisi yan yana çalışır, örtüşen kurallarda en kısıtlayıcı olan uygulanır. | [About rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets) |
| Hedef | Branch'leri ya da tag'leri hedefleyebilir. Repo başına 75 ruleset. | [About rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets) |
| Pattern | fnmatch sözdizimi. Include ve exclude hedefleri birleştirilebilir. `*` dizin ayırıcı (`/`) hariç her şeyle eşleşir. | [Creating rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/creating-rulesets-for-a-repository) |
| Bypass | Roller (repository admin, maintain, write), takımlar, GitHub App'ler, Dependabot. Modlar: "Always allow" ve "For pull requests only". | [Creating rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/creating-rulesets-for-a-repository) |
| Enforcement | Active, Disabled. "Evaluate" dokümanda sadece metadata kısıtları bağlamında geçiyor. | [Creating rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/creating-rulesets-for-a-repository) |
| Restrict creations / updates / deletions | Eşleşen branch ya da tag'leri yalnızca bypass yetkisi olanlar oluşturabilir / push edebilir / silebilir. | [Available rules](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/available-rules-for-rulesets) |
| Block force pushes | Hedef branch ya da tag'lere force push engellenir. | [Available rules](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/available-rules-for-rulesets) |
| Require a pull request | Tüm değişikliklerin bir PR ile ilişkili olmasını şart koşar. İzin verilen merge yöntemi kısıtlanabilir. | [Available rules](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/available-rules-for-rulesets) |
| Required status checks ve doğrudan push | Branch protection dokümanına göre status check'ler geçtikten sonra commit'ler ya başka bir branch'e push edilip merge edilmeli ya da doğrudan korumalı branch'e push edilmeli. | [About protected branches](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches#require-status-checks-before-merging) |
| Lock branch (branch protection) | Branch'i salt okunur yapar, silinemez. | [About protected branches](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches#lock-branch) |

**`v*` pattern'iyle ruleset tanımlanabilir mi?** Evet. fnmatch ile `v*` tüm `v` ile başlayan branch'lerle eşleşir. Exclude hedefi ile aktif versiyon hariç tutulabilir ([Creating rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/creating-rulesets-for-a-repository)). Örnek: include `v*`, exclude `v1.1`, kurallar Restrict updates + Restrict deletions + Block force pushes. Bu, kapanmış versiyonları kilitler. Ama her yeni versiyonda exclude listesini güncellemek gerekir (benim değerlendirmem). Not: `v*` hem branch hem tag adı olarak kullanılırsa `git` tarafında belirsiz ref uyarıları çıkabilir. Bu da versiyon için tag ya da branch'ten birini seçmenin bir başka nedeni. Doğrulanamadı: bu belirsizliğin GitHub tarafında ek bir soruna yol açıp açmadığı.

### 4.3 Actions tetikleyicileri

| Tetikleyici | GitHub'ın söylediği | Kaynak |
|---|---|---|
| `push` + `branches` | `branches` ya da `branches-ignore` filtresiyle workflow sadece belirli branch'lere push'ta koşar. `GITHUB_SHA` push edilen tip commit. | [Events that trigger workflows](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#push) |
| `release` | Activity type'lar: `published`, `unpublished`, `created`, `edited`, `deleted`, `prereleased`, `released`. Draft release'ler `created`/`edited`/`deleted` için tetiklemez. | [Events that trigger workflows](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#release) |
| `GITHUB_TOKEN` ile tetiklenen olaylar | `workflow_dispatch` ve `repository_dispatch` dışında, `GITHUB_TOKEN` ile tetiklenen olaylar yeni workflow run'ı oluşturmaz. | [Events that trigger workflows](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows) |
| Mevcut bir commit'i başka bir branch'e fast-forward push etmek | Dokümanda açıkça anlatılmıyor. | Doğrulanamadı |

**Kritik not (benim değerlendirmem):** Agent'lar GitHub Actions içinde `GITHUB_TOKEN` ile push ederse, o push'lar CI'ı tetiklemez. Agent push'larının CI'ı tetiklemesi isteniyorsa agent ayrı bir kimlikle (PAT ya da GitHub App token'ı) push etmeli.

### 4.4 Merge ve promote mekanizması

| Yöntem | Davranış | Kaynak |
|---|---|---|
| `git push origin v1.1:v1` (fast-forward) | `v1`'in yeni değeri eskisinin torunuysa fast-forward, tarih kaybolmaz. Değilse git push'u reddeder. Commit SHA'ları aynı kalır. | [git-push](https://git-scm.com/docs/git-push#_note_about_fast_forwards) |
| PR, "Create a merge commit" | Hedefte ek bir merge commit oluşur. | [About merge methods](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/incorporating-changes-from-a-pull-request/about-pull-request-merges) |
| PR, "Rebase and merge" | Fast-forward'a benzer ama her zaman committer bilgisini günceller ve yeni SHA'lar üretir. | [About merge methods](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/incorporating-changes-from-a-pull-request/about-pull-request-merges) |
| PR, "Squash and merge" | PR commit'leri tek commit'e sıkıştırılır. | [About merge methods](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/incorporating-changes-from-a-pull-request/about-pull-request-merges) |
| Tag + GitHub Release | Release'ler git tag'lerine dayanır. GitHub her release'e o tag'deki kaynak kodun zip ve tarball linklerini ekler. | [About releases](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases) |

Fowler, maturity branch'e geçişi "merge" değil "kopya" olarak düşünüyor, çünkü production kodunun upstream'de test edilenle birebir aynı olması isteniyor ([Fowler](https://martinfowler.com/articles/branching-patterns.html#maturity-branch)). Benim değerlendirmem: Bunu GitHub'da sağlayan yöntem fast-forward push ya da tag. PR merge yöntemlerinin hiçbiri test edilen SHA'yı hedefe aynen taşımıyor (squash da yeni commit üretiyor).

### 4.5 AFK agent'ın korunan yerlere push edememesi

| Seçenek | Değerlendirme | Kaynak |
|---|---|---|
| Agent'ı sahibin token'ıyla çalıştırmak | Repo sahibi admin olduğu için "Repository admin" bypass'ı agent'a da geçer. Ruleset agent'ı durdurmaz (benim çıkarımım). | Bypass rolleri: [Creating rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/creating-rulesets-for-a-repository) |
| Fine-grained PAT ile daraltmak | Token'a repo seçimi ve repo düzeyinde izin verilir. Token oluşturma adımlarında branch ya da tag düzeyinde bir kapsam seçeneği yok. Yani token tek başına "şu branch'e push edemez" diyemiyor. | [Managing personal access tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens) |
| Ayrı bir machine account | GitHub şartları, kişisel hesaba ek olarak bir ücretsiz machine account'a izin veriyor. Hesap yalnızca otomatik işler için kullanılmalı ve sorumluluk sahibinde. | [GitHub Terms of Service](https://docs.github.com/en/site-policy/github-terms/github-terms-of-service) |
| GitHub App | Bypass listesine eklenebilen ya da eklenmeyebilen ayrı bir aktör. | [Creating rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/creating-rulesets-for-a-repository) |

Benim değerlendirmem: Doğru yöntem **ayrı kimlik + ruleset**. Agent ayrı bir kimlikle (machine account collaborator ya da GitHub App) çalışır. Bypass listesinde sadece "Repository admin" (yani kullanıcı) olur. Böylece ruleset agent'ı durdurur, kullanıcıyı durdurmaz. Doğrulanamadı: kişisel hesaba ait bir repoda collaborator'ın tam olarak hangi rolü aldığı ve bu rolün bypass listesindeki "write" rolüyle eşleşip eşleşmediği. Kurulumdan sonra test edilmeli.

---

## 5. Bizim için ne anlama geliyor?

> Bu bölüm öneri. Hiçbiri uygulanmadı.

### 5.1 Önerilen branch modeli

**Tek trunk + doğrudan push + toplu kontrol + tag ile versiyon.**

1. Tek kalıcı branch: `main`. Default branch o.
2. Kullanıcı ve agent orkestratörü `main`'e doğrudan push eder. PR yok. Bu, trunk-based'in küçük ekipler için tarif ettiği stil ([trunkbaseddevelopment.com](https://trunkbaseddevelopment.com/)).
3. Paralel agent'lar Sandcastle'daki gibi kendi branch/worktree'lerinde çalışır. Entegrasyon tek noktadan, lokal olarak yapılır, sonra push edilir ([branch-strategies.md](https://github.com/mattpocock/sandcastle/blob/main/docs/guide/branch-strategies.md)). Bu kısa ömürlü branch'ler remote'a gitmek zorunda değil.
4. CI her push'ta koşar. İnsan review'u toplu, otomatik kontrol sürekli.
5. **Toplu kontrol döngüsü:** Kullanıcı yeni `ready-for-agent` etiketi vermeyi durdurur (ya da kontrol edilecek commit'i not eder). Biriken işi kontrol eder, bug'ları `main`'de toplu fix'ler (kendisi ya da agent'lar).
6. **Versiyonlama:** Stabil commit tag'lenir (`v1.1.0`) ve GitHub Release oluşturulur. "Büyük versiyon" (`v2`) ayrı bir branch değil, sadece `v2.0.0` tag'i.
7. **Opsiyonel:** Agent'lar stabilizasyon sırasında da çalışacaksa `main`'den `release/v1.1` kesilir. Fix'ler `main`'de yapılıp cherry-pick edilir, tag release branch'inden atılır, branch sonra silinir ([branch for release](https://trunkbaseddevelopment.com/branch-for-release/)).
8. **Eski versiyona fix gerekirse:** O tag'den o an bir branch açılır. Önceden versiyon branch'i tutulmaz.

Kullanıcı yine de versiyon branch'i modelinde kalmak isterse: default branch'i her yeni versiyonda aktif branch'e çevirmek, `v*` include + aktif versiyon exclude ruleset'iyle eski branch'leri kilitlemek ve workflow'larda `branches: ['v*']` kullanmak gerekir. Bu çalışır ama her versiyonda üç yerde elle güncelleme demek.

### 5.2 Somut GitHub konfigürasyon listesi

**Repo ayarları**

- Default branch: `main`.
- Merge yöntemleri: PR kullanılmayacağı için önemsiz. İstenirse sadece "Squash" açık bırakılır.
- Agent kimliği: Ayrı bir machine account (collaborator, write) ya da GitHub App. Agent'ın token'ı bu kimliğe ait olur, kullanıcının token'ı değil.

**Ruleset 1: `main-koruma` (branch)**

- Target: Default branch.
- Enforcement: Active.
- Kurallar: Restrict deletions, Block force pushes.
- PR zorunluluğu yok, status check zorunluluğu yok (doğrudan push niyetiyle çelişir).
- Bypass: Repository admin (Always allow). Kullanıcı gerekirse force push yapabilir.

**Ruleset 2: `versiyon-tag-koruma` (tag)**

- Target: Tag pattern `v*`.
- Enforcement: Active.
- Kurallar: Restrict creations, Restrict updates, Restrict deletions.
- Bypass: Repository admin (Always allow). Yani sadece kullanıcı versiyon tag'i atabilir. Agent versiyonlama yapamaz.

**Ruleset 3 (opsiyonel): `release-branch-koruma` (branch)**

- Target: `release/*`.
- Kurallar: Restrict creations, Restrict updates, Restrict deletions, Block force pushes.
- Bypass: Repository admin. Stabilizasyon branch'ine sadece kullanıcı dokunur.

**Workflow tetikleyicileri**

```yaml
# .github/workflows/ci.yml
on:
  push:
    branches: [main, 'release/**']
  pull_request:          # ileride PR kullanılırsa
  workflow_dispatch:
  schedule:
    - cron: '0 3 * * *'  # default branch (main) üzerinde koşar
```

```yaml
# .github/workflows/release.yml
on:
  push:
    tags: ['v*']         # tag push'unda build + GitHub Release
```

- Agent Actions içinde çalışacaksa `GITHUB_TOKEN` ile push etmemeli. Yoksa CI tetiklenmez ([Events that trigger workflows](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows)).

**Promote (versiyonlama) komutu**

```bash
# main stabil olduğunda, kullanıcı çalıştırır
git switch main && git pull --ff-only
git tag -a v1.1.0 -m "v1.1.0"
git push origin v1.1.0
gh release create v1.1.0 --verify-tag --generate-notes
```

Opsiyonel release branch akışı:

```bash
git switch -c release/v1.1 main && git push origin release/v1.1
# fix'ler main'de yapılır, sonra:
git switch release/v1.1 && git cherry-pick <fix-sha> && git push
git tag -a v1.1.0 -m "v1.1.0" && git push origin v1.1.0
gh release create v1.1.0 --verify-tag --generate-notes
```

Bu kararın bir ADR olarak `docs/adr/` altına yazılması düşünülebilir.

---

## Doğrulanamayanlar listesi

1. Kullanıcının (a) yolunun (`v1.1` → `v2`, geri merge yok, kalıcı mainline yok) birincil kaynaklarda yerleşik bir pattern adı. TBD.com'un "Mainline" ve "Cascade" modelleri ile Fowler'ın Release/Maturity Branch tanımları kısmen benziyor ama tam denk gelmiyor.
2. "Version branch", "maintenance branch", "branch per release" adlarının TBD.com ve Fowler'da isimli pattern olarak geçtiği. Bulamadım.
3. Sandcastle'ın remote'a kendisinin push edip etmediği ya da PR açıp açmadığı. Branch strategies dokümanı ve `simple-loop`/`parallel-planner` şablonlarında push adımı bulamadım.
4. Default branch'e doğrudan push edilen bir commit'teki `Closes #N`'in issue'yu kapattığı. Doküman sadece "merge" diyor.
5. Var olan bir commit'i başka bir branch'e fast-forward push etmenin (`git push origin v1.1:v1`) `push` workflow'unu tetiklediği. Dokümanda açık cümle yok.
6. Kişisel hesaba ait repoda collaborator rolünün ruleset bypass listesindeki "write" rolüyle nasıl eşleştiği.
7. Aynı adın (`v1.1`) hem branch hem tag olarak kullanılmasının GitHub tarafında ek sorun yaratıp yaratmadığı.
8. Ruleset'lerdeki "Evaluate" enforcement modunun branch/tag ruleset'lerinde Free planda kullanılabilirliği. Dokümanda sadece metadata kısıtları bağlamında geçiyor.
