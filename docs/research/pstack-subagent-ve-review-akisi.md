# pstack: parent issue + sub-issue için subagent ve review akışı

> Araştırma tarihi: 2026-11-03. Her iddianın yanında kaynağı var. "Doğrulanamadı" yazan yerler, birincil kaynağa ulaşılamayan ya da kaynakta açıkça yazmayan iddialar. "Benim değerlendirmem" yazan yerler çıkarım ya da yargı.

**Soru:** Parent issue (#1) ve 3 sub-issue (#2, #3, #4) için istenen topoloji: 1 ana session (Opus, orchestrator), sub-issue başına 1 code-writing subagent (Sonnet), her subagent'tan sonra ayrı bir Fable review döngüsü, en sonda bütün iş üzerinde son bir Fable review. Gerçek koşuda `/pstack:poteto-mode <issue url>` Feature playbook'una gitti. Tek bir Sonnet owner üç sub-issue'yu yazdı, Fable hiç spawn edilmedi, diff'i orchestrator kendisi inceledi.

**Kaynaklar (hepsi birincil):**

- Yerel kurulum, pstack v0.15.36: `/home/arda/.claude/plugins/cache/cursor-plugins/pstack/0.15.36/`. Aşağıda `skills/...` ve `docs/...` bu dizine göredir.
- Upstream: `github.com/cursor/plugins`, `pstack/` dizini, `gh api` ile okundu.
- Kullanıcı config'i: `/home/arda/.claude/rules/pstack-models.md`.
- Oturum transcript'i: `/home/arda/.claude/projects/-home-arda-repos-qa-hub/ed36b51c-4397-43ba-a3bb-0f13562958cb.jsonl`.

## TL;DR

- **Bu topolojiyi tam karşılayan bir pstack playbook'u yok.** Feature playbook'u tersini söylüyor: "Code-coupled work (one feature, one migration) goes to a single owner" (`skills/poteto-mode/playbooks/feature.md:19`). Gerçek koşudaki davranış playbook'a uygun. Orchestrator'ın gerekçesi de bu: "üç dilim aynı şemayı, router'ı ve şablonları paylaşıyor" (transcript, satır 79).
- **Konfigürasyonla topoloji ifade edilemez.** pstack'in tek config yüzeyi `~/.claude/rules/pstack-models.md` ve bu dosya yalnızca rol başına model (ve panel rollerinde liste uzunluğu) belirliyor. Kaç subagent olacağını, sub-issue başına bir tane olup olmayacağını, review adımı olup olmadığını ve final review'u belirleyemez. Fable'ı bir role atamak mümkün (`model` değerleri: `opus`, `sonnet`, `haiku`, `fable`, `docs/harness/claude-code.md:34`), ama Feature playbook'unda Fable'ın çalışabileceği bir reviewer rolü yok.
- **Fable'ın çalışabileceği roller:** `interrogate reviewers`, `swarm workers`, `arena cross-judge pool`. Bunlar yalnızca ilgili skill çalışırsa devreye girer. Feature playbook'u `interrogate`'i yalnızca "tasarım tartışmalıysa" çağırıyor (adım 7), `arena`'yı ise "birden çok geçerli şekil varsa" (adım 4). Orchestrator ikisini de atladı.
- **Benim değerlendirmem:** `arena cross-judge pool: fable xhigh` satırı Claude Code'da büyük ihtimalle hiç Fable çalıştırmaz. Arena judge'ı "parent'ın family'sinden asla olamaz" ve family'ler `claude-*` prefix'iyle ayrılıyor (`skills/arena/SKILL.md:19,30`). Opus parent'ı için Fable de Claude family'si. Test edilmedi.
- **En yakın yol:** Topolojiyi prompt'ta açıkça yazmak ve `figure-it-out`'a yönlendirmek. Poteto-mode "bundled playbook uymuyorsa" ve "geniş/çok parçalı iş" için bu yolu tarif ediyor (`skills/poteto-mode/SKILL.md:128`). Review'u da `interrogate`'e, modeli `interrogate reviewers: fable xhigh` satırına bağlamak. Tam öneri en sondaki "Öneri" bölümünde.
- **Upstream farkı yok:** Upstream main'in `pstack/plugin.json` sürümü de 0.15.36. İncelenen skill ve playbook dosyaları yerel kopyayla aynı, tek fark yerel kopyaya eklenen "Claude Code adapter" cümlesi (Bölüm 5).

---

## 1. pstack bu durum için hangi akışı öneriyor?

### 1.1 Routing: playbook'a nasıl gidiliyor?

| Konu | Kaynağın söylediği | Kaynak |
|---|---|---|
| Giriş noktası | `/poteto-mode` görevi bir playbook'a eşler ve playbook adımlarını todolist'e kopyalar. | `skills/poteto-mode/SKILL.md:126`, `docs/guide/02-poteto-mode.md:3` |
| Claude Code'da çağrı | Komutlar `pstack:` prefix'i alır: `/pstack:poteto-mode`. | `docs/guide/01-setup.md:115` |
| Başka skill çağırma | `Skill` tool'u çoğu pstack skill'ini reddeder. Adapter, `<pstack>/skills/<name>/SKILL.md` dosyasını `Read` edip izlemeyi söylüyor. Skill'lerde `disable-model-invocation: true` var (58 skill dosyasında). | `docs/harness/claude-code.md:12`, `grep -c disable-model skills/*/SKILL.md` |
| Geniş/çok parçalı iş | "A large or cross-cutting effort (a migration across many call sites, an ambitious multi-part change), or work the user steps away from to trust later, routes to the **figure-it-out** skill even when a narrower playbook like Feature fits." | `skills/poteto-mode/SKILL.md:128` |
| Uyan playbook yoksa | "Use **figure-it-out** whenever no bundled playbook fits. It designs a bespoke, rigorous playbook for the task." | `skills/poteto-mode/SKILL.md:128` |
| Skill sayma uyarısı | "don't enumerate skills in your prompt ... Name a skill only when you want to override a specific choice." | `docs/guide/02-poteto-mode.md:135` |

Bu koşuda argüman yalnızca issue URL'siydi (transcript, satır 6). Yeni davranış içeren iş olduğu için Feature'a gitmesi playbook tanımıyla tutarlı ("New or changed behavior, built from a named data shape", `skills/poteto-mode/SKILL.md:136`).

### 1.2 Playbook'ların karşılaştırması

"Review" sütunu: kim, hangi model rolü, birim başına mı sonda mı.

| Playbook | Kod yazan | Review/verification | Birim başına mı, sonda mı | Topolojiye uyum (benim değerlendirmem) |
|---|---|---|---|---|
| **Feature** (`playbooks/feature.md`) | Tek subagent, `feature, refactoring` rolü (satır 12). Kod-bağımlı iş tek owner'a gider (satır 19). | Orchestrator kendisi: "You own the design. Plan, review, verify" (satır 3). Ek olarak adım 7: tasarım tartışmalıysa `interrogate` (satır 16). | Birim başına review yok. Adım 6'da "small, ordered commits" ve `sequence-verifiable-units` var (satır 14-15), ama bu owner'ın kendi commit disiplini. | Düşük. Tek owner, ayrı reviewer rolü yok. Gerçek koşuda olan buydu. |
| **Multi-phase plan** (`playbooks/multi-phase-plan.md`) | Kimse. "The plan is the deliverable. Do not implement." (satır 3). | Plan şablonu her PR için swarm tabanlı verification yazar (satır 13, 96-118). | Plan, PR başına verify bloğu içerir. | Tek başına yetmez. Plan çıktısı, yürütme playbook'unu isimlendirir (satır 8). |
| **Autopilot-full** (`playbooks/autopilot-full.md`) | PR başına bir owner (satır 8). Owner'lar "off trunk, never stacked" paralel koşar (satır 13). | Swarm, kodu yazmayan verifier'lar. Tam swarm bir kez, code-ready head'de (satır 14). Merge yalnızca doğrulanmış koda (satır 17). | PR başına verdict. Bütün iş üzerinde ayrı final review yok. | Kısmi. Owner-per-unit ve yazar olmayan verifier var. Ama PR/merge/CI merkezli, owner'lar stack'lenmiyor (bağımlı commit'lerle çelişir), ve Cursor cloud agent varsayıyor (satır 8). |
| **Autopilot-stack** (`playbooks/autopilot-stack.md`) | Autopilot-full ile aynı. | Autopilot-full ile aynı, ama merge yok, "stop at verified" (satır 5). | PR başına. | Autopilot-full'dan biraz daha yakın: "Stack a PR only when it builds on another PR" (satır 6). Yine PR merkezli. |
| **Orchestrate** (`playbooks/orchestrate.md`) | "One owner per PR" (satır 9). Orchestrator "never author or edit code" (satır 3, 9). | "Each PR gets a verifier that isn't its author, on a different model family" (satır 13). | PR başına. Final review adımı tanımlı değil, yalnızca reply'da "what landed against the goal" (satır 25). | Topolojiyi en çok andıran playbook (orchestrator kod yazmaz, birim başına owner, birim başına yazar-olmayan verifier, "Start with one unit" satır 14). Ama kapsamı "multi-day, many stacked PRs, dozens to hundreds of subagents" (satır 3) ve "Work that one agent could finish inside the session's budget is not a program" diyor. 3 sub-issue bu tanımın altında. |
| **Autonomous run** (`playbooks/autonomous-run.md`) | Tek döngü. | Predicate (checkable exit condition), yazar olmayan reviewer yok. | Her iterasyonda predicate'e karşı doğrulama (satır 7-8). | Düşük. Tek göreve yönelik. |
| **figure-it-out** (`skills/figure-it-out/SKILL.md`) | Görev için tasarlanan workflow belirler. | "Pair delegated work with a judge" (satır 43). Final: "Check the whole against the Phase A predicate on the real product" (satır 52). | Phase C her birimi ayrı deney olarak doğrular (satır 40), Phase E bütünü doğrular. | Yapısal olarak en yakın: birim-birim doğrulama + sonda bütünü doğrulama + "judge" ile eşleme. Topolojiyi sabit bir şablon değil, görev başına tasarlıyor. Fable ataması ise yine config'e/prompt'a kalıyor. |
| **Shipping** (`playbooks/shipping.md`) | Kimse (landing). | Her PR için yazar olmayan verifier, `PASS`/`PASS+NOTES`/`FAIL` (satır 5). | PR başına, bottom-up. | Review mantığı doğru ama PR landing'e özel. Bu repoda PR yok. |
| **Babysit** (`playbooks/babysit.md`) | Kimse. | CI/review thread yönetimi. | — | Alakasız (PR durumu takibi). |
| **Opening a PR** (`playbooks/opening-a-pr.md`) | — | Her playbook'un sonunda çağrılır (satır 3). "Prefer five narrow PRs to one large PR" ve stack tarifi (satır 28). | — | Bu repoda PR yok: `CLAUDE.md`, PR yerine doğrulanmış işi doğrudan `main`'e push etmeyi söylüyor. |
| **Refactoring**, **Bug fix**, **Perf issue** | Feature gibi tek owner. | Feature gibi orchestrator. | — | Alakasız. |

