<template>
  <div>
    <PageHeader title="Master Pekerjaan" subtitle="Database pekerjaan yang dapat dipakai ulang saat menyusun SPH">
      <template #actions>
        <button
          type="button"
          class="flex items-center gap-1.5 rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700"
          @click="openCreateWi"
        >
          <span class="text-base leading-none">+</span> Pekerjaan
        </button>
      </template>
    </PageHeader>

    <p v-if="store.workItemsError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">
      {{ store.workItemsError }}
    </p>
    <p v-if="actionError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">
      {{ actionError }}
    </p>

    <div class="flex items-start gap-5">
      <!-- Filter kategori -->
      <aside class="w-60 shrink-0 rounded-xl border border-slate-200 bg-white">
        <div class="border-b border-slate-100 px-4 py-3">
          <h2 class="text-xs font-semibold uppercase tracking-wide text-slate-400">Kategori</h2>
        </div>
        <div class="p-2">
          <button
            type="button"
            class="flex w-full items-center justify-between rounded-lg px-3 py-2 text-left text-[13px] transition-colors"
            :class="store.wiCategoryId === 0 ? 'bg-brand-600 text-white' : 'text-slate-600 hover:bg-slate-100'"
            @click="selectCategory(0)"
          >
            Semua Kategori
          </button>
          <button
            v-for="cat in store.categories"
            :key="cat.id"
            type="button"
            class="flex w-full items-center justify-between gap-2 rounded-lg px-3 py-2 text-left text-[13px] transition-colors"
            :class="store.wiCategoryId === cat.id ? 'bg-brand-600 text-white' : 'text-slate-600 hover:bg-slate-100'"
            :title="cat.isActive ? '' : '(nonaktif)'"
            @click="selectCategory(cat.id)"
          >
            <span class="truncate">{{ cat.name }}</span>
            <span
              class="shrink-0 rounded-full px-1.5 py-0.5 text-[11px] font-medium"
              :class="store.wiCategoryId === cat.id ? 'bg-white/20 text-white' : 'bg-slate-100 text-slate-500'"
            >
              {{ cat.workItemCount }}
            </span>
          </button>
          <p v-if="!store.categories.length && !store.categoriesLoading" class="px-3 py-2 text-xs text-slate-400">
            Belum ada kategori.
            <RouterLink to="/pekerjaan/kategori" class="font-medium text-brand-600 hover:underline">Buat kategori</RouterLink>
          </p>
        </div>
        <div class="border-t border-slate-100 px-4 py-2.5">
          <RouterLink to="/pekerjaan/kategori" class="text-xs font-medium text-brand-600 hover:underline">Kelola Kategori →</RouterLink>
        </div>
      </aside>

      <!-- Daftar pekerjaan -->
      <section class="min-w-0 flex-1">
        <div class="mb-3 flex flex-wrap items-center gap-3">
          <div class="relative w-72">
            <svg class="pointer-events-none absolute left-2.5 top-2.5 h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke-width="1.8" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
            </svg>
            <input
              v-model="store.wiSearch"
              type="search"
              placeholder="Cari kode atau nama pekerjaan…"
              class="w-full rounded-lg border border-slate-200 py-2 pl-8 pr-3 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
            />
          </div>
          <label class="flex cursor-pointer items-center gap-2 text-[13px] text-slate-600">
            <input v-model="store.wiIncludeInactive" type="checkbox" class="h-4 w-4 rounded border-slate-300 accent-brand-600" />
            Tampilkan yang nonaktif
          </label>
          <span v-if="canDragWi" class="ml-auto text-xs text-slate-400">Seret untuk mengatur urutan pekerjaan</span>
        </div>

        <!-- Bar hapus massal pekerjaan -->
        <div v-if="selectedWi.size" class="mb-3 flex flex-wrap items-center gap-x-3 gap-y-2 rounded-xl border border-red-200 bg-red-50 px-4 py-2.5 text-[13px]">
          <label class="flex cursor-pointer items-center gap-2 font-medium text-slate-700">
            <input
              type="checkbox"
              class="h-4 w-4 rounded border-slate-300 accent-brand-600"
              :checked="allWiSelected"
              @change="toggleAllWi($event)"
            />
            Pilih semua
          </label>
          <span class="text-slate-600">{{ selectedWi.size }} dipilih</span>
          <div class="ml-auto flex items-center gap-2">
            <button
              type="button"
              class="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-medium text-slate-600 transition-colors hover:bg-slate-50"
              @click="selectedWi = new Set()"
            >
              Bersihkan
            </button>
            <button
              type="button"
              class="rounded-lg bg-red-600 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-red-700 disabled:opacity-60"
              :disabled="deleting"
              @click="askBulkDeleteWi"
            >
              Hapus Terpilih
            </button>
          </div>
        </div>

        <div v-if="store.workItemsLoading" class="rounded-xl border border-slate-200 bg-white px-4 py-6 text-center text-[13px] text-slate-400">Memuat…</div>

        <div v-else-if="!store.workItems.length" class="rounded-xl border border-slate-200 bg-white p-6">
          <EmptyState title="Belum ada pekerjaan" description="Tambahkan pekerjaan ke dalam kategori untuk mulai membangun database pekerjaan.">
            <button
              type="button"
              class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700"
              @click="openCreateWi"
            >
              + Pekerjaan Pertama
            </button>
          </EmptyState>
        </div>

        <div v-else class="space-y-3">
          <article
            v-for="(wi, idx) in store.workItems"
            :key="wi.id"
            class="rounded-xl border bg-white transition-shadow"
            :class="[wi.isActive ? 'border-slate-200' : 'border-slate-200 opacity-55', draggingWi === wi.id ? 'shadow-lg ring-2 ring-brand-200' : '']"
          >
            <header class="flex items-center gap-3 px-4 py-3">
              <input
                type="checkbox"
                class="h-4 w-4 shrink-0 rounded border-slate-300 accent-brand-600"
                :checked="selectedWi.has(wi.id)"
                @change="toggleSelectWi(wi.id, $event)"
                @click.stop
              />
              <span
                v-if="canDragWi"
                class="cursor-grab select-none text-slate-300"
                draggable="true"
                title="Seret untuk mengurutkan"
                @dragstart="startDragWi(wi.id)"
                @dragenter.prevent="enterDragWi(wi.id)"
                @dragover.prevent
                @dragend="endDragWi"
              >&#8942;&#8942;</span>
              <span v-else class="w-4 text-center text-xs text-slate-300">{{ wi.sequence || idx + 1 }}</span>

              <button type="button" class="flex min-w-0 flex-1 items-center gap-2.5 text-left" @click="toggleExpand(wi)">
                <svg
                  class="h-4 w-4 shrink-0 text-slate-400 transition-transform"
                  :class="expandedId === wi.id ? 'rotate-90' : ''"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke-width="2"
                  stroke="currentColor"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
                </svg>
                <span v-if="wi.code" class="shrink-0 rounded-md bg-slate-100 px-1.5 py-0.5 font-mono text-xs text-slate-600">{{ wi.code }}</span>
                <span class="truncate text-sm font-medium text-slate-800">{{ wi.name }}</span>
                <span class="shrink-0 text-xs text-slate-400">{{ wi.subItemCount }} sub</span>
                <span v-if="!wi.isActive" class="shrink-0 rounded-full bg-slate-100 px-2 py-0.5 text-xs text-slate-500">Nonaktif</span>
              </button>

              <div class="hidden shrink-0 text-right md:block">
                <p class="text-[13px] font-medium tabular-nums text-slate-700">{{ formatRupiah(wi.defaultServicePrice) }}</p>
                <p class="text-xs tabular-nums text-slate-400">+ {{ formatRupiah(wi.defaultMaterialPrice) }} material</p>
              </div>

              <div class="shrink-0 whitespace-nowrap">
                <button class="rounded-md px-2 py-1 text-xs font-medium text-brand-600 transition-colors hover:bg-brand-50" @click="openEditWi(wi)">Edit</button>
                <button class="rounded-md px-2 py-1 text-xs font-medium text-slate-500 transition-colors hover:bg-slate-100" @click="toggleWiActive(wi)">
                  {{ wi.isActive ? 'Nonaktifkan' : 'Aktifkan' }}
                </button>
                <button class="rounded-md px-2 py-1 text-xs font-medium text-red-600 transition-colors hover:bg-red-50" @click="askDeleteWi(wi)">Hapus</button>
              </div>
            </header>

            <!-- Sub-pekerjaan -->
            <div v-if="expandedId === wi.id" class="border-t border-slate-100 bg-slate-50/50 px-4 py-3">
              <p v-if="detailLoading" class="py-2 text-[13px] text-slate-400">Memuat sub-pekerjaan…</p>

              <table v-else-if="subRows.length" class="w-full text-left">
                <thead>
                  <tr class="border-b border-slate-200 text-xs uppercase tracking-wide text-slate-400">
                    <th class="w-8 py-1.5 pl-2 font-medium">
                      <input
                        type="checkbox"
                        class="h-3.5 w-3.5 rounded border-slate-300 accent-brand-600"
                        :checked="allSubsSelected"
                        @change="toggleAllSubs($event)"
                      />
                    </th>
                    <th class="w-10 py-1.5 pl-2 font-medium"></th>
                    <th class="px-2 py-1.5 font-medium">#</th>
                    <th class="px-2 py-1.5 font-medium">Nama</th>
                    <th class="px-2 py-1.5 font-medium">Bobot</th>
                    <th class="px-2 py-1.5 font-medium">Qty × Satuan</th>
                    <th class="px-2 py-1.5 text-right font-medium">Harga Jasa</th>
                    <th class="px-2 py-1.5 text-right font-medium">Material</th>
                    <th class="px-2 py-1.5 text-right font-medium">Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="(sub, sIdx) in subRows"
                    :key="sub.id"
                    class="border-b border-slate-100 text-[13px] last:border-b-0 hover:bg-white"
                    :class="{ 'opacity-55': !sub.isActive, 'cursor-grab': canDragSubs, 'bg-brand-50/40': draggingSub === sub.id }"
                    :draggable="canDragSubs"
                    @dragstart="startDragSub(sub.id)"
                    @dragenter.prevent="enterDragSub(sub.id)"
                    @dragover.prevent
                    @dragend="endDragSub"
                  >
                    <td class="py-2 pl-2">
                      <input
                        type="checkbox"
                        class="h-3.5 w-3.5 rounded border-slate-300 accent-brand-600"
                        :checked="selectedSubs.has(sub.id)"
                        @change="toggleSelectSub(sub.id, $event)"
                      />
                    </td>
                    <td class="py-2 pl-2 text-slate-300">
                      <span v-if="canDragSubs" class="select-none">&#8942;&#8942;</span>
                      <span v-else>{{ sub.sequence || sIdx + 1 }}</span>
                    </td>
                    <td class="px-2 py-2 tabular-nums text-slate-400">{{ sIdx + 1 }}</td>
                    <td class="px-2 py-2">
                      <p class="font-medium text-slate-700">{{ sub.name }}</p>
                      <p v-if="sub.code" class="font-mono text-[11px] text-slate-400">{{ sub.code }}</p>
                    </td>
                    <td class="px-2 py-2 tabular-nums text-slate-500">{{ sub.difficultyWeight }}%</td>
                    <td class="px-2 py-2 text-slate-600">{{ formatQty(sub.defaultQuantity) }} {{ sub.defaultUnit || '—' }}</td>
                    <td class="px-2 py-2 text-right tabular-nums text-slate-600">{{ formatRupiah(sub.defaultServicePrice) }}</td>
                    <td class="px-2 py-2 text-right tabular-nums text-slate-600">{{ formatRupiah(sub.defaultMaterialPrice) }}</td>
                    <td class="whitespace-nowrap px-2 py-2 text-right">
                      <button class="rounded-md px-1.5 py-1 text-xs font-medium text-brand-600 hover:bg-brand-50" @click="openEditSub(sub)">Edit</button>
                      <button class="rounded-md px-1.5 py-1 text-xs font-medium text-slate-500 hover:bg-slate-100" @click="toggleSubActive(sub)">
                        {{ sub.isActive ? 'Nonaktifkan' : 'Aktifkan' }}
                      </button>
                      <button class="rounded-md px-1.5 py-1 text-xs font-medium text-red-600 hover:bg-red-50" @click="askDeleteSub(sub)">Hapus</button>
                    </td>
                  </tr>
                </tbody>
              </table>

              <p v-else class="py-2 text-[13px] text-slate-400">Belum ada sub-pekerjaan.</p>

              <!-- Bar hapus massal sub-pekerjaan -->
              <div v-if="selectedSubs.size" class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px]">
                <span class="text-slate-600">{{ selectedSubs.size }} dipilih</span>
                <div class="ml-auto flex items-center gap-2">
                  <button
                    type="button"
                    class="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-medium text-slate-600 transition-colors hover:bg-slate-50"
                    @click="selectedSubs = new Set()"
                  >
                    Bersihkan
                  </button>
                  <button
                    type="button"
                    class="rounded-lg bg-red-600 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-red-700 disabled:opacity-60"
                    :disabled="deleting"
                    @click="askBulkDeleteSubs"
                  >
                    Hapus Terpilih
                  </button>
                </div>
              </div>

              <button
                type="button"
                class="mt-2 flex items-center gap-1 rounded-lg border border-dashed border-slate-300 px-3 py-1.5 text-xs font-medium text-slate-500 transition-colors hover:border-brand-400 hover:text-brand-600"
                @click="openCreateSub"
              >
                + Sub-Pekerjaan
              </button>
            </div>
          </article>
        </div>
      </section>
    </div>

    <!-- Modal form pekerjaan -->
    <AppModal v-model="wiFormOpen" :title="editingWi ? 'Edit Pekerjaan' : 'Tambah Pekerjaan'" size="lg">
      <form class="grid grid-cols-2 gap-x-4 gap-y-3.5" @submit.prevent="submitWiForm" @keydown.ctrl.enter.prevent="submitWiForm">
        <p v-if="wiSavedNote" class="col-span-2 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-[13px] text-emerald-700">{{ wiSavedNote }}</p>
        <div class="col-span-2">
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Kategori <span class="text-red-500">*</span></label>
          <select
            v-model.number="wiForm.categoryId"
            class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-[13px] outline-none transition-colors focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          >
            <option :value="0" disabled>Pilih kategori…</option>
            <option v-for="cat in store.categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
          </select>
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Kode</label>
          <input v-model="wiForm.code" type="text" disabled placeholder="(otomatis)" class="w-full cursor-not-allowed rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-[13px] text-slate-400 outline-none" />
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Nama <span class="text-red-500">*</span></label>
          <input v-model="wiForm.name" type="text" required maxlength="300" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
        </div>
        <div class="col-span-2">
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Deskripsi</label>
          <textarea v-model="wiForm.description" rows="2" maxlength="1000" class="w-full resize-none rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100"></textarea>
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Satuan Default</label>
          <input v-model="wiForm.defaultUnit" type="text" maxlength="30" placeholder="giat / lot / unit" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Qty Default <span class="text-red-500">*</span></label>
          <input v-model.number="wiForm.defaultQuantity" type="number" min="0.01" step="any" required class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Harga Jasa Default (Rp)</label>
          <input v-model.number="wiForm.defaultServicePrice" type="number" min="0" step="1" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
        </div>
        <div>
          <div class="mb-1 flex items-center justify-between gap-2">
            <label class="text-[13px] font-medium text-slate-600">Harga Material Default (Rp)</label>
            <button type="button" class="shrink-0 rounded-md px-1.5 py-0.5 text-xs font-medium text-brand-600 transition-colors hover:bg-brand-50" @click="openPick">⌕ Pilih Material</button>
          </div>
          <input v-model.number="wiForm.defaultMaterialPrice" type="number" min="0" step="1" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
          <p v-if="pickedWiMaterial" class="mt-1 truncate text-xs text-slate-400">
            Dari: <span class="font-mono">{{ pickedWiMaterial.code }}</span> {{ pickedWiMaterial.name }} · {{ formatRupiah(pickedWiMaterial.defaultPrice) }}
          </p>
        </div>
        <div class="col-span-2">
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Catatan</label>
          <textarea v-model="wiForm.notes" rows="2" maxlength="1000" class="w-full resize-none rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100"></textarea>
        </div>
        <p v-if="wiFormError" class="col-span-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px] text-red-700">{{ wiFormError }}</p>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 hover:bg-slate-50" :disabled="saving" @click="wiFormOpen = false">Batal</button>
          <button type="button" class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white hover:bg-brand-700 disabled:opacity-60" :disabled="saving" @click="submitWiForm">
            {{ saving ? 'Menyimpan…' : 'Simpan' }}
          </button>
        </div>
      </template>
    </AppModal>

    <!-- Modal form sub-pekerjaan -->
    <AppModal v-model="subFormOpen" :title="editingSub ? 'Edit Sub-Pekerjaan' : 'Tambah Sub-Pekerjaan'">
      <form class="space-y-3.5" @submit.prevent="submitSubForm" @keydown.ctrl.enter.prevent="submitSubForm">
        <p v-if="subSavedNote" class="rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-[13px] text-emerald-700">{{ subSavedNote }}</p>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Nama <span class="text-red-500">*</span></label>
          <input v-model="subForm.name" type="text" required maxlength="300" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
        </div>
        <div class="grid grid-cols-2 gap-x-4">
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Kode</label>
            <input v-model="subForm.code" type="text" disabled placeholder="(otomatis)" class="w-full cursor-not-allowed rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-[13px] text-slate-400 outline-none" />
          </div>
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Bobot Kesulitan (%)</label>
            <input v-model.number="subForm.difficultyWeight" type="number" min="0" max="100" step="1" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
          </div>
        </div>
        <div class="grid grid-cols-2 gap-x-4">
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Satuan Default</label>
            <input v-model="subForm.defaultUnit" type="text" maxlength="30" placeholder="giat / lot / unit" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
          </div>
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Qty Default <span class="text-red-500">*</span></label>
            <input v-model.number="subForm.defaultQuantity" type="number" min="0.01" step="any" required class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
          </div>
        </div>
        <div class="grid grid-cols-2 gap-x-4">
          <div>
            <label class="mb-1 block text-[13px] font-medium text-slate-600">Harga Jasa Default (Rp)</label>
            <input v-model.number="subForm.defaultServicePrice" type="number" min="0" step="1" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
          </div>
          <div>
            <div class="mb-1 flex items-center justify-between gap-2">
              <label class="text-[13px] font-medium text-slate-600">Harga Material Default (Rp)</label>
              <button type="button" class="shrink-0 rounded-md px-1.5 py-0.5 text-xs font-medium text-brand-600 transition-colors hover:bg-brand-50" @click="openPick">⌕ Pilih Material</button>
            </div>
            <input v-model.number="subForm.defaultMaterialPrice" type="number" min="0" step="1" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
            <p v-if="pickedSubMaterial" class="mt-1 truncate text-xs text-slate-400">
              Dari: <span class="font-mono">{{ pickedSubMaterial.code }}</span> {{ pickedSubMaterial.name }} · {{ formatRupiah(pickedSubMaterial.defaultPrice) }}
            </p>
          </div>
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Catatan</label>
          <textarea v-model="subForm.notes" rows="2" maxlength="1000" class="w-full resize-none rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100"></textarea>
        </div>
        <p v-if="subFormError" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-[13px] text-red-700">{{ subFormError }}</p>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 hover:bg-slate-50" :disabled="saving" @click="subFormOpen = false">Batal</button>
          <button type="button" class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white hover:bg-brand-700 disabled:opacity-60" :disabled="saving" @click="submitSubForm">
            {{ saving ? 'Menyimpan…' : 'Simpan' }}
          </button>
        </div>
      </template>
    </AppModal>

    <!-- Pemilih material master -->
    <MaterialPickerModal v-model="pickOpen" @select="applyPick" />

    <!-- Konfirmasi hapus pekerjaan -->
    <ConfirmDialog
      v-model="deleteWiOpen"
      title="Hapus Pekerjaan"
      :message="`Hapus pekerjaan '${deleteWiTarget?.name ?? ''}'?`"
      detail="Seluruh sub-pekerjaannya ikut terhapus. Data tidak akan tampil lagi namun tetap tersimpan."
      confirm-label="Hapus"
      :busy="deleting"
      @confirm="confirmDeleteWi"
    />

    <!-- Konfirmasi hapus massal pekerjaan -->
    <ConfirmDialog
      v-model="bulkDeleteWiOpen"
      title="Hapus Pekerjaan Terpilih"
      :message="`Hapus ${selectedWi.size} pekerjaan beserta seluruh sub-pekerjaannya?`"
      detail="Semua sub-pekerjaan di bawahnya ikut terhapus. Data tetap tersimpan dan tidak dapat dikembalikan dari aplikasi."
      confirm-label="Hapus Semua"
      :busy="deleting"
      @confirm="confirmBulkDeleteWi"
    />

    <!-- Konfirmasi hapus sub-pekerjaan -->
    <ConfirmDialog
      v-model="deleteSubOpen"
      title="Hapus Sub-Pekerjaan"
      :message="`Hapus sub-pekerjaan '${deleteSubTarget?.name ?? ''}'?`"
      detail="Data tidak akan tampil lagi namun tetap tersimpan."
      confirm-label="Hapus"
      :busy="deleting"
      @confirm="confirmDeleteSub"
    />

    <!-- Konfirmasi hapus massal sub-pekerjaan -->
    <ConfirmDialog
      v-model="bulkDeleteSubOpen"
      title="Hapus Sub-Pekerjaan Terpilih"
      :message="`Hapus ${selectedSubs.size} sub-pekerjaan?`"
      detail="Data tidak akan tampil lagi namun tetap tersimpan."
      confirm-label="Hapus Semua"
      :busy="deleting"
      @confirm="confirmBulkDeleteSubs"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import EmptyState from '../components/EmptyState.vue'
