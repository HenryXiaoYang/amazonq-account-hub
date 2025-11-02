<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Line } from 'vue-chartjs'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend } from 'chart.js'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Alert, AlertDescription } from '@/components/ui/alert'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend)

const emit = defineEmits<{
  logout: []
}>()

interface Stats {
  current_count: number
  total_count: number
  used_count: number
  api_call_count: number
}

interface Metric {
  id: number
  current_count: number
  total_count: number
  used_count: number
  api_call_count: number
  timestamp: string
}

const stats = ref<Stats>({ current_count: 0, total_count: 0, used_count: 0, api_call_count: 0 })
const metrics = ref<Metric[]>([])
const csvInput = ref('')
const fileInput = ref<HTMLInputElement | null>(null)
const toast = ref({ show: false, message: '', type: 'success' })
const showEnableModal = ref(false)

const API_BASE = '/api'

const showToast = (message: string, type: 'success' | 'error' = 'success') => {
  toast.value = { show: true, message, type }
  setTimeout(() => toast.value.show = false, 3000)
}

const fetchStats = async () => {
  const res = await fetch(`${API_BASE}/metrics`)
  const data = await res.json()
  metrics.value = data
  if (data.length > 0) {
    stats.value = data[data.length - 1]
  }
}

const chartData = computed(() => {
  const labels: string[] = []
  const currentData: number[] = []
  const usedData: number[] = []

  metrics.value.forEach(m => {
    labels.push(new Date(m.timestamp).toLocaleString())
    currentData.push(m.current_count)
    usedData.push(m.used_count)
  })

  return {
    labels,
    datasets: [
      { label: 'Current Accounts', data: currentData, borderColor: '#3b82f6', tension: 0.1, fill: false },
      { label: 'Used Accounts', data: usedData, borderColor: '#ef4444', tension: 0.1, fill: false }
    ]
  }
})

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { position: 'top' as const } }
}

const parseCSV = (text: string) => {
  const lines = text.split('\n').filter(l => l.trim())
  const accounts = []
  for (let i = 1; i < lines.length; i++) {
    const line = lines[i]
    if (!line) continue
    const parts = line.split(',')
    const enabled = parts[0]?.toLowerCase().trim()
    if (enabled === 'false') continue
    if (parts.length >= 4 && parts[1] && parts[2]) {
      accounts.push({
        refresh_token: parts[1],
        client_id: parts[2],
        client_secret: parts[3]
      })
    }
  }
  return accounts
}

const handleCSVInput = async () => {
  const accounts = parseCSV(csvInput.value)
  await fetch(`${API_BASE}/accounts`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ accounts })
  })
  csvInput.value = ''
  await fetchStats()
  showToast(`${accounts.length} accounts added successfully`)
}

const handleFileUpload = async (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.files?.[0]) {
    const text = await target.files[0].text()
    const accounts = parseCSV(text)
    await fetch(`${API_BASE}/accounts`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ accounts })
    })
    target.value = ''
    await fetchStats()
    showToast(`${accounts.length} accounts uploaded successfully`)
  }
}

const enableAllAccounts = async () => {
  await fetch(`${API_BASE}/accounts/enable-all`, { method: 'POST' })
  showEnableModal.value = false
  await fetchStats()
  showToast('All accounts enabled successfully')
}

onMounted(() => {
  fetchStats()
  setInterval(() => {
    fetchStats()
  }, 5000)
})
</script>

<template>
  <div class="min-h-screen">
    <div class="border-b bg-white shadow-sm">
      <div class="container mx-auto flex h-16 items-center justify-between px-6">
        <h1 class="text-xl font-semibold">Amazon Q Account Hub</h1>
        <div class="flex gap-2">
          <Button @click="showEnableModal = true">Enable All Accounts</Button>
          <Button variant="ghost" @click="emit('logout')">Logout</Button>
        </div>
      </div>
    </div>

    <div class="container mx-auto p-6 max-w-7xl">
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
        <Card>
          <CardContent class="pt-6">
            <div class="text-sm text-muted-foreground">Current</div>
            <div class="text-2xl font-bold text-blue-600">{{ stats.current_count }}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent class="pt-6">
            <div class="text-sm text-muted-foreground">Total</div>
            <div class="text-2xl font-bold">{{ stats.total_count }}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent class="pt-6">
            <div class="text-sm text-muted-foreground">Used</div>
            <div class="text-2xl font-bold text-red-600">{{ stats.used_count }}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent class="pt-6">
            <div class="text-sm text-muted-foreground">API Calls</div>
            <div class="text-2xl font-bold text-cyan-600">{{ stats.api_call_count }}</div>
          </CardContent>
        </Card>
      </div>

      <Card class="mb-6">
        <CardHeader>
          <CardTitle>Account History</CardTitle>
        </CardHeader>
        <CardContent>
          <div style="height: 300px">
            <Line :data="chartData" :options="chartOptions" />
          </div>
        </CardContent>
      </Card>

      <div class="grid md:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Paste Tokens</CardTitle>
          </CardHeader>
          <CardContent class="space-y-4">
            <Textarea
              v-model="csvInput"
              class="h-48"
              placeholder="enabled,refresh_token,client_id,client_secret&#10;True,token1,client1,secret1&#10;True,token2,client2,secret2"
            />
            <Button class="w-full" @click="handleCSVInput">Add Tokens</Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Upload CSV File</CardTitle>
          </CardHeader>
          <CardContent class="space-y-4">
            <input
              ref="fileInput"
              type="file"
              accept=".csv,.txt"
              class="hidden"
              @change="handleFileUpload"
            />
            <Button class="w-full" @click="fileInput?.click()">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
              Choose CSV File
            </Button>
            <p class="text-sm text-muted-foreground">Format: enabled,refresh_token,client_id,client_secret</p>
          </CardContent>
        </Card>
      </div>
    </div>

    <div v-if="toast.show" class="fixed top-4 right-4 z-50">
      <Alert :variant="toast.type === 'success' ? 'default' : 'destructive'">
        <AlertDescription>{{ toast.message }}</AlertDescription>
      </Alert>
    </div>

    <div v-if="showEnableModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <Card class="w-96">
        <CardHeader>
          <CardTitle>Enable All Accounts</CardTitle>
        </CardHeader>
        <CardContent class="space-y-4">
          <p>Are you sure you want to enable all disabled accounts?</p>
          <div class="flex gap-2 justify-end">
            <Button variant="ghost" @click="showEnableModal = false">Cancel</Button>
            <Button @click="enableAllAccounts">Confirm</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