### 1.3 İlgili yardımcı skill'ler

| Skill | Ne yapıyor | Review açısından | Kaynak |
|---|---|---|---|
| **interrogate** | Adversary'ler (panel) değişikliği kırmaya çalışır. Yalnızca bir çalıştırmayla kanıtlanan bulgu kalır. Fix uygulamaz. Kapsam: kullanıcının adlandırdığı şey, yoksa branch diff'i. | Gerçek "reviewer" benzeri tek skill. Birim başına ya da sonda, kapsamı sen adlandırarak. Rol: `interrogate reviewers`. | `skills/interrogate/SKILL.md:11-33` |
| **swarm** | N paralel worker, tek rapor (`PASS`/`ISSUES`/`BLOCKED`). | Verification lane'leri. Worker modeli `swarm workers` satırı. Verification her zaman bu model ve hızda kalır. | `skills/swarm/SKILL.md:24,33`, `skills/poteto-mode/SKILL.md:100` |
| **arena** | N aday aynı işi yapar, base seçilir, graft edilir. Phase C'de tek cross-judge. | Tasarım/kod bakeoff'u. Review değil. | `skills/arena/SKILL.md:19,28-30` |
| **architect** | Arena ile paralel tasarım sketch'i, sonra implement. | Tasarım aşaması. `architect runners` rolü. | `skills/architect/SKILL.md:31-33` |
| **reflect** | Transcript'ten skill iyileştirme. | Kod review'u değil. | `skills/reflect/SKILL.md:29-41` |
| **sequence-verifiable-units** (principle) | Küçük, her biri doğrulanabilir birimler. "verify each before the next". Commit/PR'ları kanıtlayan sırada diz. | Disiplin, spawn etmez. Review'u kimin yaptığını belirlemez. | `skills/principle-sequence-verifiable-units/SKILL.md:9-15` |
| **separate-before-serializing-shared-state** (principle) | Paralel aktörler aynı dosya/branch'e yazıyorsa önce paylaşımı kaldır. | Feature step 3'ün "Independent workstreams / Shared mutable state" kararını besliyor (`feature.md:9-10`). Orchestrator burada "n/a: aynı şemaya yazıyor" dedi (transcript, satır 86). | `skills/poteto-mode/SKILL.md:65` |

