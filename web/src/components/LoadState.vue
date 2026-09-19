<script setup lang="ts">
defineProps<{ loading: boolean; error: string; empty: boolean; what: string }>()
defineEmits<{ retry: [] }>()
</script>

<template>
  <p v-if="loading" class="loading" role="status">Memuat {{ what }}...</p>
  <div v-else-if="error && empty" class="callout error" role="alert">
    <strong>Tidak bisa memuat {{ what }}.</strong>
    <p>{{ error }}. Pastikan socket Docker ter-mount ke container Netradock.</p>
    <button class="btn" type="button" @click="$emit('retry')">Coba lagi</button>
  </div>
  <div v-else-if="empty" class="empty"><slot name="empty" /></div>
  <template v-else>
    <p v-if="error" class="callout error stale" role="alert">Data mungkin sudah lama: {{ error }}</p>
    <slot />
  </template>
</template>

<style scoped>
.stale { margin-bottom: 16px; padding: 12px 16px; }
</style>
