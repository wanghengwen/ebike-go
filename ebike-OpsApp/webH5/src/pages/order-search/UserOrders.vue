<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { getUserByPin, type UserDetail } from '@/api'
import OrderScopeView from './OrderScopeView.vue'
import { isoDateTime, orDash, ridingStateLabel, yuan } from './format'

const route = useRoute()
const pin = String(route.params.pin ?? '')

const user = ref<UserDetail | null>(null)
const failed = ref(false)

const scope = computed(() => ({ userPin: pin }))

onMounted(async () => {
  try {
    const { success, data } = await getUserByPin(pin)
    if (success && data?.pin) user.value = data
    else failed.value = true
  } catch {
    failed.value = true
  }
})
</script>

<template>
  <div class="page">
    <div v-if="user" class="card">
      <div class="card__head">
        <span class="card__name">{{ orDash(user.authName) }}</span>
        <span class="card__state">{{ ridingStateLabel(user.ridingState) }}</span>
      </div>
      <p class="card__row"><span>手机号</span><b>{{ orDash(user.phone) }}</b></p>
      <p class="card__row"><span>身份证</span><b>{{ orDash(user.authNo) }}</b></p>
      <p class="card__row"><span>实名</span><b>{{ user.izAuth ? '已实名' : '未实名' }}</b></p>
      <!-- 诚信金取自钱包的 depositedMount，由用户详情接口的 BFF 直接带出。 -->
      <p class="card__row"><span>诚信金</span><b>{{ yuan(user.depositAmount) }} 元</b></p>
      <p class="card__row">
        <span>押金卡</span>
        <b>{{ user.izDepositCard ? `剩余 ${user.depositedCardDays ?? 0} 天` : '未购买' }}</b>
      </p>
      <p class="card__row"><span>信用分</span><b>{{ user.creditScore ?? '--' }}</b></p>
      <p class="card__row"><span>服务区</span><b>{{ orDash(user.serviceName) }}</b></p>
      <p class="card__row"><span>注册时间</span><b>{{ isoDateTime(user.createdAt) }}</b></p>
    </div>
    <p v-else-if="failed" class="page__error">用户信息加载失败</p>

    <OrderScopeView :scope="scope" />
  </div>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f5f6f8;
  padding-bottom: 20px;
}

.page__error {
  padding: 24px 16px;
  text-align: center;
  font-size: 14px;
  color: #ee0a24;
}

.card {
  background: #ffffff;
  border-radius: 8px;
  box-shadow: 0 2px 4px 0 rgba(0, 0, 0, 0.04);
  padding: 12px 14px;
  margin: 10px 12px 0;
}

.card__head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 8px;
}

.card__name {
  font-size: 17px;
  font-weight: 600;
  color: #282828;
}

.card__state {
  font-size: 12px;
  color: #999999;
}

.card__row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  line-height: 22px;
}

.card__row span {
  color: #999999;
}

.card__row b {
  color: #282828;
  font-weight: 500;
}
</style>