### 1.4 Transcript'te ne oldu?

| Gözlem | Kaynak |
|---|---|
| Çağrı: `/pstack:poteto-mode https://github.com/ardaarslan47/qa-hub/issues/1`. | transcript satır 6 |
| Orchestrator: "Uygulamayı tek bir owner'a veriyorum. Çünkü üç dilim aynı şemayı, router'ı ve şablonları paylaşıyor." | transcript satır 79 |
| Spawn: `subagent_type: pstack:poteto-agent`, `model: sonnet`, `effort: xhigh`, `isolation: worktree`. Brief: "Do not spawn subagents; own the diff directly." | transcript satır 80 |
| Todolist: `how` ve `architect` skip (greenfield), throughput checkpoint "Tek owner", `interrogate` **skip: tartışmalı bir tasarım yok**. | transcript satır 86 |
| Review: owner'ın raporuna ek olarak orchestrator `make check`'i son commit'te kendisi koştu. Fable spawn edilmedi. | transcript satır 153 |

Bu, playbook'un yazılı davranışıyla uyumlu: Feature adım 3'ün "Smallest safe decomposition. If one worker is best, name why." maddesi (`feature.md:11`) tam olarak buna izin veriyor, ve adım 7 yalnızca "design is contested" ise `interrogate` diyor.

---

## 2. Bu pstack konfigürasyonuyla yapılabilir mi?

### 2.1 Mevcut config yüzeyleri