import AppModal from '../components/AppModal.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import MaterialPickerModal from '../components/MaterialPickerModal.vue'
import { useMasterStore } from '../stores/master'
import { errorMessage, formatRupiah, formatQty } from '../utils/format'
import { emptySubItem, emptyWorkItem, type MaterialView, type WorkItemDetail, type WorkItemView, type WorkSubItem } from '../types/master'
import { useDragSort } from '../composables/useDragSort'

const store = useMasterStore()

const actionError = ref('')
const saving = ref(false)
const deleting = ref(false)

// ===== ekspansi & detail =====
const expandedId = ref<number | null>(null)
const details = ref<Record<number, WorkItemDetail>>({})
const detailLoading = ref(false)

async function toggleExpand(wi: WorkItemView) {
  if (expandedId.value === wi.id) {
    expandedId.value = null
    return
  }
  expandedId.value = wi.id
  if (!details.value[wi.id]) {
    await loadDetail(wi.id)
  }
}

async function loadDetail(id: number) {
  detailLoading.value = true
  try {
    details.value[id] = await store.getWorkItemDetail(id)
  } catch (e) {
    actionError.value = errorMessage(e)
  } finally {
    detailLoading.value = false
  }
}

function applyDetail(d: WorkItemDetail) {
  if (d.subItems?.length) details.value[d.subItems[0].workItemId] = d
}

