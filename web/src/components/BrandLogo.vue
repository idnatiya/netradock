<script setup lang="ts">
import { IconBrandDocker } from '@tabler/icons-vue'
</script>

<template>
  <span class="brand-container d-inline-flex align-items-center gap-2">
    <span class="brand-icon"><IconBrandDocker :size="22" stroke-width="2" /></span>
    <span class="brand-title">Netradock</span>
  </span>
</template>

<style scoped>
.brand-container {
  user-select: none;
}
.brand-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
}
.brand-title {
  font-size: 15px;
  font-weight: 600;
  letter-spacing: -0.01em;
  color: #ffffff;
  line-height: 1;
}
</style>
