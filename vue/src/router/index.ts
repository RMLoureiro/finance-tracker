import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
// Main page: checks for account existence and redirects
import MainView from '../views/MainView.vue'
// Account creation page
import CreateAccountView from '../views/CreateAccountView.vue'
// Main account info page
import MainAccountView from '../views/MainAccountView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'Main',
      component: MainView,
    },
    {
      path: '/login',
      name: 'Login',
      component: LoginView,
    },
    {
      path: '/create-account',
      name: 'CreateAccount',
      component: CreateAccountView,
    },
    {
      path: '/main-account',
      name: 'MainAccount',
      component: MainAccountView,
    },
  ],
})

export default router
