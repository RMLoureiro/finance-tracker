<template>
  <header class="banner">
    <span class="banner-title">Finance Tracker</span>
  </header>
  <RouterView />
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<style scoped>
.banner {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  background: #23272f;
  color: #42b983;
  font-size: 2rem;
  font-weight: bold;
  text-align: left;
  padding: 1.5rem 0 1rem 2rem;
  letter-spacing: 2px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.25);
  margin-bottom: 2rem;
  z-index: 1000;
  display: flex;
  align-items: center;
}

.banner-title {
  margin-left: 0;
}

body, #app {
  padding-top: 4.5rem; /* Adjust if needed to prevent content underlap */
  }
</style>