| Yüzey | Ne var | Kaynak |
|---|---|---|
| `setup-pstack` skill'i | Tek yazılan dosya: Claude Code'da `~/.claude/rules/pstack-models.md`, frontmatter yok, yorum satırları + `# budget` satırı + rol başına bir satır. | `skills/setup-pstack/SKILL.md:41`, `docs/harness/claude-code.md:46` |
| `pstack-models` rule dosyası | Claude Code her yeni session'a `~/.claude/rules/` altındaki bütün `.md` dosyalarını yükler. Rol satırları zaten context'te. | `docs/harness/claude-code.md:37` |
| `# budget` satırı | Etiketler: `unlimited` (max), `large` (xhigh), `medium` (high), `small` (medium). `/setup-pstack` yeniden çalışınca her gerçek slug'ın effort token'ını buna göre ayarlar. | `skills/setup-pstack/SKILL.md:26-31` |
| Rol değer biçimi (Claude Code) | `<model> <effort>`, ör. `opus high`. `model` yalnızca `opus`, `sonnet`, `haiku`, `fable`. Cursor'daki `grok-*`/`gpt-*` slug'ları `model` alanı atlanarak parent modelinde çalışır. | `docs/harness/claude-code.md:34-35,38` |
| Alias değerler | `inherit-parent` ve `auto`: Task `model` alanı atlanır, rol parent chat modelinde çalışır. | `skills/setup-pstack/SKILL.md:33,49`, `docs/harness/claude-code.md:36` |
| Panel listeleri | `arena runners`, `architect runners`, `interrogate reviewers` liste değeri alır. Her giriş için bir subagent çalışır, alias girişler de sayılır, yani liste uzunluğu panel boyutunu belirler. | `skills/setup-pstack/SKILL.md:33` |
| `arena cross-judge pool` | Liste ama arena tek judge çalıştırır. Yalnızca parent'ın family'sinden **farklı** family'deki girişleri sırayla dener. `inherit-parent`/`auto` hiç judge olmaz. | `skills/setup-pstack/SKILL.md:33`, `skills/arena/SKILL.md:30` |
| `swarm workers` | Her swarm worker'ının varsayılan modeli (bir race/comparison arm başına model atamadıkça). | `skills/setup-pstack/SKILL.md:33`, `skills/swarm/SKILL.md:24` |
| Rule olmayan rol | Skill varsayılanı kullanılır. Satır silmek varsayılana döndürür. | `skills/setup-pstack/SKILL.md:48`, `docs/guide/01-setup.md:129` |
| Per-session/per-subagent ayarı | Poteto-mode her spawn'da effort'u bir adım yukarı/aşağı oynatabilir ve fast modu seçebilir. Family'yi değiştirmez. Verification `swarm workers` modelinde kalır. | `skills/poteto-mode/SKILL.md:100` |

**Araştırılıp bulunamayanlar** (kurulu pakette `grep` ile `env var`, `PSTACK_`, `settings.json`, `.pstack`, `pstack.json`, `config file` arandı; yalnızca alakasız eşleşmeler çıktı):

- Environment variable: yok.
- CLI flag: yok. `plugin.json` yalnızca ad/sürüm/meta içeriyor (`plugin.json`, 1-22. satırlar).
- Plugin settings: yok.
- Per-repo override: pstack kaynaklarında tanımlı değil. Adapter yalnızca `~/.claude/rules/pstack-models.md` yolunu anıyor (`docs/harness/claude-code.md:37,46`). Claude Code'un proje düzeyinde `.claude/rules/` yüklemesi bir Claude Code özelliği olabilir, ama pstack'in o yolu okuyup rol satırı olarak yorumladığı birincil kaynakta yazmıyor. Doğrulanamadı.
- "Orchestrator modeli": pstack ayarlamıyor. Ana session modeli Claude Code'un session modeli. `inherit-parent` bu modeli ifade ediyor (`docs/harness/claude-code.md:36`). Orchestrator'ı Opus yapmak pstack dışı, session ayarı.

### 2.2 `setup-pstack`'in tüm rolleri

Rol satırları, `skills/setup-pstack/SKILL.md:51-67` (kullanıcının dosyasında da aynı 17 satır var, `/home/arda/.claude/rules/pstack-models.md`):

| Rol satırı | Hangi skill/playbook okuyor | Panel mi? | Kullanıcı değeri |
|---|---|---|---|
| `feature, refactoring` | Feature/Refactoring playbook'u (kod delegesi) | Hayır | `sonnet xhigh` |
| `bug-fix` | Bug fix | Hayır | `sonnet xhigh` |
| `perf-issue` | Perf issue | Hayır | `sonnet xhigh` |
| `hillclimb` | Hillclimb | Hayır | `sonnet xhigh` |
| `judgment and prose` | Düzyazı ve yargı işleri | Hayır | `opus xhigh` |
| `hardest tasks` | En zor kod değişiklikleri (`poteto-mode/SKILL.md:98`) | Hayır | `opus xhigh` |
| `how explorer` | `how` explorer'lar | Hayır | `sonnet xhigh` |
| `how explainer` | `how` explainer | Hayır | `opus xhigh` |
| `why investigators` | `why` | Hayır | `sonnet xhigh` |
| `why synthesizer` | `why` | Hayır | `opus xhigh` |
| `reflect tooling` | `reflect` tooling lens | Hayır | `sonnet xhigh` |
| `reflect judgment, divergent, synthesizer` | `reflect` | Hayır | `opus xhigh` |
| `arena runners` | `arena` aday sayısı ve modelleri | **Evet** | `opus xhigh, sonnet xhigh` |
| `arena cross-judge pool` | `arena` Phase C judge | Pool (tek judge) | `fable xhigh` |
| `swarm workers` | `swarm` worker'ları, verification lane'leri | Hayır (tek varsayılan) | `sonnet xhigh` |
| `architect runners` | `architect` (arena üzerinden) | **Evet** | `opus xhigh, sonnet xhigh` |
| `interrogate reviewers` | `interrogate` adversary'leri | **Evet** | `opus xhigh, sonnet xhigh` |

