// Type declaration for CreateAccountView.vue
export {};

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_BASE_URL, authHeaders } from '../apiConfig'

const router = useRouter()
const name = ref('')
const balance = ref()
const account_type = ref('balace')
const error = ref('')

onMounted(() => {
  const token = localStorage.getItem('jwt')
  if (!token) {
    router.replace('/login')
  }
})

async function createAccount() {
  error.value = ''
  try {
    await axios.post(`${API_BASE_URL}/accounts`, {
      account: { name: name.value, balance: balance.value, account_type: account_type.value }
    }, { headers: { ...authHeaders() } })
    router.replace('/')
  } catch (e: any) {
    error.value = e.response?.data?.errors?.join(', ') || 'Failed to create account'
  }
}
</script>

<template>
  <div class="account-create-outer">
    <div class="account-create-container">
      <h2>Create Account</h2>
      <form @submit.prevent="createAccount">
        <input v-model="name" type="text" placeholder="Account Name" required />
        <input v-model.number="balance" type="number" placeholder="Initial Account Balance" required />
        <select v-model="account_type">
          <option value="balace">Balance</option>
          <option value="savings">Savings</option>
        </select>
        <button type="submit">Create</button>
        <div v-if="error" class="error">{{ error }}</div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.account-create-outer {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding-top: 7rem; /* Ensure enough space for the fixed banner */
  box-sizing: border-box;
}
.account-create-container {
  max-width: 400px;
  width: 100%;
  background: #23272f;
  border-radius: 10px;
  color: #f3f3f3;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  box-sizing: border-box;
}
.account-create-container h2 {
  margin-bottom: 2rem;
}
input, select {
  width: 100%;
  margin-bottom: 1rem;
  padding: 0.75rem;
  border-radius: 4px;
  border: 1px solid #444;
  background: #181b20;
  color: #f3f3f3;
}
button {
  width: 100%;
  padding: 0.75rem;
  background: #42b983;
  color: #fff;
  border: none;
  border-radius: 4px;
  font-weight: bold;
  font-size: 1rem;
  cursor: pointer;
}
.error {
  color: #ff3b3b;
  margin-top: 1rem;
}
</style> 