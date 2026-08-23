<template>
  <div>
    <PageHeader :title="(route.meta.title as string) ?? ''" />
    <EmptyState
      :title="'Halaman ' + route.meta.title + ' belum tersedia'"
      :description="'Fitur ini dikerjakan pada ' + route.meta.phase + ' sesuai rencana pengembangan.'"
    />
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import EmptyState from '../components/EmptyState.vue'

const route = useRoute()
</script>
