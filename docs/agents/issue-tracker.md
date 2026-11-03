# Issue tracker: GitHub

Bu reponun issue'ları ve spec'leri GitHub issue'su olarak tutulur. Tüm işlemler için `gh` CLI'ı kullan.

## Kurallar

- **Issue oluştur**: `gh issue create --title "..." --body "..."`. Çok satırlı gövdeler için heredoc kullan.
- **Issue oku**: `gh issue view <number> --json number,title,body,labels,comments`.
- **Issue listele**: `gh issue list --state open --json number,title,body,labels,comments --jq '[.[] | {number, title, body, labels: [.labels[].name], comments: [.comments[].body]}]'`, uygun `--label` ve `--state` filtreleriyle.
- **Bir issue'yu bir parent'ın sub-issue'su yap**: `gh issue create --parent <parent> ...`, ya da sonradan `gh issue edit <parent> --add-sub-issue <child>` (`gh` 2.94+). Eski `gh`: `gh api --method POST repos/<owner>/<repo>/issues/<parent>/sub_issues -F sub_issue_id=<child-db-id>` (database id, aşağıdaki **Engelleme** maddesindeki gibi). Sub-issue yoksa child gövdesinin en üstüne `Part of #<parent>` yaz.
- **Issue'ya yorum yaz**: `gh issue comment <number> --body "..."`
- **Label ekle / kaldır**: `gh issue edit <number> --add-label "..."` / `--remove-label "..."`
- **Kapat**: `gh issue close <number> --comment "..."`

Repoyu `git remote -v`'den çıkar. Bir clone'un içinde çalışırken `gh` bunu otomatik yapar.

## Triage yüzeyi olarak pull request'ler (Pull requests as a triage surface)

**PRs as a request surface: no.** _(Bu repo dışarıdan gelen PR'ları feature isteği olarak ele alıyorsa `yes` yap. `/triage` bu ayarı okur.)_

`yes` olduğunda PR'lar da issue'larla aynı label ve durumlardan geçer, `gh pr` karşılıklarıyla:

- **PR oku**: `gh pr view <number> --comments`, diff için `gh pr diff <number>`.
- **Triage için dış PR'ları listele**: `gh api --paginate 'repos/{owner}/{repo}/pulls?state=open' --jq '.[] | select(.author_association | IN("OWNER","MEMBER","COLLABORATOR") | not) | {number, title, author: .user.login, author_association, labels: [.labels[].name]}'`.
- **Yorum / label / kapatma**: `gh pr comment`, `gh pr edit --add-label`/`--remove-label`, `gh pr close`.

GitHub issue'lar ve PR'lar için tek bir numara alanı kullanır, yani çıplak bir `#42` ikisinden biri olabilir. Önce `gh pr view 42` dene, olmazsa `gh issue view 42`.

## Bir skill "issue tracker'a yayınla" dediğinde (publish to the issue tracker)

Bir GitHub issue'su oluştur.

## Bir skill "ilgili ticket'ı getir" dediğinde (fetch the relevant ticket)

Yukarıdaki **Issue oku** maddesindeki gibi oku.

## Wayfinding işlemleri (Wayfinding operations)

`/wayfinder` tarafından kullanılır. **Map**, ticket'ları **child** issue olan tek bir issue'dur.

- **Map**: `wayfinder:map` label'lı, Notes / Decisions-so-far / Fog gövdesini tutan tek bir issue. `gh issue create --label wayfinder:map`.
- **Child ticket**: map'e GitHub sub-issue'su olarak bağlı bir issue (bkz. **Bir issue'yu bir parent'ın sub-issue'su yap**). Sub-issue yoksa child'ı map gövdesindeki bir task list'e ekle ve child gövdesinin en üstüne `Part of #<map>` yaz. Label'lar: `wayfinder:<type>` (`research`/`prototype`/`grilling`/`task`). Sahiplenildiğinde ticket, işi yürüten geliştiriciye atanır.
- **Engelleme (Blocking)**: GitHub'ın **yerleşik issue bağımlılıkları**, UI'da görünen asıl gösterim. Bir bağımlılık eklemek için: `gh api --method POST repos/<owner>/<repo>/issues/<child>/dependencies/blocked_by -F issue_id=<blocker-db-id>`. Burada `<blocker-db-id>`, engelleyen issue'nun sayısal **database id**'si (`gh api repos/<owner>/<repo>/issues/<n> --jq .id`), `#number` ya da `node_id` _değil_. GitHub `issue_dependencies_summary.blocked_by` alanını raporlar (sadece açık engelleyiciler, canlı kapı). Bağımlılıklar yoksa child gövdesinin en üstüne `Blocked by: #<n>, #<n>` satırı yaz. Tüm engelleyicileri kapandığında ticket'ın engeli kalkar.
- **Frontier sorgusu**: Map'in açık child'larını listele (`gh issue list --state open`, map'in sub-issue'ları / task list'iyle sınırlı). Açık engelleyicisi olanları (`issue_dependencies_summary.blocked_by > 0` ya da `Blocked by` satırında açık bir issue) ve atanmış olanları çıkar. Map sırasında ilk gelen kazanır.
- **Sahiplenme (Claim)**: `gh issue edit <n> --add-assignee @me`, oturumun ilk yazma işlemi.
- **Çözme (Resolve)**: `gh issue comment <n> --body "<cevap>"`, sonra `gh issue close <n>`, sonra map'in Decisions-so-far bölümüne bir context pointer (özet + link) ekle.
