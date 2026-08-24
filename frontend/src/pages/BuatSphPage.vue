<template>
  <div>
    <PageHeader :title="isEdit ? 'Edit Draft SPH' : 'Buat SPH'" subtitle="Wizard penyusunan penawaran langkah demi langkah">
      <template #actions>
        <button type="button" class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50" @click="router.back()">
          Batal
        </button>
      </template>
    </PageHeader>

    <!-- Stepper -->
    <ol class="mb-5 flex flex-wrap gap-x-2 gap-y-1 rounded-xl border border-slate-200 bg-white px-4 py-3 text-[12px]">
      <li v-for="st in steps" :key="st.n" class="flex items-center gap-2">
        <button
          type="button"
          class="flex h-6 w-6 items-center justify-center rounded-full border font-semibold transition-colors"
          :class="stepBadgeClass(st.n)"
          @click="goTo(st.n)"
        >
          {{ st.n }}
        </button>
        <span :class="step === st.n ? 'font-semibold text-brand-700' : 'text-slate-500'">{{ st.label }}</span>
        <span v-if="st.n < steps.length" class="ml-1 text-slate-300">→</span>
      </li>
    </ol>

    <p v-if="pageError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">{{ pageError }}</p>

    <div class="rounded-xl border border-slate-200 bg-white p-5">
      <template v-if="savedView === null">
      <!-- ===== Langkah 1: Info Dokumen ===== -->
      <section v-show="step === 1">
        <h2 class="mb-4 text-[15px] font-semibold text-slate-800">Info Dokumen</h2>
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Tanggal SPH <span class="text-red-500">*</span></label>
            <input v-model="header.date" type="date" required class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
            <p class="mt-1 text-xs text-slate-400">Nomor dokumen dibuat otomatis dari periode tanggal ini.</p>
          </div>
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Masa Berlaku</label>
            <input v-model="header.validUntil" type="date" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
          </div>
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Customer <span class="text-red-500">*</span></label>
            <select v-model.number="header.customerId" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100">
              <option :value="0" disabled>— Pilih customer —</option>
              <option v-for="c in partnerStore.customers" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </div>
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Kapal</label>
            <select v-model.number="vesselChoice" :disabled="!customerVessels.length" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none disabled:bg-slate-50 disabled:text-slate-400 focus:border-brand-400 focus:ring-2 focus:ring-brand-100">
              <option :value="0">— Tanpa kapal —</option>
              <option v-for="v in customerVessels" :key="v.id" :value="v.id">{{ v.name }}</option>
            </select>
          </div>
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Nama Proyek</label>
            <input v-model="header.projectName" type="text" maxlength="300" placeholder="misal Docking & Repair 2026" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
          </div>
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Subjek</label>
            <input v-model="header.subject" type="text" maxlength="300" placeholder="Penawaran Jasa…" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
          </div>
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Referensi</label>
            <input v-model="header.reference" type="text" maxlength="200" placeholder="No. RFQ / surat masuk" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
          </div>
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Lokasi Pekerjaan</label>
            <input v-model="header.location" type="text" maxlength="200" placeholder="misal Tanjung Priok" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
          </div>
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">PIC Customer</label>
            <input v-model="header.picName" type="text" maxlength="150" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
          </div>
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Catatan Umum</label>
            <textarea v-model="header.notes" rows="2" maxlength="2000" class="w-full resize-none rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100"></textarea>
          </div>
        </div>
      </section>

      <!-- ===== Langkah 2: Sumber Pekerjaan ===== -->
      <section v-show="step === 2">
        <h2 class="mb-1 text-[15px] font-semibold text-slate-800">Sumber Pekerjaan</h2>
        <p class="mb-4 text-[13px] text-slate-500">Dari mana isi penawaran disusun? Semua nilai akan disalin sebagai snapshot.</p>

        <div class="mb-4 flex flex-wrap gap-2">
          <button
            v-for="src in sources"
            :key="src.key"
            type="button"
            class="rounded-lg border px-3.5 py-2 text-[13px] font-medium transition-colors"
            :class="sourceTab === src.key ? 'border-brand-300 bg-brand-50 text-brand-700' : 'border-slate-200 text-slate-600 hover:bg-slate-50'"
            @click="sourceTab = src.key"
          >
            {{ src.label }}
          </button>
        </div>

        <!-- Master -->
        <div v-if="sourceTab === 'master'">
          <div class="mb-3 flex flex-wrap items-end gap-2 rounded-lg border border-slate-200 bg-slate-50/70 p-3">
            <div class="min-w-[220px] flex-1">
              <label class="mb-1 block text-[13px] font-medium text-slate-600">Kategori</label>
              <select v-model="masterCatId" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100">
                <option :value="0">Semua kategori</option>
                <option v-for="c in masterStore.categories" :key="c.id" :value="c.id">{{ c.name }}</option>
              </select>
            </div>
            <div class="min-w-[220px] flex-1">
              <label class="mb-1 block text-[13px] font-medium text-slate-600">Cari pekerjaan</label>
              <input v-model="masterSearch" type="search" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
            </div>
          </div>
          <ul class="divide-y divide-slate-100 rounded-lg border border-slate-200">
            <li v-for="wi in filteredWorkItems" :key="wi.id" class="flex items-center gap-3 px-4 py-2.5 text-[13px]">
              <div class="min-w-0 flex-1">
                <p class="truncate font-medium text-slate-700">{{ wi.name }}</p>
                <p class="truncate text-xs text-slate-400">
                  {{ formatRupiah(wi.defaultServicePrice) }} jasa<span v-if="wi.defaultMaterialPrice"> · {{ formatRupiah(wi.defaultMaterialPrice) }} material</span> · {{ formatQty(wi.defaultQuantity) }} {{ wi.defaultUnit || 'giat' }}
                </p>
              </div>
              <button
                type="button"
                :disabled="items.some((it) => it.workItemId === wi.id)"
                class="shrink-0 rounded-md bg-brand-600 px-2.5 py-1.5 text-xs font-medium text-white transition-colors hover:bg-brand-700 disabled:cursor-not-allowed disabled:bg-slate-200 disabled:text-slate-400"
                @click="addFromWorkItem(wi)"
              >
                {{ items.some((it) => it.workItemId === wi.id) ? 'Sudah ada' : '+ Tambah' }}
              </button>
            </li>
            <li v-if="!filteredWorkItems.length" class="px-4 py-6 text-center text-[13px] italic text-slate-400">Tidak ada pekerjaan aktif yang cocok.</li>
          </ul>
        </div>

        <!-- Template -->
        <div v-else-if="sourceTab === 'template'">
          <ul class="divide-y divide-slate-100 rounded-lg border border-slate-200">
            <li v-for="tpl in templateStore.templates.filter((t) => t.isActive)" :key="tpl.id" class="flex items-center gap-3 px-4 py-2.5 text-[13px]">
              <div class="min-w-0 flex-1">
                <p class="truncate font-medium text-slate-700">{{ tpl.name }}</p>
                <p class="truncate text-xs text-slate-400">{{ tpl.itemCount }} pekerjaan</p>
              </div>
              <button type="button" class="shrink-0 rounded-md bg-brand-600 px-2.5 py-1.5 text-xs font-medium text-white transition-colors hover:bg-brand-700" @click="applyTemplate(tpl)">
                Pakai Template
              </button>
            </li>
            <li v-if="!templateStore.templates.length" class="px-4 py-6 text-center text-[13px] italic text-slate-400">Belum ada template aktif.</li>
          </ul>
        </div>

        <!-- SPH lama -->
        <div v-else-if="sourceTab === 'old'">
          <ul class="divide-y divide-slate-100 rounded-lg border border-slate-200">
            <li v-for="doc in oldDocs" :key="doc.id" class="flex items-center gap-3 px-4 py-2.5 text-[13px]">
              <div class="min-w-0 flex-1">
                <p class="truncate font-mono text-xs font-semibold text-slate-700">{{ doc.documentNumber }}</p>
                <p class="truncate text-xs text-slate-400">{{ doc.customerName }} · {{ doc.projectName || 'tanpa proyek' }} · {{ doc.itemCount }} pekerjaan</p>
              </div>
              <button type="button" class="shrink-0 rounded-md bg-brand-600 px-2.5 py-1.5 text-xs font-medium text-white transition-colors hover:bg-brand-700" @click="applyOldDoc(doc)">
                Salin Isi
              </button>
            </li>
            <li v-if="!oldDocs.length" class="px-4 py-6 text-center text-[13px] italic text-slate-400">Belum ada dokumen SPH lain.</li>
          </ul>
        </div>

        <!-- Manual -->
        <div v-else class="rounded-lg border border-dashed border-slate-300 p-6 text-center">
          <p class="mb-3 text-[13px] text-slate-500">Susun sendiri baris pekerjaan tanpa data master.</p>
          <button type="button" class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700" @click="addManual()">
            + Tambah Baris Manual
          </button>
        </div>

        <p v-if="items.length" class="mt-3 text-[13px] text-emerald-700">{{ items.length }} baris pekerjaan siap disusun di langkah berikutnya.</p>
      </section>

      <!-- ===== Langkah 3: Susun Main Point ===== -->
      <section v-show="step === 3">
        <h2 class="mb-1 text-[15px] font-semibold text-slate-800">Susun Main Point</h2>
        <p class="mb-4 text-[13px] text-slate-500">Atur urutan, nama, qty, dan harga tiap pekerjaan utama.</p>

        <div v-if="!items.length" class="rounded-lg border border-dashed border-slate-300 p-8 text-center text-[13px] text-slate-400">
          Belum ada baris. Kembali ke langkah 2 untuk menambah sumber.
        </div>

        <div class="space-y-3">
          <div v-for="(it, idx) in items" :key="it.uid" class="rounded-lg border border-slate-200 p-3">
            <div class="mb-2 flex items-center gap-2">
              <span class="flex h-6 w-6 items-center justify-center rounded-full bg-brand-50 text-xs font-semibold text-brand-700">{{ idx + 1 }}</span>
              <span class="flex-1 truncate text-[13px] font-medium text-slate-700">{{ it.name || '(belum diberi nama)' }}</span>
              <button type="button" :disabled="idx === 0" class="rounded px-1.5 py-1 text-xs text-slate-500 hover:bg-slate-100 disabled:opacity-30" @click="moveItem(idx, -1)">↑</button>
              <button type="button" :disabled="idx === items.length - 1" class="rounded px-1.5 py-1 text-xs text-slate-500 hover:bg-slate-100 disabled:opacity-30" @click="moveItem(idx, 1)">↓</button>
              <button type="button" class="rounded px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-50" @click="removeItem(idx)">Hapus</button>
            </div>
            <div class="grid grid-cols-1 gap-2.5 md:grid-cols-12">
              <div class="md:col-span-3">
                <label class="mb-1 block text-xs font-medium text-slate-500">Nama</label>
                <input v-model="it.name" type="text" maxlength="300" class="row-input" />
              </div>
              <div class="md:col-span-2">
                <label class="mb-1 block text-xs font-medium text-slate-500">Deskripsi</label>
                <input v-model="it.description" type="text" maxlength="500" class="row-input" />
              </div>
              <div class="md:col-span-2">
                <label class="mb-1 block text-xs font-medium text-slate-500">Mode Harga</label>
                <select v-model="it.pricingMode" class="row-input">
                  <option value="HARGA_LANGSUNG">Harga Langsung</option>
                  <option value="PEMBOBOTAN">Pembobotan</option>
                </select>
              </div>
              <div class="md:col-span-1">
                <label class="mb-1 block text-xs font-medium text-slate-500">Qty</label>
                <input v-model.number="it.quantity" type="number" min="0" step="0.001" class="row-input" />
              </div>
              <div class="md:col-span-1">
                <label class="mb-1 block text-xs font-medium text-slate-500">Satuan</label>
                <input v-model="it.unit" type="text" maxlength="30" class="row-input" />
              </div>
              <div class="md:col-span-1">
                <label class="mb-1 block text-xs font-medium text-slate-500">Jasa</label>
                <input v-model.number="it.serviceUnitPrice" type="number" min="0" class="row-input" />
              </div>
              <div class="md:col-span-1">
                <label class="mb-1 block text-xs font-medium text-slate-500">Mat.</label>
                <input v-model.number="it.materialUnitPrice" type="number" min="0" class="row-input" />
              </div>
            </div>
            <div class="mt-1 flex items-center justify-between gap-2">
              <span v-if="it.pricingMode === 'PEMBOBOTAN'" class="text-xs font-medium text-brand-700">
                Pembobotan — nilai {{ formatRupiah(rowTotal(it)) }} dibagi ke sub point via bobot (langkah 4).
              </span>
              <span v-else></span>
              <span class="whitespace-nowrap text-right text-[13px] font-semibold tabular-nums text-slate-700">{{ formatRupiah(rowTotal(it)) }}</span>
            </div>
            <div v-if="it.subItems.length" class="mt-2 border-t border-slate-100 pt-2">
              <p class="text-xs text-slate-400">{{ it.subItems.length }} sub point ikut dalam baris ini (atur di langkah 4).</p>
            </div>
          </div>
        </div>
      </section>

      <!-- ===== Langkah 4: Sub Point ===== -->
      <section v-show="step === 4">
        <h2 class="mb-1 text-[15px] font-semibold text-slate-800">Sub Point</h2>
        <p class="mb-4 text-[13px] text-slate-500">
          Mode Harga Langsung: nilai sub di-roll-up ke induknya. Mode Pembobotan: nilai induk dibagikan ke sub sesuai bobot %.
        </p>

        <div v-if="!items.length" class="rounded-lg border border-dashed border-slate-300 p-8 text-center text-[13px] text-slate-400">Belum ada main point.</div>

        <template v-else>
          <div class="mb-3 flex flex-wrap gap-1.5">
            <button
              v-for="(it, idx) in items"
              :key="it.uid"
              type="button"
              class="max-w-[240px] truncate rounded-full border px-3 py-1.5 text-xs font-medium transition-colors"
              :class="activeItemIdx === idx ? 'border-brand-300 bg-brand-50 text-brand-700' : 'border-slate-200 text-slate-600 hover:bg-slate-50'"
              @click="activeItemIdx = idx"
            >
              {{ idx + 1 }}. {{ it.name || '(tanpa nama)' }}
              <span v-if="isWeighted(it)" class="ml-1 font-semibold">Σ{{ weightSumOf(it) }}%</span>
            </button>
          </div>

          <div v-if="activeItem" class="space-y-2.5">
            <div
              v-if="isWeighted(activeItem)"
              class="flex flex-wrap items-center gap-x-2 rounded-lg border px-3 py-2 text-[13px]"
              :class="weightSumOf(activeItem) === 100 ? 'border-emerald-200 bg-emerald-50 text-emerald-700' : 'border-amber-200 bg-amber-50 text-amber-700'"
            >
              <span>Σ bobot: <strong>{{ weightSumOf(activeItem) }}%</strong></span>
              <span v-if="weightSumOf(activeItem) !== 100">
                — selisih {{ selisihLabel(weightSumOf(activeItem)) }} dari 100%. Draft tetap bisa disimpan, tapi finalisasi ditolak sampai tepat 100%.
              </span>
            </div>

            <div
              v-for="(sub, sIdx) in activeItem.subItems"
              :key="sIdx"
              class="grid grid-cols-1 items-end gap-2.5 rounded-lg border border-slate-200 p-3 md:grid-cols-12"
            >
              <div class="md:col-span-4">
                <label class="mb-1 block text-xs font-medium text-slate-500">Nama Sub {{ sIdx + 1 }}</label>
                <input v-model="sub.name" type="text" maxlength="300" class="row-input" />
              </div>
              <div class="md:col-span-3">
                <label class="mb-1 block text-xs font-medium text-slate-500">Deskripsi</label>
                <input v-model="sub.description" type="text" maxlength="500" class="row-input" />
              </div>
              <div class="md:col-span-1">
                <label class="mb-1 block text-xs font-medium text-slate-500">Qty</label>
                <input v-model.number="sub.quantity" type="number" min="0" step="0.001" class="row-input" />
              </div>
              <div class="md:col-span-1">
                <label class="mb-1 block text-xs font-medium text-slate-500">Satuan</label>
                <input v-model="sub.unit" type="text" maxlength="30" class="row-input" />
              </div>
              <template v-if="isWeighted(activeItem)">
                <div class="md:col-span-2">
                  <label class="mb-1 block text-xs font-medium text-brand-700">Bobot %</label>
                  <input v-model.number="sub.weight" type="number" min="0" max="100" class="row-input" />
                  <p class="mt-0.5 text-[11px] tabular-nums text-slate-400">≈ {{ formatRupiah(subJumlah(activeItem, sIdx)) }}</p>
                </div>
                <div class="flex items-end justify-end md:col-span-1">
                  <button type="button" class="rounded px-2 py-1.5 text-xs font-medium text-red-600 hover:bg-red-50" @click="activeItem.subItems.splice(sIdx, 1)">✕</button>
                </div>
              </template>
              <template v-else>
                <div class="md:col-span-1">
                  <label class="mb-1 block text-xs font-medium text-slate-500">Jasa</label>
                  <input v-model.number="sub.serviceUnitPrice" type="number" min="0" class="row-input" />
                </div>
                <div class="md:col-span-2 flex items-center justify-between gap-2">
                  <input v-model.number="sub.materialUnitPrice" type="number" min="0" title="Harga material" class="row-input" />
                  <button type="button" class="shrink-0 rounded px-2 py-1.5 text-xs font-medium text-red-600 hover:bg-red-50" @click="activeItem.subItems.splice(sIdx, 1)">✕</button>
                </div>
              </template>
            </div>
            <button type="button" class="rounded-lg border border-brand-200 bg-brand-50 px-3.5 py-2 text-[13px] font-medium text-brand-700 transition-colors hover:bg-brand-100" @click="addSub(activeItem)">
              + Sub Point
            </button>
          </div>
        </template>
      </section>

      <!-- ===== Langkah 5: Costing ===== -->
      <section v-show="step === 5">
        <h2 class="mb-1 text-[15px] font-semibold text-slate-800">Ringkasan Harga</h2>
        <p class="mb-4 text-[13px] text-slate-500">Perhitungan otomatis dari qty × harga satuan, dibulatkan ke Rupiah terdekat (BR-04).</p>

        <table class="w-full text-left text-[13px]">
          <thead>
            <tr class="border-b border-slate-100 text-xs uppercase tracking-wide text-slate-400">
              <th class="py-2 pr-3 font-medium">#</th>
              <th class="px-3 py-2 font-medium">Pekerjaan</th>
              <th class="px-3 py-2 text-right font-medium">Qty × Harga</th>
              <th class="px-3 py-2 text-right font-medium">Jasa</th>
              <th class="px-3 py-2 text-right font-medium">Material</th>
              <th class="px-3 py-2 text-right font-medium">Total</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(it, idx) in items" :key="it.uid" class="border-b border-slate-50">
              <td class="py-2 pr-3 text-slate-400">{{ idx + 1 }}</td>
              <td class="px-3 py-2 font-medium text-slate-700">{{ it.name || '(tanpa nama)' }}</td>
              <td class="whitespace-nowrap px-3 py-2 text-right tabular-nums text-slate-500">
                {{ formatQty(it.quantity) }} × {{ formatRupiah(it.serviceUnitPrice + it.materialUnitPrice) }}
              </td>
              <td class="px-3 py-2 text-right tabular-nums text-slate-600">{{ formatRupiah(rowService(it)) }}</td>
              <td class="px-3 py-2 text-right tabular-nums text-slate-600">{{ formatRupiah(rowMaterial(it)) }}</td>
              <td class="px-3 py-2 text-right font-semibold tabular-nums text-slate-800">{{ formatRupiah(rowTotal(it)) }}</td>
            </tr>
          </tbody>
          <tfoot>
            <tr class="border-t-2 border-slate-200 font-semibold text-slate-700">
              <td colspan="3" class="py-2.5 pl-3">Subtotal</td>
              <td class="px-3 py-2.5 text-right tabular-nums">{{ formatRupiah(totals.service) }}</td>
              <td class="px-3 py-2.5 text-right tabular-nums">{{ formatRupiah(totals.material) }}</td>
              <td class="px-3 py-2.5 text-right tabular-nums">{{ formatRupiah(totals.grand) }}</td>
            </tr>
            <tr class="text-brand-700">
              <td colspan="5" class="py-2.5 pl-3">Grand Total</td>
              <td class="px-3 py-2.5 text-right text-[15px] tabular-nums">{{ formatRupiah(totals.grand) }}</td>
            </tr>
          </tfoot>
        </table>
        <p class="mt-2 text-[13px] italic text-slate-500">{{ terbilangText || '…' }}</p>
      </section>

      <!-- ===== Langkah 6: Validasi ===== -->
      <section v-show="step === 6">
        <h2 class="mb-1 text-[15px] font-semibold text-slate-800">Validasi</h2>
        <p class="mb-4 text-[13px] text-slate-500">Pemeriksaan sebelum finalisasi nanti (BR-06). Perbaiki yang merah.</p>
        <ul class="space-y-2">
          <li v-for="chk in checks" :key="chk.label" class="flex items-start gap-2.5 rounded-lg border px-3.5 py-2.5 text-[13px]" :class="chk.ok ? 'border-emerald-200 bg-emerald-50/60 text-emerald-800' : 'border-red-200 bg-red-50/60 text-red-700'">
            <span>{{ chk.ok ? '✓' : '✕' }}</span>
            <span>{{ chk.label }}</span>
          </li>
        </ul>

        <template v-if="warnings.length">
          <h3 class="mb-2 mt-5 text-[13px] font-semibold uppercase tracking-wide text-slate-400">Peringatan Pembobotan (tidak menghalangi simpan draft)</h3>
          <ul class="space-y-2">
            <li v-for="wrn in warnings" :key="wrn.label" class="flex items-start gap-2.5 rounded-lg border px-3.5 py-2.5 text-[13px]" :class="wrn.ok ? 'border-emerald-200 bg-emerald-50/60 text-emerald-800' : 'border-amber-200 bg-amber-50/60 text-amber-700'">
              <span>{{ wrn.ok ? '✓' : '!' }}</span>
              <span>{{ wrn.label }}</span>
            </li>
          </ul>
        </template>
      </section>

      <!-- ===== Langkah 7: Preview ===== -->
      <section v-show="step === 7">
        <h2 class="mb-4 text-[15px] font-semibold text-slate-800">Preview Dokumen</h2>
        <article class="mx-auto max-w-3xl rounded-lg border border-slate-200 p-8 text-[13px] leading-relaxed">
          <header class="mb-6 border-b border-slate-200 pb-4">
            <p class="font-mono text-sm font-bold tracking-wide text-slate-800">{{ numberPreview }}</p>
            <p class="mt-0.5 text-slate-500">{{ todayLabel }}</p>
          </header>
          <div class="mb-6 space-y-1 text-slate-700">
            <p>Kepada Yth.</p>
            <p class="font-semibold">{{ selectedCustomer?.name || '—' }}</p>
            <p v-if="selectedVessel">a.n KM/Vessel: {{ selectedVessel.name }}</p>
            <p v-if="header.picName">u.p. {{ header.picName }}</p>
          </div>
          <div class="mb-6">
            <p><span class="inline-block w-28 text-slate-500">Perihal</span>: {{ header.subject || '—' }}</p>
            <p><span class="inline-block w-28 text-slate-500">Proyek</span>: {{ header.projectName || '—' }}</p>
            <p v-if="header.location"><span class="inline-block w-28 text-slate-500">Lokasi</span>: {{ header.location }}</p>
            <p v-if="header.reference"><span class="inline-block w-28 text-slate-500">Referensi</span>: {{ header.reference }}</p>
          </div>
          <table class="mb-6 w-full text-left">
            <thead>
              <tr class="border-y border-slate-300 text-xs uppercase tracking-wide text-slate-500">
                <th class="py-1.5 pr-2 font-medium">No</th>
                <th class="px-2 py-1.5 font-medium">Uraian Pekerjaan</th>
                <th class="px-2 py-1.5 text-right font-medium">Qty</th>
                <th class="px-2 py-1.5 text-center font-medium">Sat</th>
                <th class="px-2 py-1.5 text-right font-medium">Jasa</th>
                <th class="px-2 py-1.5 text-right font-medium">Mat.</th>
                <th class="pl-2 py-1.5 text-right font-medium">Jumlah</th>
              </tr>
            </thead>
            <tbody>
              <template v-for="(it, idx) in items" :key="it.uid">
                <tr class="align-top">
                  <td class="py-1.5 pr-2 text-slate-500">{{ idx + 1 }}</td>
                  <td class="px-2 py-1.5">
                    <p class="font-medium text-slate-800">{{ it.name }}</p>
                    <p v-if="it.description" class="text-slate-500">{{ it.description }}</p>
                  </td>
                  <td class="px-2 py-1.5 text-right tabular-nums">{{ formatQty(it.quantity) }}</td>
                  <td class="px-2 py-1.5 text-center">{{ it.unit }}</td>
                  <td class="whitespace-nowrap px-2 py-1.5 text-right tabular-nums">{{ formatRupiah(it.serviceUnitPrice) }}</td>
                  <td class="whitespace-nowrap px-2 py-1.5 text-right tabular-nums">{{ formatRupiah(it.materialUnitPrice) }}</td>
                  <td class="py-1.5 pl-2 text-right tabular-nums">{{ formatRupiah(rowTotal(it)) }}</td>
                </tr>
                <tr v-for="(sub, sIdx) in it.subItems" :key="`${it.uid}-${sIdx}`" class="align-top text-[12px] text-slate-500">
                  <td class="py-1 pr-2"></td>
                  <td class="py-1 pl-8 pr-2">
                    <p>
                      — {{ sub.name }}
                      <span v-if="isWeighted(it)" class="ml-0.5 rounded bg-brand-50 px-1 text-[10px] font-semibold text-brand-700">{{ sub.weight }}%</span>
                    </p>
                    <p v-if="sub.description" class="text-slate-400">{{ sub.description }}</p>
                  </td>
                  <td class="px-2 py-1 text-right tabular-nums">{{ formatQty(sub.quantity) }}</td>
                  <td class="px-2 py-1 text-center">{{ sub.unit }}</td>
                  <td class="whitespace-nowrap px-2 py-1 text-right tabular-nums">{{ isWeighted(it) ? '—' : formatRupiah(sub.serviceUnitPrice) }}</td>
                  <td class="whitespace-nowrap px-2 py-1 text-right tabular-nums">{{ isWeighted(it) ? '—' : formatRupiah(sub.materialUnitPrice) }}</td>
                  <td class="py-1 pl-2 text-right tabular-nums">{{ formatRupiah(subJumlah(it, sIdx)) }}</td>
                </tr>
              </template>
            </tbody>
            <tfoot>
              <tr class="border-t border-slate-300 font-semibold">
                <td colspan="6" class="py-2 text-right">Grand Total</td>
                <td class="py-2 pl-2 text-right tabular-nums">{{ formatRupiah(totals.grand) }}</td>
              </tr>
            </tfoot>
          </table>
          <p class="italic text-slate-600">Terbilang: <strong>{{ terbilangText }}</strong></p>
        </article>
      </section>

      <!-- ===== Langkah 8: Simpan ===== -->
      <section v-show="step === 8">
        <h2 class="mb-1 text-[15px] font-semibold text-slate-800">Simpan Draft</h2>
        <p class="mb-4 text-[13px] text-slate-500">Draft tersimpan dengan nomor otomatis; status tetap Draft sampai difinalisasi dari halaman detail.</p>
        <dl class="mb-5 grid grid-cols-2 gap-x-6 gap-y-1 rounded-lg border border-slate-200 bg-slate-50/70 p-4 text-[13px] md:grid-cols-4">
          <dt class="text-slate-500">Customer</dt><dd class="font-medium text-slate-700">{{ selectedCustomer?.name || '—' }}</dd>
          <dt class="text-slate-500">Baris</dt><dd class="font-medium text-slate-700">{{ items.length }} pekerjaan</dd>
          <dt class="text-slate-500">Grand Total</dt><dd class="font-medium tabular-nums text-slate-700">{{ formatRupiah(totals.grand) }}</dd>
          <dt class="text-slate-500">Status</dt><dd class="font-medium text-slate-700">Draft</dd>
        </dl>
        <p v-if="saveError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px] text-red-700">{{ saveError }}</p>
        <button type="button" :disabled="saving || !allChecksOk" class="rounded-lg bg-brand-600 px-5 py-2.5 text-[13px] font-semibold text-white transition-colors hover:bg-brand-700 disabled:opacity-60" @click="save()">
          {{ saving ? 'Menyimpan…' : isEdit ? 'Simpan Perubahan Draft' : 'Simpan Draft SPH' }}
        </button>
        <p v-if="!allChecksOk" class="mt-2 text-xs text-slate-400">Lengkapi pemeriksaan validasi di langkah 6 terlebih dahulu.</p>
      </section>

      <!-- Navigasi wizard -->
      <footer class="mt-6 flex items-center justify-between border-t border-slate-100 pt-4">
        <button type="button" :disabled="step === 1" class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:opacity-40" @click="goTo(step - 1)">
          ← Sebelumnya
        </button>
        <span class="text-xs text-slate-400">Langkah {{ step }} dari {{ steps.length }}</span>
        <button
          v-if="step < steps.length"
          type="button"
          :disabled="!canAdvance"
          class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-40"
          @click="goTo(step + 1)"
        >
          Lanjut →
        </button>
        <span v-else></span>
      </footer>
      </template>

      <!-- ===== Selesai ===== -->
      <section v-else>
        <div class="mx-auto max-w-md rounded-xl border border-emerald-200 bg-emerald-50/70 p-6 text-center">
          <p class="text-2xl">✓</p>
          <h3 class="mt-2 text-[15px] font-semibold text-emerald-900">Draft SPH tersimpan</h3>
          <p class="mt-1 font-mono text-sm text-emerald-800">{{ savedView?.documentNumber }}</p>
          <p class="mt-1 text-[13px] text-emerald-700">Grand total {{ formatRupiah(savedView?.grandTotal ?? 0) }} — {{ savedView?.terbilang }}</p>
          <div class="mt-4 flex justify-center gap-2">
            <button type="button" class="rounded-lg bg-brand-600 px-4 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700" @click="router.push(`/sph/${savedView?.id}`)">
              Buka Detail
            </button>
            <button type="button" class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50" @click="router.push('/sph')">
              Ke Daftar SPH
            </button>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import { usePartnerStore } from '../stores/partner'