// ===== filter =====
function selectCategory(id: number) {
  store.wiCategoryId = id
  expandedId.value = null
}

const canDragWi = computed(
  () => store.wiCategoryId !== 0 && !store.wiSearch && store.workItems.length > 1
)

let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(
  () => store.wiSearch,
  () => {
    if (searchTimer) clearTimeout(searchTimer)
    searchTimer = setTimeout(() => store.loadWorkItems(), 250)
  }
)
watch(
  () => store.wiIncludeInactive,
  () => store.loadWorkItems()
)

// ===== drag urutan pekerjaan =====
const { draggingId: draggingWi, startDrag: startDragWi, enterDrag: enterDragWi, endDrag: endDragWi } =
  useDragSort(computed(() => store.workItems), async (ids) => {
    try {
      await store.reorderWorkItems(ids)
    } catch (e) {
      actionError.value = errorMessage(e)
      await store.loadWorkItems()
    }
  })

// ===== drag urutan sub-pekerjaan =====
const subRows = computed<WorkSubItem[]>(() =>
  expandedId.value !== null ? details.value[expandedId.value]?.subItems ?? [] : []
)
const canDragSubs = computed(() => expandedId.value !== null && subRows.value.length > 1)

const { draggingId: draggingSub, startDrag: startDragSub, enterDrag: enterDragSub, endDrag: endDragSub } =
  useDragSort(subRows, async (ids) => {
    if (expandedId.value === null) return
    try {
      await store.reorderSubItems(expandedId.value, ids)
      // perbarui sequence lokal
      subRows.value.forEach((s, i) => (s.sequence = i + 1))
    } catch (e) {
      actionError.value = errorMessage(e)
      await loadDetail(expandedId.value)
    }
  })

