<script setup lang="ts">
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'

const router = useRouter()
const form = reactive({
  email: '',
  password: '',
  displayName: '',
})

async function submit() {
  message.success('Registration form is ready to connect to the API')
  await router.push('/login')
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
      <a-button block type="primary" html-type="submit">Register</a-button>
    </a-form>
  </a-card>
</template>