import { useMasterStore } from '../stores/master'
import { useTemplateStore } from '../stores/template'
import { useSphStore } from '../stores/sph'
import { Terbilang } from '../../wailsjs/go/main/App'
import { formatRupiah, formatQty, errorMessage } from '../utils/format'
import type { CustomerView, VesselView } from '../types/partner'
import type { WorkItemView } from '../types/master'
import type { TemplateView } from '../types/template'
import type { SphDocumentView, SphHeaderInput, SphItemInput, SphSubItemInput } from '../types/sph'

const route = useRoute()
const router = useRouter()
const partnerStore = usePartnerStore()
const masterStore = useMasterStore()
const templateStore = useTemplateStore()
const sphStore = useSphStore()

// ===== wizard state =====
const steps = [
  { n: 1, label: 'Info' },
  { n: 2, label: 'Sumber' },
  { n: 3, label: 'Main Point' },
  { n: 4, label: 'Sub Point' },
  { n: 5, label: 'Costing' },
  { n: 6, label: 'Validasi' },
  { n: 7, label: 'Preview' },
  { n: 8, label: 'Simpan' }
]

const isEdit = computed(() => !!route.params.id)
const editId = computed(() => Number(route.params.id || 0))
const step = ref(1)
const pageError = ref('')
const saveError = ref('')
const saving = ref(false)
const savedView = ref<SphDocumentView | null>(null)

