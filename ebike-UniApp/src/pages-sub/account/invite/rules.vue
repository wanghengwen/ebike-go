<template>
  <view class="page">
    <view class="card" v-if="loading">{{ t('common.loading') }}</view>
    <view class="card" v-else>
      <view class="title">{{ t('account.inviteRules') }}</view>
      <view class="row">1. {{ t('account.inviteRuleNewUser', { brand }) }}</view>
      <view class="row">
        2. {{ t('account.inviteRuleSuccess') }}
        <text class="sub">{{ t('account.inviteRuleValidRide') }}</text>
      </view>
      <view class="row" v-if="rule.inviteNum != null">
        3. {{ t('account.inviteRuleNum', { n: rule.inviteNum }) }}
        <text v-if="rewardDetail"> — {{ rewardDetail }}</text>
      </view>
      <view class="row" v-if="rewardTypeName">
        4. {{ t('account.inviteRuleValidPeriod', { type: rewardTypeName, days: availableDays }) }}
      </view>
      <view class="row" v-if="rewardTypeName">
        5. {{ t('account.inviteRuleUse', { type: rewardTypeName }) }}
      </view>
      <view class="row">6. {{ t('account.inviteRuleAntiFraud', { brand }) }}</view>
      <view class="row" v-if="rewardTypeName">
        7. {{ t('account.inviteRuleAbandon', { brand, type: rewardTypeName }) }}
      </view>
      <view class="row">8. {{ t('account.inviteRuleLatest') }}</view>
      <view class="row">
        9. {{ t('account.inviteRuleCs', { brand, phone: phone || '--' }) }}
        <text v-if="startTime || endTime">
          ({{ t('account.inviteRuleCsHours', { start: startTime || '--', end: endTime || '--' }) }})
        </text>
      </view>
      <view class="body" v-if="rule.content || rule.rule || rule.desc">
        {{ rule.content || rule.rule || rule.desc }}
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { getInviteRule } from '@/api/invite'
import { getCustomerService } from '@/api/service'
import { fenToYuan } from '@/features/pay/usePay'
import { getTenantConfig } from '@/shared/config'
import { setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { logger } from '@/shared/logger'

const { t } = useI18n()
const loading = ref(true)
const rule = ref<Record<string, unknown>>({})
const phone = ref('')
const startTime = ref('')
const endTime = ref('')

const brand = computed(() => getTenantConfig().name || 'E-Bike')

const rewardTypeName = computed(() => {
  if (rule.value.rewardType == null && rule.value.ridingCardNum == null && rule.value.amount == null) {
    return ''
  }
  return Number(rule.value.rewardType) === 1
    ? t('account.inviteRideCard')
    : t('account.inviteBalance')
})

const rewardDetail = computed(() => {
  if (Number(rule.value.rewardType) === 1 || rule.value.ridingCardNum != null) {
    return t('account.inviteRuleCards', {
      n: rule.value.ridingCardNum || 0,
      d: rule.value.validDay || 0,
    })
  }
  if (rule.value.amount != null) {
    return t('account.inviteRuleAmount', { m: fenToYuan(rule.value.amount) })
  }
  return ''
})

const availableDays = computed(() => {
  if (Number(rule.value.rewardType) === 1 || rule.value.ridingCardNum != null) {
    return `${rule.value.validDay || 0}${t('account.inviteDayUnit')}`
  }
  return t('account.inviteNoExpiry')
})

onShow(() => setNavTitle(t('account.inviteRules')))

onLoad(async (q) => {
  const sid = decodeURIComponent(String(q?.serviceId || storage.get('serviceId', '') || ''))
  loading.value = true
  try {
    const [ruleRes, csRes] = await Promise.all([
      getInviteRule(sid ? { serviceId: sid } : {}),
      getCustomerService(sid ? { serviceId: sid } : {}),
    ])
    if (ruleRes.success && ruleRes.data) rule.value = ruleRes.data as Record<string, unknown>
    if (csRes.success && csRes.data) {
      const data = csRes.data as { tel?: string; startTime?: string; endTime?: string }
      phone.value = data.tel ? String(data.tel).split(',')[0] : ''
      startTime.value = String(data.startTime || '')
      endTime.value = String(data.endTime || '')
    }
  } catch (e) {
    logger.warn('invite rules soft fail', e)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped lang="scss">
.title {
  font-weight: 700;
  font-size: 32rpx;
  margin-bottom: 24rpx;
}
.row {
  color: #555;
  line-height: 1.6;
  margin-bottom: 16rpx;
  font-size: 26rpx;
}
.sub {
  display: block;
  color: #666;
  margin-top: 6rpx;
}
.body {
  margin-top: 20rpx;
  color: #666;
  line-height: 1.6;
  white-space: pre-wrap;
}
</style>
