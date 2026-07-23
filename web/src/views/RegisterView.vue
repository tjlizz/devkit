<script setup lang="ts">
import axios from 'axios'
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'

import { register } from '../api/auth'

const router = useRouter()
const loading = ref(false)
const form = reactive({
  email: '',
  password: '',
  displayName: '',
})

async function submit() {
  loading.value = true
  try {
    await register(form.email, form.password, form.displayName)
    message.info(
      'Registration successful! Check the server log for the activation link. You can close this window.',
    )
    await router.push('/login')
  } catch (error) {
    if (axios.isAxiosError(error)) {
      if (error.response?.status === 409) {
        message.error('This email is already registered')
        return
      }
      message.error(error.response?.data.error || error.message)
      return
    }
    message.error('Registration failed')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <a-card class="page-card" title="Create an account">
    <a-form layout="vertical" :model="form" @finish="submit">
      <a-form-item
        label="Display name"
        name="displayName"
        :rules="[{ required: true, message: 'Enter your display name' }]"
      >
        <a-input v-model:value="form.displayName" autocomplete="name" />
      </a-form-item>
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
        :rules="[{ required: true, min: 8, message: 'Use at least 8 characters' }]"
      >
        <a-input-password v-model:value="form.password" autocomplete="new-password" />
      </a-form-item>
      <a-button block type="primary" html-type="submit" :loading="loading">Register</a-button>
    </a-form>
  </a-card>
</template>