function emptyHeader(): SphHeaderInput {
  const today = new Date()
  const ymd =
    today.getFullYear() +
    '-' +
    String(today.getMonth() + 1).padStart(2, '0') +
    '-' +
    String(today.getDate()).padStart(2, '0')
  return {
    date: ymd,
    customerId: 0,
    vesselId: undefined,
    projectName: '',
    subject: '',
    reference: '',
    location: '',
    validUntil: '',
    picName: '',
    notes: ''
  }
}

let uidSeq = 0
function nextUid(): number {
  uidSeq -= 1
  return uidSeq
}

const header = reactive(emptyHeader())
interface WizardRow {
  uid: number
}
const items = ref<(WizardRow & SphItemInput)[]>([])
const vesselChoice = ref(0)

const sourceTab = ref<'master' | 'template' | 'old' | 'manual'>('master')
const sources = [
  { key: 'master', label: 'Master Pekerjaan' },
  { key: 'template', label: 'Template' },
  { key: 'old', label: 'SPH Lama' },
  { key: 'manual', label: 'Manual' }
] as const

const masterCatId = ref(0)
const masterSearch = ref('')
const oldDocs = ref<SphDocumentView[]>([])
const activeItemIdx = ref(0)

// ===== turunan =====
const selectedCustomer = computed<CustomerView | undefined>(() =>
  partnerStore.customers.find((c) => c.id === header.customerId)
)
const customerVessels = computed<VesselView[]>(() =>
  (selectedCustomer.value?.vessels ?? []).filter((v) => v.isActive)
)
const selectedVessel = computed(() => customerVessels.value.find((v) => v.id === vesselChoice.value))

