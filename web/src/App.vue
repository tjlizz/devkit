<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { useAuthStore } from './stores/auth'

const route = useRoute()
const auth = useAuthStore()
const selectedKeys = computed(() => [route.path])
</script>

<template>
  <a-layout class="app-shell">
    <a-layout-header class="app-header">
      <RouterLink class="brand" to="/">DevKit</RouterLink>
      <a-menu
        :selected-keys="selectedKeys"
        class="navigation"
        mode="horizontal"
        theme="dark"
      >
        <a-menu-item key="/">
          <RouterLink to="/">Home</RouterLink>
        </a-menu-item>
        <a-menu-item key="/login">
          <RouterLink to="/login">Login</RouterLink>
        </a-menu-item>
        <a-menu-item key="/register">
          <RouterLink to="/register">Register</RouterLink>
        </a-menu-item>
        <a-menu-item v-if="auth.isAuthenticated" key="/change-password">
          <RouterLink to="/change-password">Change Password</RouterLink>
        </a-menu-item>
        <a-menu-item v-if="auth.isAuthenticated" key="/profile-settings">
          <RouterLink to="/profile-settings">Profile Settings</RouterLink>
        </a-menu-item>
      </a-menu>
      <div v-if="auth.isAuthenticated && auth.user" class="header-user">
        <a-avatar :src="auth.user.avatarUrl" :size="32" />
        <span>{{ auth.user.displayName }}</span>
      </div>
    </a-layout-header>
    <a-layout-content class="app-content">
      <RouterView />
    </a-layout-content>
  </a-layout>
</template>