// ===== seleksi massal =====
const selectedWi = ref(new Set<number>())
const selectedSubs = ref(new Set<number>())

function toggleSelectWi(id: number, ev: Event) {
  const on = (ev.target as HTMLInputElement).checked
  const next = new Set(selectedWi.value)
  if (on) next.add(id)
  else next.delete(id)
  selectedWi.value = next
}

const allWiSelected = computed(
  () => store.workItems.length > 0 && store.workItems.every((w) => selectedWi.value.has(w.id))
)

function toggleAllWi(ev: Event) {
  const on = (ev.target as HTMLInputElement).checked
  selectedWi.value = on ? new Set(store.workItems.map((w) => w.id)) : new Set()
}

// buang ID yang sudah tidak ada di daftar setiap kali data dimuat ulang
watch(
  () => store.workItems,
  (list) => {
    if (!selectedWi.value.size) return
    const alive = new Set(list.map((w) => w.id))
    const next = new Set([...selectedWi.value].filter((id) => alive.has(id)))
    if (next.size !== selectedWi.value.size) selectedWi.value = next
  }
)

watch(expandedId, () => {
  selectedSubs.value = new Set()
})

const allSubsSelected = computed(
  () => subRows.value.length > 0 && subRows.value.every((s) => selectedSubs.value.has(s.id))
)