// Ganti customer → pilihan kapal di-reset bila tidak termasuk miliknya.
watch(
  () => header.customerId,
  () => {
    if (!customerVessels.value.some((v) => v.id === vesselChoice.value)) {
      vesselChoice.value = 0
    }
  }
)

const filteredWorkItems = computed<WorkItemView[]>(() => {
  const q = masterSearch.value.trim().toLowerCase()
  return masterStore.workItems.filter(
    (wi) =>
      wi.isActive &&
      (masterCatId.value === 0 || wi.categoryId === masterCatId.value) &&
      (!q || wi.name.toLowerCase().includes(q))
  )
})

// ===== pembobotan (BR-02..BR-04) =====
function isWeighted(it: SphItemInput): boolean {
  return it.pricingMode === 'PEMBOBOTAN'
}

function weightSumOf(it: SphItemInput): number {
  return it.subItems.reduce((acc, s) => acc + (s.weight || 0), 0)
}

function selisihLabel(sum: number): string {
  const diff = 100 - sum
  return `${diff > 0 ? '+' : ''}${diff}%`
}

// Mirror Go allocateLargestRemainder: dasar floor(pool×w/Σw), sisa ke pecahan
// terbesar, tie-break urutan baris. Σ hasil selalu tepat = pool.
function allocateLargestRemainder(pool: number, weights: number[]): number[] {
  const out = new Array<number>(weights.length).fill(0)
  const sumW = weights.reduce((a, b) => a + b, 0)
  if (!weights.length || sumW <= 0 || pool === 0) return out
  const bases = weights.map((w) => Math.floor((pool * w) / sumW))
  const order = weights
    .map((w, i) => ({ frac: (pool * w) % sumW, i }))
    .sort((a, b) => b.frac - a.frac || a.i - b.i)
  let leftover = pool - bases.reduce((a, b) => a + b, 0)
  for (const o of order) {
    if (leftover <= 0) break
    out[o.i]++
    leftover--
  }
  return bases.map((b, i) => b + out[i])
}

