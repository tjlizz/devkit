<script setup>
import axios from 'axios'
import { computed, onMounted, ref } from 'vue'
import { message } from 'ant-design-vue'

import { approveApp, listAdminApps, rejectApp } from '../api/apps'

const apps = ref([])
const loading = ref(false)
const reviewingId = ref(null)
const notes = ref({})
const status = ref('pending_review')

const statusOptions = [
  { label: 'Pending', value: 'pending_review' },
  { label: 'Approved', value: 'approved' },
  { label: 'Rejected', value: 'rejected' },
  { label: 'All', value: 'all' },
]

const emptyDescription = computed(() =>
  status.value === 'pending_review' ? 'No pending app submissions' : 'No apps found',
)

function formatPrice(app) {
  if (app.priceCents === 0) return 'Free'
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: app.currency,
  }).format(app.priceCents / 100)
}

function formatPlanPrice(plan) {
  if (plan.priceCents === 0) return 'Free'
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: plan.currency,
  }).format(plan.priceCents / 100)
}

async function loadApps() {
  loading.value = true
  try {
    const response = await listAdminApps(status.value)
    apps.value = response.apps
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 403) {
      message.error('Admin access required')
      return
    }
    message.error('Could not load app submissions')
  } finally {
    loading.value = false
  }
}

async function approve(app) {
  reviewingId.value = app.id
  try {
    await approveApp(app.id, notes.value[app.id] || '')
    message.success('App approved and published')
    await loadApps()
  } catch (error) {
    message.error(axios.isAxiosError(error) ? error.response?.data.error || error.message : 'Approval failed')
  } finally {
    reviewingId.value = null
  }
}

async function reject(app) {
  const note = notes.value[app.id]?.trim()
  if (!note) {
    message.warning('Add a review note before rejecting')
    return
  }
  reviewingId.value = app.id
  try {
    await rejectApp(app.id, note)
    message.success('App rejected')
    await loadApps()
  } catch (error) {
    message.error(axios.isAxiosError(error) ? error.response?.data.error || error.message : 'Rejection failed')
  } finally {
    reviewingId.value = null
  }
}

onMounted(loadApps)
</script>

<template>
  <section class="admin-page">
    <div class="page-heading">
      <div>
        <a-typography-title :level="2">App reviews</a-typography-title>
        <a-typography-text type="secondary">
          Review developer software submissions before they reach the marketplace.
        </a-typography-text>
      </div>
      <a-space>
        <a-select v-model:value="status" class="status-filter" :options="statusOptions" @change="loadApps" />
        <a-button :loading="loading" @click="loadApps">Refresh</a-button>
      </a-space>
    </div>

    <a-spin :spinning="loading">
      <a-empty v-if="apps.length === 0" :description="emptyDescription" />
      <a-list v-else :data-source="apps" item-layout="vertical">
        <template #renderItem="{ item }">
          <a-list-item>
            <a-card :title="item.name" class="review-card">
              <template #extra>
                <a-tag :color="item.status === 'approved' ? 'green' : item.status === 'rejected' ? 'red' : 'gold'">
                  {{ item.status }}
                </a-tag>
              </template>
              <a-descriptions :column="1" size="small">
                <a-descriptions-item label="Developer">
                  {{ item.developerName }} <span v-if="item.developerEmail">({{ item.developerEmail }})</span>
                </a-descriptions-item>
                <a-descriptions-item label="Slug">{{ item.slug }}</a-descriptions-item>
                <a-descriptions-item label="Category">{{ item.category }}</a-descriptions-item>
                <a-descriptions-item label="Price">{{ formatPrice(item) }}</a-descriptions-item>
                <a-descriptions-item v-if="item.plans && item.plans.length > 0" label="Plans">
                  <a-space direction="vertical" size="small">
                    <a-space v-for="plan in item.plans" :key="plan.id" wrap>
                      <a-tag color="blue">{{ plan.name }}</a-tag>
                      <span>{{ formatPlanPrice(plan) }}</span>
                      <template v-if="plan.description">
                        <span class="plan-description">{{ plan.description }}</span>
                      </template>
                      <a-tag v-for="feature in plan.features" :key="feature" color="default">{{ feature }}</a-tag>
                    </a-space>
                  </a-space>
                </a-descriptions-item>
                <a-descriptions-item label="Tagline">{{ item.tagline }}</a-descriptions-item>
                <a-descriptions-item label="Description">{{ item.description }}</a-descriptions-item>
                <a-descriptions-item label="Version">{{ item.version }}</a-descriptions-item>
                <a-descriptions-item label="Links">
                  <a-space wrap>
                    <a v-if="item.demoUrl" :href="item.demoUrl" target="_blank" rel="noreferrer">Demo</a>
                    <a v-if="item.docsUrl" :href="item.docsUrl" target="_blank" rel="noreferrer">Docs</a>
                    <a v-if="item.sourceUrl" :href="item.sourceUrl" target="_blank" rel="noreferrer">Source</a>
                    <a v-if="item.supportUrl" :href="item.supportUrl" target="_blank" rel="noreferrer">Support</a>
                  </a-space>
                </a-descriptions-item>
                <a-descriptions-item label="Tags">
                  <a-space wrap>
                    <a-tag v-for="tag in item.tags" :key="tag">{{ tag }}</a-tag>
                  </a-space>
                </a-descriptions-item>
                <a-descriptions-item v-if="item.reviewNote" label="Review note">
                  {{ item.reviewNote }}
                </a-descriptions-item>
              </a-descriptions>
              <template v-if="item.status === 'pending_review'">
                <a-textarea
                  v-model:value="notes[item.id]"
                  class="review-note"
                  :maxlength="500"
                  placeholder="Review note"
                  show-count
                />
                <a-space>
                  <a-button type="primary" :loading="reviewingId === item.id" @click="approve(item)">
                    Approve
                  </a-button>
                  <a-button danger :loading="reviewingId === item.id" @click="reject(item)">
                    Reject
                  </a-button>
                </a-space>
              </template>
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
  max-width: 1040px;
}

.page-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 24px;
}

.status-filter {
  width: 150px;
}

.review-card {
  width: 100%;
}

.review-note {
  margin: 16px 0;
}

.plan-description {
  color: rgba(0, 0, 0, 0.45);
  font-size: 12px;
}
</style>
