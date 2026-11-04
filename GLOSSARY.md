# QA Hub

Case'lerin Library'de yazıldığı, run'lar içinde çalıştırıldığı ve status'larının kaydedildiği basit bir test yönetim uygulaması.

## Language

**Project**:
Run'ların ait olduğu, adı olan alan. Her run tam olarak bir project'e aittir. Case tutmaz.
_Avoid_: Proje, workspace

**Library**:
Bütün case'lerin durduğu, hiçbir project'e ait olmayan tek depo.
_Avoid_: Kütüphane, repository, bilgi bankası, cases

**Case**:
Library'de duran, tek bir amacı doğrulayan testin tanımı. Kendisi hiçbir zaman çalıştırılmaz ve status taşımaz.
_Avoid_: Test, test case, vaka, senaryo

**Precondition**:
Bir case'in çalışabilmesi için önce çalışıyor olması gereken başka bir case. Bilgi amaçlıdır, bir execution'a status girilmesini engellemez.
_Avoid_: Prerequisite, dependency, ön koşul

**Priority**:
Bir case'in önem derecesi: `Low`, `Medium` ya da `High`.
_Avoid_: Severity, önem

**Run**:
Adı olan bir çalıştırma oturumu. Başladığı anda Library'deki her case için bir execution içerir.
_Avoid_: Test run, cycle, koşu, koşum

**Execution**:
Bir case'in bir run içindeki tek çalıştırılması. Case'in run başladığı andaki halinin kopyasını taşır. Sonradan değişen, oluşturulan ya da silinen case'ler başlamış bir run'ı etkilemez.
_Avoid_: Test, run case, run item, kayıt

**Status**:
Bir execution'ın değeri: `Passed` ya da `Failed`. Yeni bir status öncekinin yerini alır. Hiç status girilmemiş bir execution "Untested"tır, bu bir status değildir.
_Avoid_: Result, sonuç, state, outcome