function weightedShares(it: SphItemInput): { svc: number[]; mat: number[] } {
  const weights = it.subItems.map((s) => s.weight || 0)
  return {
    svc: allocateLargestRemainder(Math.round((it.quantity || 0) * (it.serviceUnitPrice || 0)), weights),
    mat: allocateLargestRemainder(Math.round((it.quantity || 0) * (it.materialUnitPrice || 0)), weights)
  }
}

// Jumlah sebuah sub point: hasil alokasi (pembobotan) atau qty×harga (langsung).
function subJumlah(it: SphItemInput, sIdx: number): number {
  const sub = it.subItems[sIdx]
  if (!sub) return 0
  if (isWeighted(it)) {
    const shares = weightedShares(it)
    return (shares.svc[sIdx] || 0) + (shares.mat[sIdx] || 0)
  }
  return subLineTotal(sub)
}

function rowService(it: SphItemInput): number {
  const main = Math.round((it.quantity || 0) * (it.serviceUnitPrice || 0))
  if (isWeighted(it)) return main
  let total = main
  for (const s of it.subItems) total += Math.round((s.quantity || 0) * (s.serviceUnitPrice || 0))
  return total
}
function rowMaterial(it: SphItemInput): number {
  const main = Math.round((it.quantity || 0) * (it.materialUnitPrice || 0))
  if (isWeighted(it)) return main
  let total = main
  for (const s of it.subItems) total += Math.round((s.quantity || 0) * (s.materialUnitPrice || 0))
  return total
}
function rowTotal(it: SphItemInput): number {
  return rowService(it) + rowMaterial(it)
}
function subLineTotal(sub: SphSubItemInput): number {
  return Math.round((sub.quantity || 0) * ((sub.serviceUnitPrice || 0) + (sub.materialUnitPrice || 0)))
}
const totals = computed(() => {
  let service = 0
  let material = 0
  for (const it of items.value) {
    service += rowService(it)
    material += rowMaterial(it)
  }
  return { service, material, grand: service + material }
})