Bu listede "feature reviewer", "per-unit reviewer" ya da "final reviewer" diye bir rol **yok**. Reviewer işi görebilecek roller yalnızca `interrogate reviewers`, `swarm workers` (verification) ve `arena cross-judge pool` (arena içinde).

### 2.3 Soru-soru: config ne kontrol edebilir?

| Soru | Kontrol edilebilir mi? | Neden / nasıl | Kaynak |
|---|---|---|---|
| (a) Subagent sayısı/topolojisi | **Kısmen, yalnızca panel rollerinde.** Liste uzunluğu = panel boyutu. Kod yazan subagent sayısı config'den gelmiyor. Onu Feature step 3'ün throughput checkpoint'i (orchestrator'ın o anki kararı) belirliyor. | `skills/setup-pstack/SKILL.md:33`, `feature.md:7-11` |
| (b) Sub-issue başına bir subagent | **Hayır.** Config'de "sub-issue", "unit" ya da "owner" kavramı yok. Bu karar playbook'un çıktısı. Playbook "single owner" ya da "one worker is best, name why" diyebilir. Zorlayan tek şey prompt ya da playbook seçimi (Orchestrate/Autopilot'ta "One owner per PR"). | `feature.md:11,19`, `orchestrate.md:9`, `autopilot-full.md:8` |
| (c) Her unit sonrası review adımı | **Hayır.** Review adımının varlığı playbook'ta. Feature'da yalnızca koşullu `interrogate` (adım 7, sonda) var. Autopilot/Orchestrate'te per-PR verifier var ama bunlar PR/merge odaklı. | `feature.md:16`, `orchestrate.md:13`, `autopilot-full.md:14` |
| (d) Review'u hangi model yapar | **Evet, rol varsa.** `fable` geçerli bir `model` değeri. Atanabilecek roller: `interrogate reviewers`, `swarm workers`, `arena cross-judge pool` ve diğer rollerin hepsi (ama bunlar reviewer değil). Alias ile (`inherit-parent`) ana session modeline bağlamak da mümkün. | `docs/harness/claude-code.md:34`, `skills/setup-pstack/SKILL.md:51-67` |
| (e) Final review | **Hayır.** Playbook adımı. `interrogate`'i sona koyan Feature adım 7 ya da kullanıcı prompt'u. Config yalnızca çalışırsa kimin yapacağını belirler. | `feature.md:16`, `skills/interrogate/SKILL.md:15` |

### 2.4 Rol bazında etki alanı (yanlış anlamaya açık noktalar)

| Satır | Gerçekte neyi etkiler | Neyi etkilemez |
|---|---|---|
| `arena cross-judge pool: fable xhigh` | Yalnızca `arena` Phase C'deki tek, readonly judge. | Feature'ın kod delegesini, orchestrator'ın review'unu ya da `interrogate`'i. Arena çalışmazsa hiç devreye girmez. Feature adım 4 arena'yı yalnızca "birden çok geçerli şekil" varsa öneriyor (`feature.md:12`), orchestrator atladı. |
| `swarm workers: <model>` | `swarm` worker'ları ve poteto-mode'daki tüm "verification" işleri: "Verification is the exception and stays on the `swarm workers` model" (`SKILL.md:100`). Autopilot-full/stack'in verifier'ları ve multi-phase plan'ın lane'leri bu modelle koşar. | Kod yazan owner'ı, `interrogate` adversary'lerini. |
| `interrogate reviewers: <liste>` | `interrogate` panelinin boyutu ve modelleri. | `interrogate`'in ne zaman çağrılacağını. Çağrı karar noktası Feature adım 7 ya da kullanıcı. |
| `feature, refactoring: <model>` | Feature/Refactoring'deki **kod delegesinin** modeli. | Reviewer'ı. |
| `inherit-parent`/`auto` | Task `model` alanını atlar, rol ana session modelini kullanır. | Hiçbir topoloji kararı. |

### 2.5 Yalnızca model config'iyle neler yapılabilir, neler yapılamaz?

**Yapılabilir:**

- Kod yazan delegeyi Sonnet'te tutmak (`feature, refactoring: sonnet xhigh`, zaten var).
- `interrogate` çalışırsa adversary'leri Fable yapmak (`interrogate reviewers: fable xhigh`). Tek giriş tek reviewer demek: arena Phase A adım 3 "or one when the line has one entry" (`skills/arena/SKILL.md:19`), `interrogate` o adımı kullanıyor (`skills/interrogate/SKILL.md:19`).
- Verification işlerini Fable'a vermek (`swarm workers: fable xhigh`). Yan etkisi: `/swarm` ve Autopilot/Orchestrate verifier'larının hepsi Fable'a gider.

**Yapılamaz:**

- Her sub-issue için ayrı owner'ı zorlamak.
- Her owner'dan sonra review adımını zorlamak.
- Sonda bütün iş için review zorlamak.
- Orchestrator'ın kod yazmamasını zorlamak (Orchestrate dışında playbook bunu söylemiyor, Feature'da "Delegate implementation. Stay in the lead" var, `feature.md:3`).

