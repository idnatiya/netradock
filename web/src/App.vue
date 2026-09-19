<script setup lang="ts">
import { RouterView, useRoute } from 'vue-router'
import AppShell from '@/components/AppShell.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'

const route = useRoute()
</script>

<template>
  <RouterView v-if="route.meta.guest" />
  <AppShell v-else />
  <ConfirmDialog />
</template>