const terbilangText = ref('')
async function refreshTerbilang() {
  try {
    terbilangText.value = await Terbilang(Math.round(totals.value.grand))
  } catch {
    terbilangText.value = ''
  }
}

const checks = computed(() => [
  { ok: !!header.date, label: 'Tanggal SPH terisi.' },
  { ok: header.customerId > 0, label: 'Customer dipilih.' },
  { ok: items.value.length > 0, label: 'Minimal satu pekerjaan utama.' },
  { ok: items.value.every((it) => (it.quantity || 0) > 0 && !!it.name.trim()), label: 'Semua baris punya nama dan qty > 0.' },
  {
    ok: items.value.every(
      (it) =>
        (it.serviceUnitPrice || 0) >= 0 &&
        (it.materialUnitPrice || 0) >= 0 &&
        it.subItems.every((s) => (s.serviceUnitPrice || 0) >= 0 && (s.materialUnitPrice || 0) >= 0)
    ),
    label: 'Tidak ada harga negatif.'
  }
])
const canAdvance = computed(() => {
  if (step.value === 1) return !!header.date && header.customerId > 0
  return true
})
const allChecksOk = computed(() => checks.value.every((c) => c.ok))

// Warning non-blokir (BR-03): draft boleh disimpan dengan Σ≠100,
// tapi finalisasi akan ditolak backend sampai Σ bobot tepat 100%.
const warnings = computed(() =>
  items.value
    .filter(isWeighted)
    .map((it) => {
      const sum = weightSumOf(it)
      const label = it.name.trim() || '(tanpa nama)'
      return {
        ok: sum === 100,
        label:
          sum === 100
            ? `Σ bobot "${label}" = 100%.`
            : `Σ bobot "${label}" = ${sum}%, selisih ${selisihLabel(sum)} dari 100%. Finalisasi akan ditolak sampai tepat 100%.`
      }
    })
)

