<script setup lang="ts">
import { useI18n } from "vue-i18n"
const { t } = useI18n()
document.title = t("title")
import Unlock from "@/components/Unlock.vue";
</script>

<template>
  <div class="mdui-container main">
    <div class="mdui-card card-container">
      <Unlock />
    </div>

  </div>
</template>

<style scoped>
@media (min-width: 840px) {
  .main {
    padding-top: 80px;
  }
}

@media (min-width: 480px) {
  .main {
    padding-top: 60px;
  }
}

.main {
  padding-top: 40px;
}

.card-container {
  background-color: #f0f0f0;
  opacity: .85;
  padding-bottom: 10px;
}
</style>
