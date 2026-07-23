<script setup lang="ts">
import axios from 'axios'
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'

import { login } from '../api/auth'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const form = reactive({
  email: '',
  password: '',
})

async function submit() {
  loading.value = true
  try {
    const response = await login(form.email, form.password)
    auth.login({
      token: response.token,
      email: response.user.email,
      displayName: response.user.displayName,
      avatarUrl: response.user.avatarUrl,
    })
    message.success('Signed in successfully')
    await router.push('/')
  } catch (error) {
    if (axios.isAxiosError(error)) {
      if (error.response?.status === 401) {
        message.error('Invalid email or password')
        return
      }
      if (error.response?.status === 403) {
        message.warning('Email not verified. Please check your activation link.')
        return
      }
      message.error(error.response?.data.error || error.message)
      return
    }
    message.error('Sign in failed')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <a-card class="page-card" title="Sign in">
    <a-form layout="vertical" :model="form" @finish="submit">
      <a-form-item
        label="Email"
        name="email"
        :rules="[{ required: true, type: 'email', message: 'Enter a valid email' }]"
      >
        <a-input v-model:value="form.email" autocomplete="email" />
      </a-form-item>
      <a-form-item
        label="Password"
        name="password"
        :rules="[{ required: true, message: 'Enter your password' }]"
      >
        <a-input-password v-model:value="form.password" autocomplete="current-password" />
      </a-form-item>
      <a-button block type="primary" html-type="submit" :loading="loading">Sign in</a-button>
    </a-form>
  </a-card>
</template>