function toggleSelectSub(id: number, ev: Event) {
  const on = (ev.target as HTMLInputElement).checked
  const next = new Set(selectedSubs.value)
  if (on) next.add(id)
  else next.delete(id)
  selectedSubs.value = next
}

function toggleAllSubs(ev: Event) {
  const on = (ev.target as HTMLInputElement).checked
  selectedSubs.value = on ? new Set(subRows.value.map((s) => s.id)) : new Set()
}

watch(subRows, (rows) => {
  if (!selectedSubs.value.size) return
  const alive = new Set(rows.map((s) => s.id))
  const next = new Set([...selectedSubs.value].filter((id) => alive.has(id)))
  if (next.size !== selectedSubs.value.size) selectedSubs.value = next
})

// ===== form pekerjaan =====
const wiFormOpen = ref(false)
const editingWi = ref<WorkItemView | null>(null)
const wiForm = ref(emptyWorkItem())
const wiFormError = ref('')
const wiSavedNote = ref('')

let savedNoteTimer: ReturnType<typeof setTimeout> | null = null
function flashSaved(target: 'wi' | 'sub', text: string) {
  const note = target === 'wi' ? wiSavedNote : subSavedNote
  note.value = text
  if (savedNoteTimer) clearTimeout(savedNoteTimer)
  savedNoteTimer = setTimeout(() => (note.value = ''), 3000)
}

