<script setup lang="ts">
defineProps<{ pretitle?: string; title: string }>()
</script>

<template>
  <div class="page-header d-print-none">
    <div class="container-xl">
      <div class="row g-2 align-items-center">
        <div class="col-12 col-md min-w-0">
          <div v-if="pretitle" class="page-pretitle">{{ pretitle }}</div>
          <h2 class="page-title text-break-all">{{ title }}</h2>
          <slot name="meta" />
        </div>
        <div v-if="$slots.actions" class="col-12 col-md-auto ms-md-auto">
          <slot name="actions" />
        </div>
      </div>
    </div>
  </div>
</template>
