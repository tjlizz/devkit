<script setup lang="ts">
import axios from 'axios'
import { reactive, ref } from 'vue'
import { message } from 'ant-design-vue'

import { forgotPassword } from '../api/auth'

const loading = ref(false)
const sent = ref(false)
const form = reactive({
  email: '',
})

async function submit() {
  loading.value = true
  sent.value = false
  try {
    await forgotPassword(form.email)
    sent.value = true
    message.success('If that email is registered, a reset link has been sent.')
  } catch (error) {
    if (axios.isAxiosError(error)) {
      message.error(error.response?.data.error || error.message)
      return
    }
    message.error('Password reset request failed')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <a-card class="page-card" title="Forgot password">
    <a-alert
      v-if="sent"
      class="form-alert"
      type="success"
      show-icon
      message="Check your email"
      description="If the address is registered, it will receive a password reset link shortly."
    />
    <a-form layout="vertical" :model="form" @finish="submit">
      <a-form-item
        label="Email"
        name="email"
        :rules="[{ required: true, type: 'email', message: 'Enter a valid email' }]"
      >
        <a-input v-model:value="form.email" autocomplete="email" />
      </a-form-item>
      <a-button block type="primary" html-type="submit" :loading="loading">
        Send reset link
      </a-button>
      <div class="form-footer">
        <RouterLink to="/login">Back to sign in</RouterLink>
      </div>
    </a-form>
  </a-card>
</template>

<style scoped>
.form-alert {
  margin-bottom: 16px;
}

.form-footer {
  margin-top: 16px;
  text-align: center;
}
</style>
