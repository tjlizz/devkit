<script setup lang="ts">
import { ref } from 'vue'

import apiClient from '../api/client'

const health = ref<'idle' | 'checking' | 'ok' | 'error'>('idle')

async function checkHealth() {
  health.value = 'checking'
  try {
    const response = await apiClient.get<{ status: string }>('/health')
    health.value = response.data.status === 'ok' ? 'ok' : 'error'
  } catch {
    health.value = 'error'
  }
}
</script>

<template>
  <section class="hero">
    <a-tag color="blue">Vue 3 + Go + SQLite</a-tag>
    <h1>DevKit</h1>
    <p>
      A modular foundation for building and managing developer applications, with a
      typed frontend and a lightweight Go API.
    </p>
    <a-space>
      <RouterLink to="/register">
        <a-button type="primary" size="large">Get started</a-button>
      </RouterLink>
      <a-button :loading="health === 'checking'" size="large" @click="checkHealth">
        Check API
      </a-button>
    </a-space>
    <a-alert
      v-if="health === 'ok'"
      message="API is healthy"
      type="success"
      show-icon
      style="max-width: 360px; margin: 2rem auto 0"
    />
    <a-alert
      v-else-if="health === 'error'"
      message="API is unavailable"
      description="Start the Go server and try again."
      type="error"
      show-icon
      style="max-width: 360px; margin: 2rem auto 0"
    />
  </section>
</template>
