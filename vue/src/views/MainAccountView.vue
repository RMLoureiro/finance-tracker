// Type declaration for MainAccountView.vue
export {};

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { API_BASE_URL, authHeaders } from '../apiConfig'
import { useRouter } from 'vue-router'

interface Account {
  id: number
  name: string
  account_type: string
  balance: number
}

interface Transaction {
  id: number
  created_at: string
  ammount: number
  transaction_type_id: number
}

interface Budget {
  id: number
  name: string
  amount: number
  active: boolean
}

const router = useRouter()
const account = ref<Account | null>(null)
const transactions = ref<Transaction[]>([])
const budgets = ref<Budget[]>([])
const error = ref('')

onMounted(async () => {
  try {
    const res = await axios.get(`${API_BASE_URL}/accounts/main`, { headers: { ...authHeaders() } })
    account.value = res.data.account
    transactions.value = res.data.recent_transactions
    budgets.value = res.data.budgets
  } catch (e: any) {
    if (e.response?.status === 404) {
      router.replace('/create-account')
      return
    }
    error.value = e.response?.data?.error || 'Failed to load main account info'
    if (e.response?.status === 401) {
      router.replace('/login')
    }
  }
})
</script>

<template>
  <div class="main-account-outer">
    <div class="banner-spacer"></div>
    <div class="main-account-center">
      <div class="main-account-header">
        <h2><span class="icon">💼</span> Main Account</h2>
        <div v-if="account" class="account-summary-card">
          <div class="account-name-type">
            <span class="account-name">{{ account.name }}</span>
            <span class="account-type">({{ account.account_type === 'balace' ? 'Balance' : 'Savings' }})</span>
          </div>
          <div class="account-balance-label">Current Balance</div>
          <div class="account-balance">${{ account.balance.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</div>
        </div>
      </div>
      <div class="main-account-content">
        <div class="dashboard-section">
          <h3><span class="icon">🕑</span> Recent Transactions</h3>
          <ul class="transactions-list">
            <li v-if="transactions.length === 0" class="empty">No recent transactions.</li>
            <li v-for="tx in transactions" :key="tx.id" class="transaction-item">
              <span class="tx-date">{{ new Date(tx.created_at).toLocaleDateString() }}</span>
              <span class="tx-amount">${{ tx.ammount.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</span>
              <span class="tx-type">Type #{{ tx.transaction_type_id }}</span>
            </li>
          </ul>
        </div>
        <div class="dashboard-section">
          <h3><span class="icon">📊</span> Budgets</h3>
          <ul class="budgets-list">
            <li v-if="budgets.length === 0" class="empty">No budgets for this account.</li>
            <li v-for="b in budgets" :key="b.id" class="budget-item">
              <span class="budget-name">{{ b.name }}</span>
              <span class="budget-amount">${{ b.amount.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</span>
              <span class="budget-status" :class="{ active: b.active, inactive: !b.active }">
                {{ b.active ? 'Active' : 'Inactive' }}
              </span>
            </li>
          </ul>
        </div>
      </div>
      <div v-if="error" class="error">{{ error }}</div>
      <div v-else-if="!account" class="loading">Loading...</div>
    </div>
  </div>
</template>

<style scoped>
.main-account-outer {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  box-sizing: border-box;
}
.banner-spacer {
  height: 5.5rem;
  width: 100%;
  flex-shrink: 0;
}
.main-account-center {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.main-account-header {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  margin-bottom: 2.5rem;
}
.account-summary-card {
  background: #2d323c;
  border-radius: 12px;
  padding: 2rem 2.5rem;
  margin-top: 1rem;
  box-shadow: 0 2px 12px rgba(0,0,0,0.12);
  width: 100%;
  max-width: 420px;
}
.account-name-type {
  display: flex;
  align-items: center;
  gap: 1rem;
  font-size: 1.3rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}
.account-type {
  color: #42b983;
  font-size: 1rem;
  font-weight: 500;
}
.account-balance-label {
  color: #bbb;
  font-size: 0.95rem;
  margin-bottom: 0.2rem;
}
.account-balance {
  font-size: 2.2rem;
  font-weight: bold;
  color: #42b983;
  margin-bottom: 0.5rem;
}
.main-account-content {
  display: flex;
  gap: 2.5rem;
  flex-wrap: wrap;
  justify-content: space-between;
}
.dashboard-section {
  flex: 1 1 350px;
  background: #252a33;
  border-radius: 10px;
  padding: 1.5rem 1.2rem 1.2rem 1.2rem;
  margin-bottom: 1.5rem;
  box-shadow: 0 1px 6px rgba(0,0,0,0.10);
}
.dashboard-section h3 {
  margin-bottom: 1.2rem;
  font-size: 1.15rem;
  color: #42b983;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.icon {
  font-size: 1.2em;
}
.transactions-list, .budgets-list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.transaction-item, .budget-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.7rem 0.2rem;
  border-bottom: 1px solid #333;
  font-size: 1rem;
}
.transaction-item:last-child, .budget-item:last-child {
  border-bottom: none;
}
.tx-date {
  color: #bbb;
  font-size: 0.95rem;
  min-width: 90px;
}
.tx-amount {
  color: #42b983;
  font-weight: 600;
  min-width: 110px;
  text-align: right;
}
.tx-type {
  color: #888;
  font-size: 0.95rem;
  min-width: 80px;
  text-align: right;
}
.budget-name {
  font-weight: 500;
  color: #fff;
}
.budget-amount {
  color: #42b983;
  font-weight: 600;
  min-width: 110px;
  text-align: right;
}
.budget-status {
  font-size: 0.95rem;
  margin-left: 0.7rem;
  padding: 0.2rem 0.7rem;
  border-radius: 8px;
  font-weight: 500;
}
.budget-status.active {
  background: #2e7d32;
  color: #fff;
}
.budget-status.inactive {
  background: #b71c1c;
  color: #fff;
}
.empty {
  color: #bbb;
  font-style: italic;
  padding: 1rem 0;
  text-align: center;
}
.error {
  color: #ff3b3b;
  margin-bottom: 1rem;
  text-align: center;
}
.loading {
  color: #bbb;
  text-align: center;
  margin-top: 2rem;
}
</style> 