function openCreateWi() {
  editingWi.value = null
  wiForm.value = emptyWorkItem(store.wiCategoryId)
  pickedWiMaterial.value = null
  wiFormError.value = ''
  wiFormOpen.value = true
}

function openEditWi(wi: WorkItemView) {
  editingWi.value = wi
  wiForm.value = { ...emptyWorkItem(), ...wi }
  pickedWiMaterial.value = null
  wiFormError.value = ''
  wiFormOpen.value = true
}

async function submitWiForm() {
  wiFormError.value = ''
  if (!wiForm.value.name.trim()) {
    wiFormError.value = 'Nama pekerjaan wajib diisi.'
    return
  }
  if (!wiForm.value.categoryId) {
    wiFormError.value = 'Kategori wajib dipilih.'
    return
  }
  saving.value = true
  try {
    const payload = {
      ...wiForm.value,
      defaultQuantity: Number(wiForm.value.defaultQuantity) || 0,
      defaultServicePrice: Math.round(Number(wiForm.value.defaultServicePrice) || 0),
      defaultMaterialPrice: Math.round(Number(wiForm.value.defaultMaterialPrice) || 0)
    }
    let saved: WorkItemDetail
    if (editingWi.value) {
      saved = await store.updateWorkItem(editingWi.value.id, payload)
    } else {
      saved = await store.createWorkItem(payload)
    }
    applyDetail(saved)
    if (editingWi.value) {
      wiFormOpen.value = false
    } else {
      wiForm.value = emptyWorkItem(store.wiCategoryId)
      pickedWiMaterial.value = null
      wiFormError.value = ''
      flashSaved('wi', '✓ Data pekerjaan tersimpan.')
    }
    await Promise.all([store.loadWorkItems(), store.loadCategories()])
  } catch (e) {
    wiFormError.value = errorMessage(e)
  } finally {
    saving.value = false
  }
}