const numberPreview = computed(
  () => `SPH/GEI/${romanMonth(header.date)}/${header.date.slice(0, 4)}/XXX`
)
function romanMonth(dateStr: string): string {
  const m = Number(dateStr.slice(5, 7))
  const romans = ['', 'I', 'II', 'III', 'IV', 'V', 'VI', 'VII', 'VIII', 'IX', 'X', 'XI', 'XII']
  return romans[m] || '?'
}
const todayLabel = computed(() => new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }))

// ===== aksi =====
async function addFromWorkItem(wi: WorkItemView) {
  if (items.value.some((it) => it.workItemId === wi.id)) return
  try {
    const detail = await masterStore.getWorkItemDetail(wi.id)
    items.value.push({
      uid: nextUid(),
      workItemId: wi.id,
      name: detail.name,
      description: detail.description,
      quantity: detail.defaultQuantity > 0 ? detail.defaultQuantity : 1,
      unit: detail.defaultUnit,
      serviceUnitPrice: detail.defaultServicePrice,
      materialUnitPrice: detail.defaultMaterialPrice,
      pricingMode: 'HARGA_LANGSUNG',
      notes: '',
      subItems: (detail.subItems ?? [])
        .filter((s) => s.isActive)
        .map((s): SphSubItemInput => ({
          name: s.name,
          description: s.description,
          quantity: s.defaultQuantity > 0 ? s.defaultQuantity : 1,
          unit: s.defaultUnit,
          serviceUnitPrice: s.defaultServicePrice,
          materialUnitPrice: s.defaultMaterialPrice,
          weight: s.difficultyWeight || 0,
          notes: ''
        }))
    })
  } catch (e) {
    // fallback tanpa detail
    items.value.push(manualFromWorkItem(wi))
  }
}

function manualFromWorkItem(wi: WorkItemView): WizardRow & SphItemInput {
  return {
    uid: nextUid(),
    workItemId: wi.id,
    name: wi.name,
    description: wi.description,
    quantity: wi.defaultQuantity > 0 ? wi.defaultQuantity : 1,
    unit: wi.defaultUnit,
    serviceUnitPrice: wi.defaultServicePrice,
    materialUnitPrice: wi.defaultMaterialPrice,
    pricingMode: 'HARGA_LANGSUNG',
    notes: '',
    subItems: []
  }
}

