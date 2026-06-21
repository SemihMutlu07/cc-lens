# CHECKUP.md — cc_lens recurring error patterns

Loop-agent: her iterasyonda bu 7 check'i çalıştır. Sıralama önem sırasına göre.

## 1. Parser gaps — "detected ama 0 kayıt"

**Pattern:** Kaynak dosyada bulunuyor (`state="detected"`), ama parser 0 kayıt döndürüyor. Hermes + Cursor en sık fail edenler.

**Check:**
```
curl -s localhost:8080/api/wrapped | jq '.sources[] | select(.state=="detected" and .count==0)'
```
Boş değilse → yeni log formatı parser'ı kırmış. `parser.go`'ya yeni format eklenmeli.

**Real incident:** Hermes diskteki dosyaları buluyor ama 0 kayıt — "hermes okunmuyor" şikayetinin kökü bu.

---

## 2. Türetilmiş metrik — yanıltıcı sayı

**Pattern:** UI'da gösterilen sayı gerçek bir ölçüm değil; tahmin, türetme, veya proxy. Otoriter görünüyor ama değil.

**Check:** Dashboard'daki HER sayı için sor: bu gerçek bir count mu yoksa hesaplanmış/tahmin mi?
- `chars/4` → ✗ (bugün kaldırıldı)
- İsim uzunluğundan türetilmiş bar yüksekliği → ✗ (önceden fixlendi)
- Gerçek token count (API'den) → ✓

**Real incident:** chars/4 token tahmini "tokens" etiketiyle gösteriliyordu — yanıltıcı.

---

## 3. go:embed stale-serve tuzağı

**Pattern:** `static/` altındaki HTML/JS değişikliği, `go build` + server restart yapılmadan görünmez. Eski binary eski static'i serve eder.

**Check:** Her static edit'ten sonra:
```bash
go build -o wrapminal . && pkill wrapminal; ./wrapminal &
sleep 1 && curl -s localhost:8080 | grep <yeni-içerik>
```
Eğer yeni içerik grep'te çıkmıyorsa → eski binary hâlâ çalışıyor.

**Real incident:** Bu session'da 3 kez oldu: HTML değişti, curl eski sayfayı gösterdi.

---

## 4. Port 8080 orphan server

**Pattern:** Önceki session'dan kalma cc-lens/wrapminal process'i 8080'i tutuyor, eski binary'yi serve ediyor. Gerçek durumu maskeliyor.

**Check:**
```bash
lsof -i :8080 | grep LISTEN
```
Tek bir `wrapminal` process'i olmalı. Fazladan varsa → `kill <pid>`.

**Real incident:** Eski `cc-lens` binary'si 8080'de takılı kalmıştı; yeni build'in çıktısı hiç görünmüyordu.

---

## 5. Gizlilik regresyonu

**Pattern:** Ürünün sözü "hiçbir şey sızmaz." Her yeni render yüzeyi sızdırabilir.

**Check:** Dashboard HTML + SVG + API response'ta şunlar OLMAMALI:
- Ham proje ismi (`<project-name>`, dosya adından türemiş string)
- Uzunluk-korelasyonlu görsel (bar yüksekliği = isim uzunluğu)
- Ham prompt metni
- Dosya yolu

```bash
# API kontrolü
curl -s localhost:8080/api/wrapped | jq '..|strings' | grep -iE '(proje|project|prompt|\.json|\.jsonl)' || echo "clean"

# HTML kontrolü
curl -s localhost:8080 | grep -iE '(proje|project|prompt)' || echo "clean"
```

---

## 6. Perf: her load'da tam re-scan

**Pattern:** `resolved_loops.go:64` — Claude session dosyalarını her istekte sıfırdan tarıyor. Cache yok. Büyük history'de yavaş.

**Check:**
```bash
time curl -s localhost:8080/api/wrapped > /dev/null
```
1 saniyeden uzunsa → cache düşün. Ponytail comment'i orada zaten:
```
ponytail: O(n) scan over all session files per request, no cache.
Upgrade path: in-memory TTL cache keyed by mtime, invalidate on file change.
```

---

## 7. Kaynak sayısı overclaim (README)

**Pattern:** README başlığı "11+ tools" diyor ama gerçekte parse edilen (`state="loaded"`) 5 kaynak var. Detay tablosu dürüst, üst satır iyimser.

**Check:** README.md satır ~1-3'teki araç sayısı iddiası ile `curl -s localhost:8080/api/wrapped | jq '[.sources[] | select(.state=="loaded")] | length'` sayısı eşleşmeli.

**Real incident:** Bu session'da fark edildi; düzeltme kararı Semih'e bırakıldı.

---

## Loop-agent protokolü

Her iterasyon:
1. Yukarıdaki 7 check'i sırayla çalıştır
2. Fail eden var mı? → önce onu fix'le
3. Hepsi yeşilse → asıl task'a geç
4. Herhangi bir check'te yeni bir pattern keşfedersen → bu dosyaya ekle (önce Semih'e sor)

Son güncelleme: 2026-06-21 — receipt dashboard commit'i sonrası.