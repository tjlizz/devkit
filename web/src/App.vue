<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
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
      </a-menu>
    </a-layout-header>
    <a-layout-content class="app-content">
      <RouterView />
    </a-layout-content>
  </a-layout>
</template>
