<script setup lang="ts">
import { ref } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'

const emit = defineEmits<{
  login: []
}>()

const passkey = ref('')
const error = ref('')

const handleLogin = async () => {
  try {
    const response = await fetch('/api/auth', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ passkey: passkey.value })
    })
    const data = await response.json()
    if (response.ok) {
      localStorage.setItem('token', data.token)
      emit('login')
    } else {
      error.value = data.error || 'Invalid passkey'
    }
  } catch {
    error.value = 'Connection failed'
  }
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center">
    <div class="flex flex-col items-center gap-6">
      <div class="text-center">
        <h1 class="text-5xl font-bold">Amazon Q Account Hub</h1>
        <p class="py-6 text-muted-foreground">Enter your passkey to access the dashboard</p>
      </div>
      <Card class="w-96">
        <CardHeader>
          <CardTitle>Login</CardTitle>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <label class="text-sm font-medium">Passkey</label>
            <Input
              v-model="passkey"
              type="password"
              placeholder="Enter passkey"
              @keyup.enter="handleLogin"
            />
          </div>
          <Alert v-if="error" variant="destructive">
            <AlertDescription>{{ error }}</AlertDescription>
          </Alert>
          <Button class="w-full" @click="handleLogin">Login</Button>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
