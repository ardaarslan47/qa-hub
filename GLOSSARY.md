# QA Hub

Test case'lerin yazıldığı, run'lar içinde çalıştırıldığı ve status'larının kaydedildiği basit bir test yönetim uygulaması.

## Language

**Case**:
Neyin, nasıl test edileceğinin tanımı. Kendisi hiçbir zaman çalıştırılmaz ve status taşımaz.
_Avoid_: Test, test case, vaka, senaryo

**Run**:
Adı olan bir çalıştırma oturumu. Başladığı anda var olan her case için bir execution içerir.
_Avoid_: Test run, cycle, koşu, koşum

**Execution**:
Bir case'in bir run içindeki tek çalıştırılması. Case'in run başladığı andaki halinin kopyasını taşır. Sonradan değişen ya da oluşturulan case'ler başlamış bir run'ı etkilemez.
_Avoid_: Test, run case, run item, kayıt

**Status**:
Bir execution'ın değeri: `Passed` ya da `Failed`. Yeni bir status öncekinin yerini alır. Hiç status girilmemiş bir execution "Untested"tır, bu bir status değildir.
_Avoid_: Result, sonuç, state, outcome
