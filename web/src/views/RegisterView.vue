<script setup lang="ts">
import { onMounted } from 'vue'

onMounted(() => {
  window.location.href = '/register'
})
</script>

<template>
  <a-card class="page-card" title="Create an account">
    <p class="text-zinc-500 dark:text-zinc-400">
      Registration is handled by the marketplace.
      <a href="/register" class="font-medium underline underline-offset-2">Go to registration</a>
    </p>
  </a-card>
</template>
