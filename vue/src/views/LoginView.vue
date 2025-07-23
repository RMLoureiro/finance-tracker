<script setup lang="ts">
import axios from 'axios'
import { ref, nextTick } from 'vue'
import { RouterView, useRouter } from 'vue-router'
import { API_BASE_URL } from '../apiConfig';
import { authHeaders } from '../apiConfig';
const showLogin = ref(true)
const loginEmail = ref('')
const loginPassword = ref('')
const registerEmail = ref('')
const registerPassword = ref('')
const registerConfirm = ref('')

const registerError = ref('')
const registerErrorVisible = ref(false)

const loginError = ref('')
const loginErrorVisible = ref(false)

async function showRegisterError(message: string) {
  registerError.value = message
  registerErrorVisible.value = true
  await nextTick()
  setTimeout(() => {
    registerErrorVisible.value = false
  }, 3500)
}

async function showLoginError(message: string) {
  loginError.value = message
  loginErrorVisible.value = true
  await nextTick()
  setTimeout(() => {
    loginErrorVisible.value = false
  }, 3500)
}

async function handleLogin() {
  loginError.value = ''
  try {
    const res = await axios.post(`${API_BASE_URL}/api/login`, {
      username: loginEmail.value,
      password: loginPassword.value
    })
    // Store JWT
    localStorage.setItem('jwt', res.data.token)
    window.location.href = '/'
  } catch (e: any) {
    if (e.response && e.response.data && e.response.data.error) {
      showLoginError(e.response.data.error)
    } else {
      showLoginError('Login failed')
    }
  }
}

async function handleRegister() {
  const password = registerPassword.value
  const confirm = registerConfirm.value
  const passwordRegex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[\W_]).{8,}$/

  if (password !== confirm) {
    showRegisterError('Passwords do not match!')
    return
  }
  if (!passwordRegex.test(password)) {
    showRegisterError(
      'Password must be at least 8 characters long and include a number, a special character, an uppercase and a lowercase letter.'
    )
    return
  }
  try {
    const res = await axios.post(`${API_BASE_URL}/users`, {
      user: {
        username: registerEmail.value,
        name: registerEmail.value,
        password
      }
    }, { withCredentials: true })
    // On success, switch to login view
    showLogin.value = true
  } catch (e: any) {
    if (e.response && e.response.data && e.response.data.errors) {
      showRegisterError(e.response.data.errors.join(', '))
    } else {
      showRegisterError('Registration failed')
    }
  }
}

function logout() {
  localStorage.removeItem('jwt')
  window.location.href = '/login'
}
</script>

<template>
  <div class="auth-outer">
    <main class="auth-container">
      <div v-if="showLogin" class="auth-form">
        <h2>Login</h2>
        <form @submit.prevent="handleLogin">
          <input v-model="loginEmail" type="email" placeholder="Email" required />
          <input v-model="loginPassword" type="password" placeholder="Password" required />
          <button type="submit">Login</button>
          <transition name="fade">
            <div v-if="loginErrorVisible" class="tooltip-error">
              {{ loginError }}
            </div>
          </transition>
        </form>
        <p>
          Don't have an account?
          <a href="#" @click.prevent="showLogin = false">Create Account</a>
        </p>
      </div>
      <div v-else class="auth-form">
        <h2>Create Account</h2>
        <form @submit.prevent="handleRegister">
          <input v-model="registerEmail" type="email" placeholder="Email" required />
          <input v-model="registerPassword" type="password" placeholder="Password" required />
          <input v-model="registerConfirm" type="password" placeholder="Confirm Password" required />
          <button type="submit">Register</button>
          <transition name="fade">
            <div v-if="registerErrorVisible" class="tooltip-error">
              {{ registerError }}
            </div>
          </transition>
        </form>
        <p>
          Already have an account?
          <a href="#" @click.prevent="showLogin = true">Login</a>
        </p>
      </div>
    </main>
    <RouterView />
  </div>
</template>

<style scoped>
.auth-outer {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
}

.auth-container {
  width: 400px;
  min-height: 420px;
  padding: 2.5rem 2rem;
  background: #23272f;
  border-radius: 10px;
  box-shadow: 0 2px 16px rgba(0,0,0,0.28);
  color: #f3f3f3;
}

.auth-form {
  width: 100%;
  min-height: 340px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.auth-form h2 {
  margin-bottom: 1.5rem;
  color: #fff;
  text-align: center;
}

.auth-form form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  position: relative;
}

.auth-form input {
  padding: 0.75rem;
  border: 1px solid #444;
  border-radius: 4px;
  background: #181b20;
  color: #f3f3f3;
  font-size: 1rem;
}

.auth-form input::placeholder {
  color: #888;
}

.auth-form button {
  padding: 0.75rem;
  background: #42b983;
  color: #fff;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
  font-size: 1rem;
}

.auth-form button:hover {
  background: #369870;
}

.auth-form p {
  margin-top: 1.5rem;
  font-size: 1rem;
  text-align: center;
  color: #bbb;
}

.auth-form a {
  color: #42b983;
  cursor: pointer;
  text-decoration: underline;
}

.tooltip-error {
  position: absolute;
  left: 0;
  right: 0;
  top: 100%;
  margin: 0 auto;
  background: #ff3b3b;
  color: #fff;
  padding: 0.75rem 1rem;
  border-radius: 6px;
  font-size: 1rem;
  text-align: center;
  box-shadow: 0 2px 8px rgba(0,0,0,0.18);
  z-index: 10;
  animation: shake 0.3s;
}

@keyframes shake {
  0% { transform: translateX(0);}
  20% { transform: translateX(-5px);}
  40% { transform: translateX(5px);}
  60% { transform: translateX(-5px);}
  80% { transform: translateX(5px);}
  100% { transform: translateX(0);}
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}

body {
  background: #181b20;
}
</style>