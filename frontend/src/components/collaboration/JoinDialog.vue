<template>
  <AppModal :model-value="modelValue" title="Gabung Room Kolaborasi" @update:model-value="$emit('update:modelValue', $event)">
    <div v-if="pageError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700">{{ pageError }}</div>

    <div v-if="room" class="mb-4 rounded-lg border border-slate-200 bg-slate-50 px-4 py-3 text-[13px]">
      <p class="text-slate-500">Room ditemukan:</p>
      <p class="font-semibold text-slate-800">{{ room.roomName }}</p>
      <p class="text-xs text-slate-400">{{ room.documentNumber }} · Host: {{ room.hostName }} · {{ room.users }} pengguna</p>
    </div>

    <div class="space-y-4">
      <div>
        <label class="mb-1 block text-[13px] font-medium text-slate-600">Nama Tampilan <span class="text-red-500">*</span></label>
        <input v-model="displayName" type="text" maxlength="100" placeholder="nama Anda" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
      </div>
      <div>
        <label class="mb-1 block text-[13px] font-medium text-slate-600">Device Name</label>
        <input v-model="deviceName" type="text" maxlength="100" placeholder="nama device ini" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
      </div>
      <div>
        <label class="mb-1 block text-[13px] font-medium text-slate-600">Access Code</label>
        <input v-model="accessCode" type="text" maxlength="6" placeholder="6 digit angka" class="w-full rounded-lg border border-slate-200 px-3 py-2 text-[13px] font-mono outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100" />
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-2">
        <button type="button" class="rounded-lg border border-slate-200 px-3.5 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50" @click="$emit('update:modelValue', false)">
          Batal
        </button>
        <button type="button" class="rounded-lg bg-brand-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-60" :disabled="busy || !canJoin" @click="doJoin">
          {{ busy ? 'Menyambung…' : 'Gabung' }}
        </button>
      </div>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import AppModal from '../AppModal.vue'
import { useCollaborationStore } from '../../stores/collaboration'
import type { DiscoveredRoom } from '../../types/collaboration'

const props = defineProps<{ modelValue: boolean; room: DiscoveredRoom | null }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: boolean): void; (e: 'joined'): void }>()

const collabStore = useCollaborationStore()

const displayName = ref('')
const deviceName = ref('')
const accessCode = ref('')
const busy = ref(false)
const pageError = ref('')

const canJoin = computed(() => displayName.value.trim() && props.room)

async function doJoin() {
  if (!props.room) return
  busy.value = true
  pageError.value = ''
  try {
    await collabStore.joinRoom(
      props.room.hostIP,
      props.room.port,
      accessCode.value.trim(),
      '',
      displayName.value.trim()
    )
    emit('update:modelValue', false)
    emit('joined')
  } catch (e) {
    pageError.value = String(e)
  } finally {
    busy.value = false
  }
}

watch(
  () => props.modelValue,
  (open) => {
    if (!open) return
    displayName.value = collabStore.defaults?.displayName ?? ''
    deviceName.value = collabStore.defaults?.deviceName ?? ''
    accessCode.value = ''
    pageError.value = ''
  }
)
</script>
