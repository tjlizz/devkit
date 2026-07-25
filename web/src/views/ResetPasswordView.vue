<script setup lang="ts">
import axios from 'axios'
import { computed, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'

import { resetPassword } from '../api/auth'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const token = computed(() => String(route.query.token || ''))
const form = reactive({
  newPassword: '',
})

async function submit() {
  if (!token.value) {
    message.error('Reset token is missing')
    return
  }
  loading.value = true
  try {
    await resetPassword(token.value, form.newPassword)
    message.success('Password reset successfully')
    await router.push('/login')
  } catch (error) {
    if (axios.isAxiosError(error)) {
      message.error(error.response?.data.error || error.message)
      return
    }
    message.error('Password reset failed')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <a-card class="page-card" title="Reset password">
    <a-alert
      v-if="!token"
      class="form-alert"
      type="error"
      show-icon
      message="Reset token is missing"
    />
    <a-form layout="vertical" :model="form" @finish="submit">
      <a-form-item
        label="New password"
        name="newPassword"
        :rules="[{ required: true, min: 8, message: 'Use at least 8 characters' }]"
      >
        <a-input-password v-model:value="form.newPassword" autocomplete="new-password" />
      </a-form-item>
      <a-button block type="primary" html-type="submit" :loading="loading" :disabled="!token">
        Reset password
      </a-button>
    </a-form>
  </a-card>
</template>

<style scoped>
.form-alert {
  margin-bottom: 16px;
}
</style>
