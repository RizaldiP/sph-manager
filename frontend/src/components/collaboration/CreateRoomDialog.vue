<template>
  <AppModal :model-value="modelValue" title="Mulai Room Kolaborasi" size="lg" @update:model-value="$emit('update:modelValue', $event)">
    <div v-if="pageError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">{{ pageError }}</div>

    <!-- Step 1: Pilih Draft SPH -->
    <section v-if="step === 1">
      <h3 class="mb-3 text-[13px] font-semibold text-slate-700">Pilih Dokumen SPH (Draft)</h3>
      <div v-if="sphLoading" class="px-4 py-8 text-center text-[13px] text-slate-400">Memuat daftar SPH…</div>
      <ul v-else class="max-h-[280px] divide-y divide-slate-100 overflow-y-auto rounded-lg border border-slate-200">
        <li v-for="d in drafts" :key="d.id" class="flex items-center gap-3 px-4 py-2.5 text-[13px]">
          <div class="min-w-0 flex-1">
            <p class="font-medium text-slate-700">{{ d.documentNumber }}</p>
            <p class="truncate text-xs text-slate-400">{{ d.customerName }} · {{ d.projectName }}</p>
          </div>
          <button type="button" class="shrink-0 rounded-md bg-brand-600 px-2.5 py-1.5 text-xs font-medium text-white transition-colors hover:bg-brand-700" @click="pickDoc(d)">
            Pilih
          </button>
        </li>
        <li v-if="!drafts.length" class="px-4 py-6 text-center text-[13px] italic text-slate-400">
          Tidak ada draft SPH. Buat draft terlebih dahulu.
        </li>
      </ul>
    </section>

    <!-- Step 2: Nama Room + Access Code -->
    <section v-if="step === 2">
      <div class="mb-4 rounded-lg border border-slate-200 bg-slate-50 px-4 py-3 text-[13px]">
        <p class="text-slate-500">Dokumen dipilih:</p>
        <p class="font-semibold text-slate-800">{{ pickedDoc?.documentNumber }} — {{ pickedDoc?.projectName }}</p>
      </div>

      <div class="space-y-4">
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Nama Room <span class="text-red-500">*</span></label>
          <input v-model="roomName" type="text" maxlength="100" placeholder="misal Review SPH bersama" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Nama Tampilan <span class="text-red-500">*</span></label>
          <input v-model="displayName" type="text" maxlength="100" placeholder="nama Anda" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
        </div>
        <div>
          <label class="mb-1 block text-[13px] font-medium text-slate-600">Access Code (opsional)</label>
          <input v-model="accessCode" type="text" maxlength="6" placeholder="6 digit angka (auto jika kosong)" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] font-mono outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
        </div>
      </div>
    </section>

    <template #footer>
      <div v-if="step === 1" class="flex justify-end">
        <button type="button" class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50" @click="$emit('update:modelValue', false)">
          Batal
        </button>
      </div>
      <div v-if="step === 2" class="flex justify-end gap-2">
        <button type="button" class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50" :disabled="busy" @click="step = 1">
          Kembali
        </button>
        <button type="button" class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-60" :disabled="busy || !canStart" @click="startRoom">
          {{ busy ? 'Memproses…' : 'Mulai Room' }}
        </button>
      </div>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import AppModal from '../AppModal.vue'
import { useSphStore } from '../../stores/sph'
import { useCollaborationStore } from '../../stores/collaboration'
import type { SphDocumentView } from '../../types/sph'

const props = defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: boolean): void; (e: 'started'): void }>()

const sphStore = useSphStore()
const collabStore = useCollaborationStore()

const step = ref(1)
const pickedDoc = ref<SphDocumentView | null>(null)
const roomName = ref('')
const displayName = ref('')
const accessCode = ref('')
const busy = ref(false)
const pageError = ref('')

const drafts = computed(() => sphStore.list.filter((d) => d.status === 'DRAFT'))
const sphLoading = computed(() => sphStore.listLoading)
const canStart = computed(() => roomName.value.trim() && displayName.value.trim())

function pickDoc(d: SphDocumentView) {
  pickedDoc.value = d
  roomName.value = d.projectName || d.documentNumber
  step.value = 2
}

async function startRoom() {
  if (!pickedDoc.value || !canStart.value) return
  busy.value = true
  pageError.value = ''
  try {
    await collabStore.createRoom(pickedDoc.value.id, roomName.value.trim(), displayName.value.trim())
    emit('update:modelValue', false)
    emit('started')
  } catch (e) {
    pageError.value = String(e)
  } finally {
    busy.value = false
  }
}

watch(
  () => props.modelValue,
  async (open) => {
    if (!open) return
    step.value = 1
    pickedDoc.value = null
    roomName.value = ''
    displayName.value = collabStore.defaults?.displayName ?? ''
    accessCode.value = ''
    pageError.value = ''
    await sphStore.loadList()
  }
)
</script>
