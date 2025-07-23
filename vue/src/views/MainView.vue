// Type declaration for MainView.vue
export {};

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL, authHeaders } from '../apiConfig'

const router = useRouter()

onMounted(async () => {
  try {
    const res = await axios.get(`${API_BASE_URL}/accounts`, { headers: { ...authHeaders() } })
    if (res.data.length === 0) {
      router.replace('/create-account')
    } else {
      router.replace('/main-account')
    }
  } catch (e) {
    router.replace('/login')
  }
})
</script>

<template>
  <div>Loading...</div>
</template> 