<script setup lang="ts">
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'

import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const form = reactive({
  email: '',
  password: '',
})

async function submit() {
  auth.login('development-token')
  message.success('Signed in with the example auth store')
  await router.push('/')
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
      <a-button block type="primary" html-type="submit">Sign in</a-button>
    </a-form>
  </a-card>
</template>