async function toggleWiActive(wi: WorkItemView) {
  actionError.value = ''
  try {
    await store.setWorkItemActive(wi.id, !wi.isActive)
  } catch (e) {
    actionError.value = errorMessage(e)
  }
}

// ===== hapus pekerjaan =====
const deleteWiOpen = ref(false)
const deleteWiTarget = ref<WorkItemView | null>(null)

function askDeleteWi(wi: WorkItemView) {
  actionError.value = ''
  deleteWiTarget.value = wi
  deleteWiOpen.value = true
}

async function confirmDeleteWi() {
  if (!deleteWiTarget.value) return
  deleting.value = true
  try {
    await store.deleteWorkItem(deleteWiTarget.value.id)
    if (expandedId.value === deleteWiTarget.value.id) expandedId.value = null
    deleteWiOpen.value = false
    await store.loadCategories()
  } catch (e) {
    actionError.value = errorMessage(e)
    deleteWiOpen.value = false
  } finally {
    deleting.value = false
  }
}

// ===== hapus massal pekerjaan =====
const bulkDeleteWiOpen = ref(false)

function askBulkDeleteWi() {
  if (!selectedWi.value.size) return
  actionError.value = ''
  bulkDeleteWiOpen.value = true
}

async function confirmBulkDeleteWi() {
  deleting.value = true
  try {
    const ids = [...selectedWi.value]
    await store.deleteWorkItems(ids)
    if (expandedId.value !== null && ids.includes(expandedId.value)) expandedId.value = null
    selectedWi.value = new Set()
    actionError.value = ''
    bulkDeleteWiOpen.value = false
    await Promise.all([store.loadWorkItems(), store.loadCategories()])
  } catch (e) {
    actionError.value = errorMessage(e)
    bulkDeleteWiOpen.value = false
  } finally {
    deleting.value = false
  }
}

// ===== form sub-pekerjaan =====
const subFormOpen = ref(false)
const editingSub = ref<WorkSubItem | null>(null)
const subForm = ref(emptySubItem())
const subFormError = ref('')
const subSavedNote = ref('')

function openCreateSub() {
  if (expandedId.value === null) return
  editingSub.value = null
  subForm.value = emptySubItem(expandedId.value)
  pickedSubMaterial.value = null
  subFormError.value = ''
  subFormOpen.value = true
}

