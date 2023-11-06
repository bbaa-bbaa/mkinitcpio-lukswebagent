<template>
  <header>
    <div class="mdui-container">
      <div class="mdui-row">
        <div class="mdui-col-xs-12 mdui-typo">
          <h1 style="margin-bottom: 0" v-t="'title'"></h1>
          <h4 style="margin-top: 0" v-t="'tips'"></h4>
        </div>
      </div>
    </div>
  </header>
  <form @submit="unlock" action="./unlock" method="post">
    <main>
      <div class="mdui-container">
        <div class="mdui-row">
          <div class="mdui-col-xs-12">
            <div class="mdui-textfield mdui-textfield-floating-label">
              <label class="mdui-textfield-label" v-t="'password'"></label>
              <input v-model="password" name="password" class="mdui-textfield-input" type="password" />
            </div>
          </div>
        </div>
        <div class="mdui-row">
          <div class="mdui-col-xs-12">
            <div class="mdui-col">
              <button id="unlock" :disabled="lock" type="submit"
                class="mdui-btn mdui-btn-block mdui-color-theme-accent mdui-ripple"
                :class="{ error: state == -1, success: state == 1 && !showProgress, nocontent: message != '' || showProgress, progress: showProgress }">
                {{ !showProgress ? message : "" }}
                <div v-if="showProgress" class="mdui-spinner"></div>
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>
  </form>
</template>

<script setup lang="ts">
import { onUpdated, ref, computed } from "vue";
import { useI18n } from "vue-i18n"
import mdui from "mdui";
const { t } = useI18n()
interface UnlockResp {
  error?: string
  code: number
  message?: string
}
let state = ref(0)
let lock = ref(false);
let password = ref("");
let message = ref("");
let showProgress = ref(false)
let i18n = computed(() => {
  return {
    unlock: `"${t("unlock")}"`,
    unlocking: `"${t("unlocking")}"`,
    unlocked: `"${t("unlocked")}"`
  }
});
onUpdated(() => {
  mdui.mutation();
});
function checkServerOnline() {
  const controller = new AbortController();
  const id = setTimeout(() => controller.abort(), 3500);
  fetch("/", {
    signal: controller.signal,
    mode:"no-cors",
    redirect: "manual"
  }).then(() => {
    clearTimeout(id);
    showProgress.value = false
    message.value = t("redirecting")
    setTimeout(() => {
      location.reload();
    }, 500)
  }).catch((e) => {
    if (controller.signal.aborted) {
      checkServerOnline()
    } else {
      setTimeout(() => {
        checkServerOnline()
      }, 3500)
    }
  })
}
function unlock(e: Event) {
  if (lock.value && state.value == 1) return;
  lock.value = true;
  e.preventDefault();
  setTimeout(() => {
    fetch("/unlock", {
      method: "post",
      body: JSON.stringify({ password: password.value }),
      headers: { "Content-Type": "application/json" }
    }).then((resp) => {
      return resp.json()
    }).then((unlockResp: UnlockResp) => {
      if (unlockResp.error) {
        state.value = -1;
        message.value = t("code." + unlockResp.code)
        return
      }
      state.value = 1;
      setTimeout(() => {
        message.value = t("waitserver")
        setTimeout(() => {
          message.value = ""
          showProgress.value = true;
          checkServerOnline()
        }, 2000)
      }, 2000)
    }).catch(() => {
      message.value = t("errnetwork")
      state.value = -1;
    }).finally(() => {
      if (state.value !== 1) {
        setTimeout(() => {
          lock.value = false
          message.value = ""
          state.value = 0;
        }, 1000)
      }
    });
  }, 500);
}
</script>

<style scoped>
header {
  text-align: center;
}

button#unlock.nocontent::after {
  content: "" !important;
}

button#unlock.progress {
  padding-top: 4px;
  background-color: #FF4081 !important;
  color: #ffffff !important;
}

#unlock:disabled::after {
  content: v-bind('i18n.unlocking');
}

#unlock.success {
  background-color: green !important;
  color: #fff !important;
}

#unlock.error {
  background-color: red !important;
  color: #fff !important;
}

#unlock.error::after {
  content: "" !important;
}

#unlock.success::after {
  content: v-bind('i18n.unlocked');
}

#unlock {
  transition: 0.5s all ease-in-out;
}

#unlock:after {
  content: v-bind('i18n.unlock');
}
</style>