**Benim değerlendirmem, `fable` ve "family" kuralı:** `arena cross-judge pool` ve Orchestrate'in "on a different model family" ifadesi (`orchestrate.md:13`) family'yi model ailesi olarak tanımlıyor. Arena'da family'ler prefix'e göre ayrılıyor: `claude-*`, `grok-*` (`skills/arena/SKILL.md:19`). Claude Code adapter'ı `claude-<family>-...` slug'larını `opus`/`sonnet`/`haiku`/`fable` alanlarına çeviriyor (`docs/harness/claude-code.md:34`), yani dördü de aynı "claude" family'si. Opus parent'ı ile `arena cross-judge pool: fable xhigh` büyük ihtimalle "parent'ın family'si" sayılıp elenir ve arena "skip the judge" yoluna düşer (`skills/arena/SKILL.md:30`). Adapter bu durumu açıkça ele almıyor, test de edilmedi. Doğrulanamadı: Fable'ın Opus parent'ına göre ayrı family sayılıp sayılmadığı.

### 2.6 En yakın "yalnızca config" yaklaşımı

Yalnızca config ile yaklaşılabilecek şey, "Fable'ı reviewer yapmak ve doğru playbook'u/skill'i çağırmak":

```
interrogate reviewers: fable xhigh
swarm workers: fable xhigh
```

Bunlar yalnızca şu çağrılarda etkili olur:

| Config | Hangi çağrı onu çalıştırır |
|---|---|
| `interrogate reviewers: fable xhigh` | Kullanıcının "interrogate" ya da "adversarial review" demesi (`skills/interrogate/SKILL.md:3`) ya da Feature adım 7 (tasarım tartışmalıysa). |
| `swarm workers: fable xhigh` | `/swarm`, Autopilot-full/stack, Orchestrate (verifier), Multi-phase plan'ın lane'leri. |

Hiçbiri birim başına/ sonda ayrımını ya da owner sayısını belirlemez. Bunlar için prompt gerekiyor (Bölüm 3).

---

## 3. Config ifade edemiyorsa pstack iş akışı talimatı için ne diyor?

pstack, proje düzeyinde "workflow talimatı"nın nereye yazılacağını tek bir yerde toplu olarak söylemiyor. Dağınık ipuçları:

| Konu | Kaynağın söylediği | Kaynak |
|---|---|---|
| Always-applied kurallar | Cursor'daki `.cursor/rules/*.mdc` always-applied kuralları Claude Code'da `CLAUDE.md`'ye (yoksa `AGENTS.md`'ye) eşlenir. Model kuralı bunun istisnası. | `docs/harness/claude-code.md:54` |
| Model satırları CLAUDE.md'ye yazılmaz | "Never put the role lines in `CLAUDE.md` or `AGENTS.md`." | `docs/harness/claude-code.md:46` |
| Tekrarlanan hata için düzeltme sırası | 1) mimari/veri yapısı, 2) type/lint/CI kontrolü, 3) test, 4) "doc or agent rule. Nothing fails when an agent skips a rule, so this comes last." | `docs/guide/09-make-it-yours.md:35-40` |
| Küçük başla | "add a skill or a check when the same failure shows up twice." | `docs/guide/09-make-it-yours.md:5` |
| Tekrar eden workflow | "A workflow you keep hitting but isn't captured → propose a new skill." | `skills/poteto-mode/playbooks/authoring-a-skill.md:22` |
| Kişisel stil skill'i | `/automate-me`, transcript'lerden tercihleri çıkarıp `<isim>-mode` skill'i yazar. Claude Code'da skill yolu `.claude/skills/` ya da `~/.claude/skills/`. | `docs/guide/09-make-it-yours.md:13`, `docs/harness/claude-code.md:52` |
| Tek seferlik özel akış | `figure-it-out`: görev uymuyorsa workflow'u o görev için tasarlar, kararları `show-me-your-work` ile loglar. | `skills/poteto-mode/SKILL.md:128`, `skills/figure-it-out/SKILL.md:11,34` |
| Plan dosyasında topoloji | `multi-phase-plan` şablonu, "Spawn owners" bölümünde PR başına owner, bağımlılık grafiği, dosya sınırları ve review gate'i yazdırır; yürütme playbook'unu isimlendirir. | `skills/poteto-mode/playbooks/multi-phase-plan.md:45-52,8` |
| Prompt | "Name a skill only when you want to override a specific choice." Playbook sıralamasını bozmamak için skill sayılmaz. | `docs/guide/02-poteto-mode.md:135` |

Doğrulanamadı: pstack'in "bir repo için kalıcı bir playbook seçimi/topoloji tercihi" bırakılacak resmi bir yer tanımladığı. Bulunan en yakın mekanizmalar: kalıcı skill (`authoring-a-skill`), tek seferlik prompt (`figure-it-out`) ve plan dosyası (`multi-phase-plan`).

