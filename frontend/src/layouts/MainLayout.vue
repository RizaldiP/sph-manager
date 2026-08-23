<template>
  <div class="flex h-full">
    <SideNav />
    <div class="flex min-w-0 flex-1 flex-col">
      <TopBar />
      <main class="min-h-0 flex-1 overflow-y-auto p-6">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import SideNav from '../components/SideNav.vue'
import TopBar from '../components/TopBar.vue'
import { useAppStore } from '../stores/app'

const store = useAppStore()

onMounted(() => {
  if (!store.loaded) {
    store.load()
  }
})
</script>
