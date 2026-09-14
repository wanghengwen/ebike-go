<template>
  <view class="page">
    <scroll-view scroll-y class="content">
      <view class="container">
        <rich-text v-if="ruleInfo" :nodes="ruleInfo" />
        <view v-else-if="loading" class="hint">{{ t('common.loading') }}</view>
        <view v-else class="hint">{{ t('common.empty') }}</view>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import { ridingConfigGetRule } from '@/api/card'
import { formatRichText } from '@/shared/format'
import { navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'

const DEFAULT_RULE =
  '<h3>生效期间 & 使用规则</h3><br/><p>骑行卡须在有效期内使用，具体的有效期根据页面当前商品展示说明为准。</p><br/><p>骑行卡在订单结算时是否可用，以开锁时骑行卡是否在生效期间为准。</p><br/><p>骑行卡仅可抵扣正常骑行费用，不可抵扣调度管理费、车辆管理费等。</p><br/><h3>购买 & 退款规则</h3><br/><p>骑行卡生效期间，不支持退换。</p><br/><h3>使用禁止规定</h3><br/><p>根据相关法律规定，严禁未成年用户骑行电动车，故不建议16周岁以下用户购买骑行卡。</p>'

const { t } = useI18n()
const loading = ref(false)
const ruleInfo = ref('')

onShow(() => setNavTitle(t('account.cardRule')))

onLoad(async (q) => {
  const type = String(q?.type || '1')
  if (type !== '1') {
    ruleInfo.value = formatRichText(DEFAULT_RULE)
    return
  }
  loading.value = true
  try {
    const sid = storage.get<string>('serviceId', '') || ''
    const res = await ridingConfigGetRule(sid ? { serviceId: sid } : {})
    if (res.success && res.data) {
      const data = res.data as { ruleInfo?: string; url?: string; content?: string }
      const html = String(data.ruleInfo || data.content || '')
      if (html) {
        ruleInfo.value = formatRichText(html)
        return
      }
      if (data.url) {
        navigate('redirect', `/pages/webview/webview?url=${encodeURIComponent(String(data.url))}`)
        return
      }
    }
    ruleInfo.value = formatRichText(DEFAULT_RULE)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped lang="scss">
.page {
  width: 100vw;
  height: 100vh;
  background: #fff;
}
.content {
  height: 100%;
}
.container {
  padding: 32rpx;
}
.hint {
  text-align: center;
  color: #999;
  padding: 80rpx 0;
}
</style>