async function applyTemplate(tpl: TemplateView) {
  pageError.value = ''
  try {
    const detail = await templateStore.getTemplateDetail(tpl.id)
    for (const row of detail.items ?? []) {
      if (!row.workItem) continue
      const fake: WorkItemView = {
        id: row.workItemId,
        categoryId: 0,
        code: row.workItem.code ?? '',
        name: row.workItem.name ?? '',
        description: row.workItem.description ?? '',
        defaultUnit: row.workItem.defaultUnit ?? '',
        defaultQuantity: row.workItem.defaultQuantity ?? 1,
        defaultServicePrice: row.workItem.defaultServicePrice ?? 0,
        defaultMaterialPrice: row.workItem.defaultMaterialPrice ?? 0,
        notes: row.notes,
        sequence: 0,
        isActive: true,
        subItemCount: 0,
        createdAt: '',
        updatedAt: ''
      }
      if (!items.value.some((it) => it.workItemId === fake.id)) await addFromWorkItem(fake)
    }
  } catch (e) {
    pageError.value = errorMessage(e)
  }
}

async function applyOldDoc(doc: SphDocumentView) {
  pageError.value = ''
  try {
    const detail = await sphStore.getDetail(doc.id)
    items.value = (detail.items ?? []).map((row) => ({
      uid: nextUid(),
      workItemId: row.workItemId ?? undefined,
      name: row.nameSnapshot,
      description: row.descriptionSnapshot,
      quantity: row.quantity,
      unit: row.unit,
      serviceUnitPrice: row.serviceUnitPrice,
      materialUnitPrice: row.materialUnitPrice,
      pricingMode: row.pricingMode || 'HARGA_LANGSUNG',
      notes: row.notes,
      subItems: (row.subItems ?? []).map((s): SphSubItemInput => ({
        name: s.nameSnapshot,
        description: s.descriptionSnapshot,
        quantity: s.quantity,
        unit: s.unit,
        serviceUnitPrice: s.serviceUnitPrice,
        materialUnitPrice: s.materialUnitPrice,
        weight: s.weight ?? 0,
        notes: s.notes
      }))
    }))
  } catch (e) {
    pageError.value = errorMessage(e)
  }
}

function addManual() {
  items.value.push({
    uid: nextUid(),
    workItemId: undefined,
    name: '',
    description: '',
    quantity: 1,
    unit: 'giat',
    serviceUnitPrice: 0,
    materialUnitPrice: 0,
    pricingMode: 'HARGA_LANGSUNG',
    notes: '',
    subItems: []
  })
}

const activeItem = computed(() => items.value[activeItemIdx.value] ?? null)

function addSub(parent: WizardRow & SphItemInput) {
  parent.subItems.push({ name: '', description: '', quantity: 1, unit: 'giat', serviceUnitPrice: 0, materialUnitPrice: 0, weight: 0, notes: '' })
}

function removeItem(idx: number) {
  items.value.splice(idx, 1)
  if (activeItemIdx.value >= items.value.length) activeItemIdx.value = items.value.length - 1
}

function moveItem(idx: number, dir: -1 | 1) {
  const target = idx + dir
  if (target < 0 || target >= items.value.length) return
  const [row] = items.value.splice(idx, 1)
  items.value.splice(target, 0, row)
}

function goTo(n: number) {
  if (n < 1 || n > steps.length || savedView.value) return
  step.value = n
  if (n === 5 || n === 7) void refreshTerbilang()
}

function stepBadgeClass(n: number): string {
  if (n === step.value) return 'border-brand-600 bg-brand-600 text-white'
  if (n < step.value) return 'border-brand-200 bg-brand-50 text-brand-700'
  return 'border-slate-200 text-slate-400 bg-white'
}

async function save() {
  saving.value = true
  saveError.value = ''
  try {
    const payload = {
      header: { ...header, vesselId: vesselChoice.value > 0 ? vesselChoice.value : undefined },
      items: items.value.map(({ uid: _uid, ...rest }) => ({ ...rest }))
    }
    if (isEdit.value) {
      await sphStore.updateDraft(editId.value, payload as never)
      await router.push(`/sph/${editId.value}`)
      return
    }
    savedView.value = await sphStore.create(payload as never)
  } catch (e) {
    saveError.value = errorMessage(e)
  } finally {
    saving.value = false
  }
}

// ===== muat awal =====
onMounted(async () => {
  try {
    await Promise.all([
      partnerStore.load(),
      masterStore.loadCategories(),
      masterStore.loadWorkItems(),
      templateStore.loadTemplates(),
      sphStore.loadList().then(() => {
        oldDocs.value = sphStore.list
      })
    ])
    if (isEdit.value) {
      const doc = await sphStore.getDetail(editId.value)
      Object.assign(header, emptyHeader(), {
        date: String(doc.date ?? '').slice(0, 10),
        customerId: doc.customerId,
        vesselId: doc.vesselId ?? undefined,
        projectName: doc.projectName,
        subject: doc.subject,
        reference: doc.reference,
        location: doc.location,
        validUntil: doc.validUntil ? String(doc.validUntil).slice(0, 10) : '',
        picName: doc.picName,
        notes: doc.notes
      })
      vesselChoice.value = doc.vesselId ?? 0
      items.value = (doc.items ?? []).map((row) => ({
        uid: nextUid(),
        workItemId: row.workItemId ?? undefined,
        name: row.nameSnapshot,
        description: row.descriptionSnapshot,
        quantity: row.quantity,
        unit: row.unit,
        serviceUnitPrice: row.serviceUnitPrice,
        materialUnitPrice: row.materialUnitPrice,
        pricingMode: row.pricingMode || 'HARGA_LANGSUNG',
        notes: row.notes,
        subItems: (row.subItems ?? []).map((s): SphSubItemInput => ({
          name: s.nameSnapshot,
          description: s.descriptionSnapshot,
          quantity: s.quantity,
          unit: s.unit,
          serviceUnitPrice: s.serviceUnitPrice,
          materialUnitPrice: s.materialUnitPrice,
          weight: s.weight ?? 0,
          notes: s.notes
        }))
      }))
    }
  } catch (e) {
    pageError.value = errorMessage(e)
  }
})
</script>

<style scoped>
.row-input {
  width: 100%;
  border-radius: 0.5rem;
  border: 1px solid #e2e8f0;
  padding: 0.45rem 0.6rem;
  font-size: 13px;
  outline: none;
}
.row-input:focus {
  border-color: #f59e0b66;
}
</style>
