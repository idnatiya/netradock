<script setup lang="ts">
defineProps<{ loading: boolean; error: string; empty: boolean; what: string }>()
defineEmits<{ retry: [] }>()
</script>

<template>
  <div v-if="loading" class="d-flex align-items-center gap-2 text-secondary p-4" role="status">
    <span class="spinner-border spinner-border-sm" aria-hidden="true"></span>
    Memuat {{ what }}...
  </div>
  <div v-else-if="error && empty" class="alert alert-danger m-3" role="alert">
    <h4 class="alert-title">Tidak bisa memuat {{ what }}</h4>
    <div class="text-secondary text-break-all">{{ error }}. Pastikan socket Docker ter-mount ke container Netradock.</div>
    <button class="btn btn-danger mt-3" type="button" @click="$emit('retry')">Coba lagi</button>
  </div>
  <div v-else-if="empty" class="empty">
    <slot name="empty" />
  </div>
  <template v-else>
    <div v-if="error" class="alert alert-warning m-3" role="alert">Data mungkin sudah lama: {{ error }}</div>
    <slot />
  </template>
</template>