Benim değerlendirmem: `CLAUDE.md`, pstack'in kendi sıralamasında en zayıf seçenek (`09-make-it-yours.md:40`) ama topoloji gibi "bir tool'a bağlanamayan" tercih için tek kalıcı ve ucuz yer. Adapter da always-applied talimatın Claude Code karşılığının `CLAUDE.md` olduğunu söylüyor (`claude-code.md:54`). Bu repoda `CLAUDE.md` zaten pstack'e bir istisna ekliyor: PR yok, doğrudan `main`'e push (`/home/arda/repos/qa-hub/CLAUDE.md`).

---

## 4. Karar tablosu: hangi yol neyi sağlıyor?

| Yol | Owner/unit | Fable per-unit review | Fable final review | Kalıcı mı | Maliyet |
|---|---|---|---|---|---|
| Sadece `pstack-models.md` (Fable rollere atanmış) | Hayır (Feature tek owner) | Hayır (interrogate çağrılmazsa rol çalışmaz) | Hayır | Evet | Düşük |
| Config + prompt'ta topoloji ve `figure-it-out` | Evet (prompt) | Evet (prompt, `interrogate`) | Evet (prompt, `interrogate`) | Hayır (her koşuda yaz) | Düşük |
| Config + `autopilot-stack` | Evet (PR başına owner, ama "off trunk" ve PR varsayımı) | Kısmen (swarm lane'leri, `swarm workers: fable`) | Hayır (tanımlı değil) | Hayır | Yüksek (ağır swarm, 10 live lane, `cursor-team-kit` control skill'leri) |
| Config + `orchestrate` | Evet | Evet (verifier, "different model family") | Hayır (tanımlı değil) | Hayır | Yüksek, 3 birim için tanım dışı |
| Config + `CLAUDE.md` bölümü | Evet (talimat) | Evet (talimat) | Evet (talimat) | Evet | Düşük ama her session'a yüklenir |
| Özel `-mode` skill'i veya proje skill'i | Evet | Evet | Evet | Evet | Yazma maliyeti, `authoring-a-skill` yolu |

Autopilot-full/stack'in `control-ui`/`control-cli` skill'leri `cursor-team-kit` plugin'inden geliyor (`autopilot-full.md:14`, `multi-phase-plan.md:15`). Bu plugin'in bu makinede kurulu olup olmadığı kontrol edilmedi. Doğrulanamadı.

---

## 5. Upstream karşılaştırması

| Kontrol | Sonuç | Kaynak |
|---|---|---|
| Upstream `pstack/plugin.json` sürümü | 0.15.36, yerel kopyayla aynı. | `gh api repos/cursor/plugins/contents/pstack/plugin.json` |
| Upstream'in son pstack commit'i | `c742dd61`, 2026-10-27, "make Orchestrate agentic and delete the orch scripts (#1043)". Yerel `orchestrate.md` upstream ile aynı. | `gh api repos/cursor/plugins/commits?path=pstack` |
| `setup-pstack/SKILL.md` | Yerel kopyada yalnızca fazladan "Claude Code adapter" cümlesi var. Rol listesi aynı. | `diff` upstream/yerel |
| `feature.md`, `multi-phase-plan.md`, `orchestrate.md`, `autopilot-full.md` | Aynı. | `diff` |
| `poteto-mode/SKILL.md`, `swarm`, `arena`, `interrogate`, `architect`, `figure-it-out` | Yalnızca yerel kopyaya eklenen "Claude Code adapter" cümlesi farklı. | `diff` |
| `docs/harness/claude-code.md` | Aynı. | `diff` |
| Skill dizin listesi | Aynı. | `gh api repos/cursor/plugins/contents/pstack/skills` |

Sonuç: Upstream main'de yeni bir rol ya da yeni bir config yüzeyi yok. Ancak yalnızca bu dosyalar karşılaştırıldı, `references/` altındakiler ve `docs/guide/` karşılaştırılmadı.

---

## Doğrulanamayanlar listesi