function openEditSub(sub: WorkSubItem) {
  editingSub.value = sub
  subForm.value = { ...emptySubItem(), ...sub }
  pickedSubMaterial.value = null
  subFormError.value = ''
  subFormOpen.value = true
}

async function submitSubForm() {
  subFormError.value = ''
  if (!subForm.value.name.trim()) {
    subFormError.value = 'Nama sub-pekerjaan wajib diisi.'
    return
  }
  if (!subForm.value.workItemId) {
    subFormError.value = 'Pekerjaan induk tidak valid.'
    return
  }
  saving.value = true
  try {
    const payload = {
      ...subForm.value,
      defaultQuantity: Number(subForm.value.defaultQuantity) || 0,
      difficultyWeight: Math.round(Number(subForm.value.difficultyWeight) || 0),
      defaultServicePrice: Math.round(Number(subForm.value.defaultServicePrice) || 0),
      defaultMaterialPrice: Math.round(Number(subForm.value.defaultMaterialPrice) || 0)
    }
    let saved: WorkItemDetail
    if (editingSub.value) {
      saved = await store.updateSubItem(editingSub.value.id, payload)
    } else {
      saved = await store.createSubItem(payload)
    }
    applyDetail(saved)
    if (editingSub.value) {
      subFormOpen.value = false
    } else {
      subForm.value = emptySubItem(expandedId.value ?? 0)
      pickedSubMaterial.value = null
      subFormError.value = ''
      flashSaved('sub', '✓ Data sub-pekerjaan tersimpan.')
    }
    await store.loadWorkItems()
  } catch (e) {
    subFormError.value = errorMessage(e)
  } finally {
    saving.value = false
  }
}

async function toggleSubActive(sub: WorkSubItem) {
  actionError.value = ''
  try {
    await store.setSubItemActive(sub.id, !sub.isActive)
    await loadDetail(sub.workItemId)
  } catch (e) {
    actionError.value = errorMessage(e)
  }
}

// ===== pemilih material master =====
const pickOpen = ref(false)
const pickedWiMaterial = ref<MaterialView | null>(null)
const pickedSubMaterial = ref<MaterialView | null>(null)

function openPick() {
  pickOpen.value = true
}

// Isi harga material default dari master material; satuan ikut jika masih kosong.
// Untuk sub-pekerjaan: nama, satuan, dan harga langsung ditimpa dari material terpilih.
function applyPick(m: MaterialView) {
  if (wiFormOpen.value) {
    wiForm.value.defaultMaterialPrice = Math.round(m.defaultPrice)
    if (!wiForm.value.defaultUnit.trim()) wiForm.value.defaultUnit = m.unit || ''
    pickedWiMaterial.value = m
  } else if (subFormOpen.value) {
    subForm.value.name = m.name
    subForm.value.defaultUnit = m.unit || ''
    subForm.value.defaultMaterialPrice = Math.round(m.defaultPrice)
    pickedSubMaterial.value = m
  }
}

// ===== hapus sub-pekerjaan =====
const deleteSubOpen = ref(false)
const deleteSubTarget = ref<WorkSubItem | null>(null)

function askDeleteSub(sub: WorkSubItem) {
  actionError.value = ''
  deleteSubTarget.value = sub
  deleteSubOpen.value = true
}

async function confirmDeleteSub() {
  if (!deleteSubTarget.value) return
  deleting.value = true
  try {
    await store.deleteSubItem(deleteSubTarget.value.id)
    deleteSubOpen.value = false
    await loadDetail(deleteSubTarget.value.workItemId)
    await store.loadWorkItems()
  } catch (e) {
    actionError.value = errorMessage(e)
    deleteSubOpen.value = false
  } finally {
    deleting.value = false
  }
}

// ===== hapus massal sub-pekerjaan =====
const bulkDeleteSubOpen = ref(false)

function askBulkDeleteSubs() {
  if (!selectedSubs.value.size) return
  actionError.value = ''
  bulkDeleteSubOpen.value = true
}

async function confirmBulkDeleteSubs() {
  if (expandedId.value === null) return
  deleting.value = true
  try {
    await store.deleteSubItems([...selectedSubs.value])
    selectedSubs.value = new Set()
    actionError.value = ''
    bulkDeleteSubOpen.value = false
    await loadDetail(expandedId.value)
    await store.loadWorkItems()
  } catch (e) {
    actionError.value = errorMessage(e)
    bulkDeleteSubOpen.value = false
  } finally {
    deleting.value = false
  }
}

onMounted(() => {
  store.loadCategories()
  store.loadWorkItems()
})
</script>
