<script setup lang="ts">
import axios from 'axios'
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'

import { changePassword } from '../api/auth'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const loading = ref(false)
const form = reactive({
  oldPassword: '',
  newPassword: '',
})

async function submit() {
  if (!auth.isAuthenticated) {
    message.error('Please sign in first')
    await router.push('/login')
    return
  }
  loading.value = true
  try {
    await changePassword(form.oldPassword, form.newPassword)
    form.oldPassword = ''
    form.newPassword = ''
    message.success('Password changed successfully')
  } catch (error) {
    if (axios.isAxiosError(error)) {
      if (error.response?.status === 401) {
        message.error('Current password is incorrect')
        return
      }
      message.error(error.response?.data.error || error.message)
      return
    }
    message.error('Password change failed')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <a-card class="page-card" title="Change password">
    <a-form layout="vertical" :model="form" @finish="submit">
      <a-form-item
        label="Current password"
        name="oldPassword"
        :rules="[{ required: true, message: 'Enter your current password' }]"
      >
        <a-input-password v-model:value="form.oldPassword" autocomplete="current-password" />
      </a-form-item>
      <a-form-item
        label="New password"
        name="newPassword"
        :rules="[{ required: true, min: 8, message: 'Use at least 8 characters' }]"
      >
        <a-input-password v-model:value="form.newPassword" autocomplete="new-password" />
      </a-form-item>
      <a-button block type="primary" html-type="submit" :loading="loading">
        Change password
      </a-button>
    </a-form>
  </a-card>
</template>