1. Per-repo (proje düzeyi) bir pstack model override'ının var olup olmadığı. Pstack kaynakları yalnızca `~/.claude/rules/pstack-models.md` yolunu anıyor.
2. `fable`'ın, Opus parent'ına göre "farklı family" sayılıp sayılmadığı (arena cross-judge ve Orchestrate'in "different model family" ifadesi). Benim çıkarımım: sayılmaz.
3. `fable xhigh` değerinin, `Agent` tool şemasında Fable için geçerli bir effort olup olmadığı. Kullanıcının dosyası bunu kullanıyor, adapter şemanın `effort` değerlerine bakmayı söylüyor (`docs/harness/claude-code.md:44`). Bu oturumda şema doğrulanmadı.
4. Autopilot ve multi-phase plan'ın referans verdiği `cursor-team-kit` control skill'lerinin bu makinede kurulu olması.
5. pstack'in proje başına kalıcı bir playbook/topoloji tercihi için resmi bir yer tanımladığı.
6. Bir `poteto-agent` owner'ın Claude Code'da iç içe subagent spawn edip edemeyeceği. Transcript'teki brief "Do not spawn subagents" diyor (satır 80); bu, Feature adım 4'ün "A subagent forbidden to spawn" yolunu izliyor (`feature.md:12`). Sınır Claude Code'dan mı yoksa brief'ten mi, doğrulanmadı.

---

## Öneri

> Bu bölüm benim değerlendirmem. Hiçbiri uygulanmadı.

**Ana fikir:** Config yalnızca modeli belirleyebilir, topolojiyi yazan prompt (ya da kalıcı bir talimat) olmalı. İki parçalı kur: config'e Fable'ı reviewer role'üne bağla, topolojiyi orchestrator'a söyle.

### 1. Config satırı (`~/.claude/rules/pstack-models.md`)

```
interrogate reviewers: fable xhigh
```

- Tek giriş = tek Fable reviewer (`skills/arena/SKILL.md:19`). İki bağımsız bakış istenirse `fable xhigh, fable xhigh`.
- `swarm workers`'ü Fable yapmak isteğe bağlı. Yan etkisi tüm swarm verification'ının Fable'a gitmesi. Şu an gerekmiyor çünkü review `interrogate` ile yapılacak. Değiştirme.
- `arena cross-judge pool: fable xhigh` satırı zararsız ama muhtemelen etkisiz (Bölüm 2.5). Bırakılabilir.
- `feature, refactoring: sonnet xhigh` zaten doğru.
- Orchestrator'ın Opus olması bu dosyayla değil, Claude Code session modeliyle ayarlanır.

### 2. Çağrı (kalıcı bir şey yazmadan)

`figure-it-out`'a yönlendiren ve topolojiyi sabitleyen bir prompt. Poteto-mode "uyan playbook yoksa figure-it-out" diyor (`SKILL.md:128`), ve "Name a skill only when you want to override a specific choice" kuralı (`02-poteto-mode.md:135`) burada geçerli, çünkü Feature'ın varsayılan seçimini bilerek ezmek istiyoruz:

```
/pstack:poteto-mode https://github.com/ardaarslan47/qa-hub/issues/1

Bu iş figure-it-out ile yürüsün, Feature playbook'unun tek owner kararı burada geçerli değil. Topoloji:
- Sen orchestrator'sın, kod yazma.
- Her sub-issue için ayrı, taze bir Sonnet owner (pstack:poteto-agent). Sub-issue'lar aynı şemaya dokunuyor, o yüzden paralel değil, sırayla: her owner bir önceki commit'in üstünde çalışsın.
- Bir owner bitince, o commit'i kapsam adlandırarak interrogate ile incele (reviewer Fable, kod yazmasın). Doğrulanmış bulgu varsa onu yeni bir Sonnet'e ver, düzeltsin. Temiz çıkmadan sonraki sub-issue'ya geçme.
- Üç sub-issue bitince interrogate'i #1'in bütün commit aralığı üzerinde bir kez daha koştur.
- Her commit bir sub-issue'yu kapatsın, PR yok, doğrudan main.
```

Bu metnin dayandığı pstack mekanizmaları: per-unit doğrulama (`figure-it-out` Phase C, `SKILL.md:40`), judge eşleme (satır 43), sonda bütünü doğrulama (Phase E, satır 52), taze subagent kuralı (`poteto-mode/SKILL.md:104`), `interrogate` kapsamı kullanıcı adlandırır (`interrogate/SKILL.md:15`).

### 3. Eğer her seferinde yazmak istenmiyorsa: minimum `CLAUDE.md` metni

Yalnızca bu tip işler için tekrarlanan bir tercihse. Pstack'in kendi sıralaması bunu en zayıf seviye sayıyor (`09-make-it-yours.md:40`), o yüzden yalnızca ikinci kez düzeltmek zorunda kalırsan ekle:

```
### Parent issue ve sub-issue'lar için pstack akışı

Sub-issue'lı bir parent issue verildiğinde: orchestrator kod yazmaz. Her sub-issue için sırayla ayrı, taze bir Sonnet owner çalışır. Her owner'dan sonra o commit üzerinde interrogate (Fable reviewer) koşar, doğrulanmış bulgular yeni bir owner'a düzelttirilir. Hepsi bitince interrogate bütün commit aralığında bir kez daha koşar.
```

Model satırları ise `CLAUDE.md`'ye girmez (`claude-code.md:46`), sadece bu topoloji cümlesi girer. Daha kalıcı ve daha az dağınık seçenek, bu akışı bir proje skill'ine (`.claude/skills/`) çevirmek, ama bu, `authoring-a-skill` playbook'una göre yazılmalı (`playbooks/authoring-a-skill.md:5-20`).

### 4. Önermediğim yollar

- **Autopilot-stack/Autopilot-full:** PR/CI/merge merkezli, owner'lar paralel ve "never stacked" (full), her PR için ağır bir swarm (10 live lane). Bu repoda PR yok ve üç küçük birim için orantısız.
- **Orchestrate:** Kapsamı "multi-day, dozens of subagents". Playbook kendisi 3 birimlik işi program saymıyor (`orchestrate.md:3`).
- **Sadece config:** Topoloji ifade edemez (Bölüm 2.3), Fable reviewer'ı hiç çağrılmaz.
