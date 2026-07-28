<script setup lang="ts">
import axios from 'axios'
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'

import { resetAvatar, updateAvatar } from '../api/auth'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const loading = ref(false)
const resetting = ref(false)
const avatarFile = ref<File | null>(null)
const objectUrl = ref('')
const previewUrl = computed(() => objectUrl.value || auth.user?.avatarUrl || '')

watch(
  avatarFile,
  (file) => {
    if (objectUrl.value) {
      URL.revokeObjectURL(objectUrl.value)
      objectUrl.value = ''
    }
    if (file) {
      objectUrl.value = URL.createObjectURL(file)
    }
  },
)

onBeforeUnmount(() => {
  if (objectUrl.value) {
    URL.revokeObjectURL(objectUrl.value)
  }
})

function selectAvatar(event = new Event('change')) {
  if (!(event.target instanceof HTMLInputElement)) {
    avatarFile.value = null
    return
  }
  avatarFile.value = event.target.files?.[0] || null
}

async function submit() {
  if (!auth.isAuthenticated) {
    message.error('Please sign in first')
    await router.push('/login')
    return
  }
  if (!avatarFile.value) {
    message.error('Choose an image file first')
    return
  }
  loading.value = true
  try {
    const response = await updateAvatar(avatarFile.value)
    auth.updateUser(response.user)
    avatarFile.value = null
    message.success('Avatar updated successfully')
  } catch (error) {
    if (axios.isAxiosError(error)) {
      message.error(error.response?.data.error || error.message)
      return
    }
    message.error('Avatar update failed')
  } finally {
    loading.value = false
  }
}

async function resetToDefault() {
  if (!auth.isAuthenticated) {
    message.error('Please sign in first')
    await router.push('/login')
    return
  }
  resetting.value = true
  try {
    const response = await resetAvatar()
    auth.updateUser(response.user)
    avatarFile.value = null
    message.success('Default avatar restored')
  } catch (error) {
    if (axios.isAxiosError(error)) {
      message.error(error.response?.data.error || error.message)
      return
    }
    message.error('Avatar reset failed')
  } finally {
    resetting.value = false
  }
}
</script>

<template>
  <a-card class="page-card" title="Profile settings">
    <a-form layout="vertical" @finish="submit">
      <a-flex align="center" gap="middle" class="avatar-preview">
        <a-avatar :src="previewUrl" :size="64" />
        <div class="avatar-copy">
          <strong>{{ auth.user?.displayName }}</strong>
          <span>{{ auth.user?.email }}</span>
        </div>
      </a-flex>

      <a-form-item label="Avatar image" extra="PNG, JPG, GIF, or WebP. Maximum size is 2MB.">
        <input
          accept="image/png,image/jpeg,image/gif,image/webp"
          class="avatar-input"
          type="file"
          @change="selectAvatar"
        >
      </a-form-item>
      <a-space direction="vertical" style="width: 100%">
        <a-button block type="primary" html-type="submit" :disabled="!avatarFile" :loading="loading">
          Upload avatar
        </a-button>
        <a-button block :loading="resetting" @click="resetToDefault">Use default avatar</a-button>
      </a-space>
    </a-form>
  </a-card>
</template>

<style scoped>
.avatar-preview {
  margin-bottom: 24px;
}

.avatar-copy {
  display: flex;
  min-width: 0;
  flex-direction: column;
}

.avatar-copy strong,
.avatar-copy span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.avatar-copy span {
  color: #6b7280;
  font-size: 0.875rem;
}

.avatar-input {
  display: block;
  width: 100%;
}
</style>
