<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Dashboard from './components/Dashboard.vue'
import Login from './components/Login.vue'

const isAuthenticated = ref(false)

onMounted(() => {
  isAuthenticated.value = localStorage.getItem('authenticated') === 'true'
})

const handleLogin = () => {
  isAuthenticated.value = true
  localStorage.setItem('authenticated', 'true')
}

const handleLogout = () => {
  isAuthenticated.value = false
  localStorage.removeItem('authenticated')
}
</script>

<template>
  <div class="min-h-screen bg-gray-100">
    <Login v-if="!isAuthenticated" @login="handleLogin" />
    <Dashboard v-else @logout="handleLogout" />
  </div>
</template>
