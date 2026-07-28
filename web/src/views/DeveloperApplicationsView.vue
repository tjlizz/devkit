<script setup>
import axios from 'axios'
import { onMounted, ref } from 'vue'
import { message } from 'ant-design-vue'

import {
  approveDeveloperApplication,
  listDeveloperApplications,
  rejectDeveloperApplication,
} from '../api/auth'

const applications = ref([])
const loading = ref(false)
const reviewingId = ref(null)
const notes = ref({})

async function loadApplications() {
  loading.value = true
  try {
    const response = await listDeveloperApplications()
    applications.value = response.applications
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 403) {
      message.error('Admin access required')
      return
    }
    message.error('Could not load developer applications')
  } finally {
    loading.value = false
  }
}

async function approve(application) {
  reviewingId.value = application.id
  try {
    await approveDeveloperApplication(application.id, notes.value[application.id] || '')
    message.success('Developer application approved')
    await loadApplications()
  } catch (error) {
    message.error(axios.isAxiosError(error) ? error.response?.data.error || error.message : 'Approval failed')
  } finally {
    reviewingId.value = null
  }
}

async function reject(application) {
  const note = notes.value[application.id]?.trim()
  if (!note) {
    message.warning('Add a review note before rejecting')
    return
  }
  reviewingId.value = application.id
  try {
    await rejectDeveloperApplication(application.id, note)
    message.success('Developer application rejected')
    await loadApplications()
  } catch (error) {
    message.error(axios.isAxiosError(error) ? error.response?.data.error || error.message : 'Rejection failed')
  } finally {
    reviewingId.value = null
  }
}

onMounted(loadApplications)
</script>

<template>
  <section class="admin-page">
    <div class="page-heading">
      <div>
        <a-typography-title :level="2">Developer applications</a-typography-title>
        <a-typography-text type="secondary">
          Review pending marketplace developer onboarding requests.
        </a-typography-text>
      </div>
      <a-button :loading="loading" @click="loadApplications">Refresh</a-button>
    </div>

    <a-spin :spinning="loading">
      <a-empty
        v-if="applications.length === 0"
        description="No pending developer applications"
      />
      <a-list v-else :data-source="applications" item-layout="vertical">
        <template #renderItem="{ item }">
          <a-list-item>
            <a-card :title="item.displayName" class="application-card">
              <a-descriptions :column="1" size="small">
                <a-descriptions-item label="Email">{{ item.email }}</a-descriptions-item>
                <a-descriptions-item label="Profile">
                  <a v-if="item.profileUrl" :href="item.profileUrl" target="_blank" rel="noreferrer">
                    {{ item.profileUrl }}
                  </a>
                  <span v-else>Not provided</span>
                </a-descriptions-item>
                <a-descriptions-item label="Reason">{{ item.reason }}</a-descriptions-item>
                <a-descriptions-item label="Submitted">{{ item.createdAt }}</a-descriptions-item>
              </a-descriptions>
              <a-textarea
                v-model:value="notes[item.id]"
                class="review-note"
                :maxlength="500"
                placeholder="Review note"
                show-count
              />
              <a-space>
                <a-button
                  type="primary"
                  :loading="reviewingId === item.id"
                  @click="approve(item)"
                >
                  Approve
                </a-button>
                <a-button
                  danger
                  :loading="reviewingId === item.id"
                  @click="reject(item)"
                >
                  Reject
                </a-button>
              </a-space>
            </a-card>
          </a-list-item>
        </template>
      </a-list>
    </a-spin>
  </section>
</template>

<style scoped>
.admin-page {
  margin: 0 auto;
  max-width: 960px;
}

.page-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 24px;
}

.application-card {
  width: 100%;
}

.review-note {
  margin: 16px 0;
}
</style>
