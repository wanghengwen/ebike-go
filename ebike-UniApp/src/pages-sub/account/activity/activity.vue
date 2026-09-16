<template>
  <view class="page">
    <view v-if="loading" class="state">{{ t('common.loading') }}</view>
    <view v-else-if="!visibleBlocks.length" class="empty">
      <image v-if="emptyIcon" class="empty__img" :src="emptyIcon" mode="aspectFit" />
      <text>{{ t('account.activityEmpty') }}</text>
    </view>

    <scroll-view v-else scroll-y class="scroll" :style="{ height: scrollH }">
      <view v-for="(item, index) in activityList" :key="index">
        <!-- type 1: banner carousel -->
        <view class="banner" v-if="item.type === 1 && isShowContentItem(item)">
          <swiper class="banner__swiper" :indicator-dots="true" :autoplay="false">
            <swiper-item
              v-for="(banner, i) in bannerList(item)"
              :key="i"
              @click="onBannerClick(banner)"
            >
              <image class="banner__img" :src="String(banner.thumb || '')" mode="aspectFill" />
            </swiper-item>
          </swiper>
        </view>

        <!-- type 7: invite -->
        <view class="invitation" v-if="item.type === 7 && isShowContentItem(item)">
          <image
            v-if="inviteBg"
            class="invitation__bg"
            :src="inviteBg"
            mode="widthFix"
          />
          <view class="invitation__content" :class="{ 'invitation__content--plain': !inviteBg }">
            <image
              v-if="inviteTextImg"
              class="invitation__text-img"
              :src="inviteTextImg"
              mode="widthFix"
            />
            <text v-else class="invitation__title">{{ t('account.invite') }}</text>
            <text class="invitation__desc">
              {{
                t('account.activityInviteReward', {
                  reward: formatInviteReward(item.inviteDetail),
                })
              }}
            </text>
            <view class="invitation__btn" @click="goInvite(item.inviteDetail)">
              <text>{{ t('account.activityInviteNow') }}</text>
              <image
                v-if="arrowIcon"
                class="invitation__arrow"
                :src="arrowIcon"
                mode="aspectFit"
              />
            </view>
          </view>
        </view>

        <!-- type 3: riding cards -->
        <view class="card section" v-if="item.type === 3 && isShowContentItem(item)">
          <view class="section__head">
            <text class="section__title">{{ t('account.activityRidingCard') }}</text>
            <view class="section__more" @click="viewAllCards">
              <text>{{ t('account.activityViewAll') }}</text>
              <image
                v-if="arrowIcon"
                class="section__arrow"
                :src="arrowIcon"
                mode="aspectFit"
              />
            </view>
          </view>
          <view
            class="ride-row"
            v-for="(riding, i) in ridingCards(item)"
            :key="i"
          >
            <image
              class="ride-row__img"
              :src="cardThumb(riding)"
              mode="aspectFill"
            />
            <view class="ride-row__des">
              <view class="ride-row__name">
                {{ riding.ridingCardName || riding.cardName || riding.name || '-' }}
              </view>
              <view class="ride-row__tags" v-if="cardTags(riding).length">
                <template v-for="(tag, ind) in cardTags(riding)" :key="ind">
                  <text class="ride-row__tag">{{ tag }}</text>
                  <view
                    v-if="ind + 1 !== cardTags(riding).length"
                    class="ride-row__line"
                  />
                </template>
              </view>
              <view class="ride-row__price">
                <text class="ride-row__cur">{{ fenToYuan(cardCurCost(riding)) }}</text>
                <text>{{ t('ride.yuan') }}</text>
                <text
                  v-if="cardOriginCost(riding) > cardCurCost(riding)"
                  class="ride-row__old"
                >
                  {{ fenToYuan(cardOriginCost(riding)) }}{{ t('ride.yuan') }}
                </text>
              </view>
            </view>
            <view class="ride-row__buy" @click="openCardDetail(riding)">
              {{ t('account.activityBuy') }}
            </view>
          </view>
        </view>

        <!-- type 6: task center -->
        <view class="card section" v-if="item.type === 6 && isShowContentItem(item)">
          <view class="section__head">
            <text class="section__title">{{ t('account.activityTaskCenter') }}</text>
          </view>
          <view
            class="task-row"
            v-for="(act, i) in regularList(item)"
            :key="i"
          >
            <image
              class="task-row__img"
              :src="taskIcon"
              mode="aspectFit"
            />
            <view class="task-row__body">
              <view class="task-row__top">
                <text class="task-row__name">{{ act.name || '-' }}</text>
                <text
                  v-if="remainTimes(act) >= 0"
                  class="task-row__remain"
                >
                  {{ t('account.activityRemaining', { n: remainTimes(act) }) }}
                </text>
              </view>
              <view class="task-row__bottom">
                <view class="task-row__meta">
                  <text>{{ t('account.activityDeadline', { t: formatEndTime(act.endTime) }) }}</text>
                  <text v-if="countdownText(act)" class="task-row__cd">
                    {{ countdownText(act) }}
                  </text>
                </view>
                <view
                  class="task-row__btn"
                  :class="{ disabled: !canComplete(act) }"
                  @click="onCompleteTask(act)"
                >
                  {{
                    canComplete(act)
                      ? t('account.activityGoComplete')
                      : t('account.activityCompleted')
                  }}
                </view>
              </view>
            </view>
          </view>
        </view>

        <!-- type 2: activity explain -->
        <view class="card section" v-if="item.type === 2 && isShowContentItem(item)">
          <view class="section__head">
            <text class="section__title">{{ explainTitle(item) }}</text>
          </view>
          <view class="explain">
            <text v-if="explainSub(item)">{{ explainSub(item) }}</text>
            <view class="explain__body" v-if="explainBody(item)">{{ explainBody(item) }}</view>
          </view>
        </view>
      </view>
    </scroll-view>

    <!-- card detail sheet (card-shop pattern) -->
    <view v-if="showDetail" class="mask" @click.self="showDetail = false">
      <view class="sheet">
        <view class="sheet__head">
          <view>
            <view class="sheet__name">
              {{
                selectedCard.ridingCardName ||
                selectedCard.cardName ||
                selectedCard.name ||
                '-'
              }}
            </view>
            <view class="sheet__price">
              ¥{{ fenToYuan(cardCurCost(selectedCard)) }}
              <text
                v-if="cardOriginCost(selectedCard) > cardCurCost(selectedCard)"
                class="sheet__origin"
              >
                ¥{{ fenToYuan(cardOriginCost(selectedCard)) }}
              </text>
            </view>
          </view>
          <text class="sheet__close" @click="showDetail = false">×</text>
        </view>
        <view class="sheet__section">
          <view class="sheet__row">
            <text class="sheet__label">{{ t('account.cardUsage') }}</text>
            <text v-if="ruleUrl" class="sheet__link" @click="openRule">
              {{ t('account.cardRule') }} ›
            </text>
          </view>
          <rich-text v-if="detailHtml" class="sheet__html" :nodes="detailHtml" />
          <view v-else class="sheet__empty">{{ t('common.empty') }}</view>
        </view>
        <view class="sheet__section">
          <view class="sheet__label">{{ t('pay.payChannel') }}</view>
          <text class="sheet__channel">{{ t('pay.wechatPay') }}</text>
        </view>
        <view class="sheet__actions">
          <view class="btn-ghost" @click="showDetail = false">{{ t('common.cancel') }}</view>
          <view class="btn-primary" @click="onBuyCard">
            {{ t('pay.payNow') }} ¥{{ fenToYuan(cardCurCost(selectedCard)) }}
            <text v-if="saveFen > 0" class="sheet__save">
              {{ t('pay.savedAmount', { m: fenToYuan(saveFen) }) }}
            </text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import { onHide, onLoad, onShow } from '@dcloudio/uni-app'
import { useI18n } from 'vue-i18n'
import {
  getActivityList,
  regularGetRegisterReward,
  regularGetVerifyReward,
} from '@/api/activity'
import { ridingConfigGetRule } from '@/api/card'
import { createChannelPay, fenToYuan } from '@/features/pay/usePay'
import { checkPayCertification } from '@/features/pay/checkPayCertification'
import { openJumpAction } from '@/shared/openNotice'
import { getLoginPath, navigate, setNavTitle } from '@/shared/navigate'
import { storage } from '@/shared/storage'
import { useUserStore } from '@/stores/user'
import { getIconCfg, getMapCfg } from '@/shared/tenantSkin'
import { logger } from '@/shared/logger'

type ActivityBlock = Record<string, unknown> & {
  type?: number
  content?: unknown
  inviteDetail?: Record<string, unknown> | null
  ridingCardList?: Array<Record<string, unknown>>
  regularList?: Array<Record<string, unknown>>
}

const { t } = useI18n()
const user = useUserStore()

const loading = ref(true)
const activityList = ref<ActivityBlock[]>([])
const serviceId = ref('')
const showDetail = ref(false)
const selectedCard = ref<Record<string, unknown>>({})
const paying = ref(false)
const ruleUrl = ref('')
const nowTs = ref(Date.now())

let tickTimer: ReturnType<typeof setInterval> | null = null

const inviteBg = computed(() => getIconCfg('invitationAwardBg'))
const inviteTextImg = computed(() => getIconCfg('invitationAwardText'))
const arrowIcon = computed(() => getMapCfg('iconRight'))
const taskIcon = computed(() => getIconCfg('taskCenter') || getIconCfg('activityCenter'))
const emptyIcon = computed(() => getIconCfg('underReview'))
const scrollH = computed(() => '100vh')

const detailHtml = computed(() =>
  String(selectedCard.value.detailInfo || selectedCard.value.detail_info || ''),
)
const saveFen = computed(() => {
  const cur = cardCurCost(selectedCard.value)
  const origin = cardOriginCost(selectedCard.value)
  return origin > cur ? origin - cur : 0
})

const visibleBlocks = computed(() => activityList.value.filter((item) => isShowContentItem(item)))

onLoad((q) => {
  user.hydrateFromStorage()
  if (!ensureLogin()) return
  const sid = String(q?.service_id || q?.serviceId || storage.get('serviceId', '') || '')
  serviceId.value = sid
  const userSid = String((user.userInfo as Record<string, unknown>)?.serviceId || '')
  if (userSid && sid && userSid !== sid) {
    uni.showToast({ title: t('account.activityWrongService'), icon: 'none' })
    setTimeout(() => navigate('back'), 800)
  }
})

onShow(() => {
  setNavTitle(t('account.activity'))
  user.hydrateFromStorage()
  if (!user.isLoggedIn) return
  startTick()
  void loadList()
  void loadRule()
})

onHide(() => stopTick())
onUnmounted(() => stopTick())

function ensureLogin(): boolean {
  user.hydrateFromStorage()
  if (user.isLoggedIn) return true
  uni.showModal({
    title: t('auth.loginTitle'),
    showCancel: false,
    confirmText: t('auth.loginTitle'),
    success: (r) => {
      if (r.confirm) navigate('redirect', getLoginPath())
    },
  })
  return false
}

function startTick() {
  stopTick()
  nowTs.value = Date.now()
  tickTimer = setInterval(() => {
    nowTs.value = Date.now()
  }, 1000)
}

function stopTick() {
  if (tickTimer) {
    clearInterval(tickTimer)
    tickTimer = null
  }
}

function isShowContentItem(item: ActivityBlock): boolean {
  if (!item) return false
  switch (Number(item.type)) {
    case 1:
      return bannerList(item).length > 0
    case 2: {
      const c = item.content as Record<string, unknown> | undefined
      return Boolean(c && (c.title || c.subTitle || c.content))
    }
    case 3:
      return ridingCards(item).length > 0
    case 6:
      return regularList(item).length > 0
    case 7:
      return item.inviteDetail != null
    default:
      return false
  }
}

function bannerList(item: ActivityBlock): Array<Record<string, unknown>> {
  return Array.isArray(item.content) ? (item.content as Array<Record<string, unknown>>) : []
}

function ridingCards(item: ActivityBlock): Array<Record<string, unknown>> {
  return Array.isArray(item.ridingCardList) ? item.ridingCardList : []
}

function regularList(item: ActivityBlock): Array<Record<string, unknown>> {
  return Array.isArray(item.regularList) ? item.regularList : []
}

function explainContent(item: ActivityBlock): Record<string, unknown> {
  return (item.content && typeof item.content === 'object' && !Array.isArray(item.content)
    ? item.content
    : {}) as Record<string, unknown>
}

function explainTitle(item: ActivityBlock) {
  return String(explainContent(item).title || t('account.activity'))
}

function explainSub(item: ActivityBlock) {
  return String(explainContent(item).subTitle || '')
}

function explainBody(item: ActivityBlock) {
  return String(explainContent(item).content || '')
}

function cardTags(riding: Record<string, unknown>): string[] {
  const raw = String(riding.descriptionTag || riding.description_tag || '')
  return raw
    .split(/[,|]/)
    .map((s) => s.trim())
    .filter(Boolean)
}

function cardCurCost(riding: Record<string, unknown>) {
  return Number(riding.curCost ?? riding.current_cost ?? riding.price ?? riding.amount ?? 0)
}

function cardOriginCost(riding: Record<string, unknown>) {
  return Number(riding.originCost ?? riding.origin_cost ?? riding.originalPrice ?? 0)
}

function cardThumb(riding: Record<string, unknown>) {
  return (
    String(riding.backOfCardUrl || riding.thumb || '') ||
    getMapCfg('inviteCard1') ||
    getIconCfg('inviteCardbgBlue') ||
    getIconCfg('inviteCardbg') ||
    ''
  )
}

function formatInviteReward(detail?: Record<string, unknown> | null) {
  if (!detail) return '-'
  const type = Number(detail.rewardType)
  const info = String(detail.rewardInfo || '').split(',')[0]
  if (type === 1) return t('account.activityRewardCard', { n: info || '0' })
  if (type === 2) {
    const yuan = fenToYuan(Number(info || 0))
    return t('account.activityRewardBalance', { m: yuan })
  }
  return String(detail.rewardInfo || '-')
}

function remainTimes(act: Record<string, unknown>) {
  return Number(act.canComplete ?? 0) - Number(act.userCompleted ?? 0)
}

function canComplete(act: Record<string, unknown>) {
  return Number(act.canComplete ?? 0) > Number(act.userCompleted ?? 0)
}

function parseEndMs(endTime: unknown): number {
  if (endTime == null || endTime === '') return 0
  if (typeof endTime === 'number') {
    return endTime < 1e12 ? endTime * 1000 : endTime
  }
  const s = String(endTime).trim()
  if (/^\d+$/.test(s)) {
    const n = Number(s)
    return n < 1e12 ? n * 1000 : n
  }
  const ms = Date.parse(s.replace(/-/g, '/'))
  return Number.isFinite(ms) ? ms : 0
}

function pad2(n: number) {
  return n < 10 ? `0${n}` : String(n)
}

function formatEndTime(endTime: unknown) {
  const ms = parseEndMs(endTime)
  if (!ms) return String(endTime || '-')
  const d = new Date(ms)
  const y = d.getFullYear()
  const m = pad2(d.getMonth() + 1)
  const day = pad2(d.getDate())
  const h = pad2(d.getHours())
  const min = pad2(d.getMinutes())
  const sec = pad2(d.getSeconds())
  return `${y}-${m}-${day} ${h}:${min}:${sec}`
}

function countdownText(act: Record<string, unknown>) {
  const end = parseEndMs(act.endTime)
  if (!end) return ''
  let left = Math.floor((end - nowTs.value) / 1000)
  if (left < 0) left = 0
  const days = Math.floor(left / 86400)
  const hours = Math.floor((left % 86400) / 3600)
  const minutes = Math.floor((left % 3600) / 60)
  const seconds = left % 60
  return t('account.activityCountdown', {
    d: pad2(days),
    h: pad2(hours),
    m: pad2(minutes),
    s: pad2(seconds),
  })
}

function onBannerClick(banner: Record<string, unknown>) {
  const content = (banner.content || {}) as Record<string, unknown>
  const linkUrl = String(content.path || content.url || content.linkUrl || '')
  const appId = String(content.appId || content.appid || '')
  const title = String(content.title || '')
  const chainType = Number(banner.type ?? content.type ?? 3)
  // Legacy UserClickAction: 0 internal, 1 H5, 2 other mini program, 3 noop
  // openJumpAction: 1 internal, 2 other mini, 3 H5
  if (chainType === 3) return
  const clickEvent = chainType === 0 ? 1 : chainType === 1 ? 3 : chainType === 2 ? 2 : 0
  if (!clickEvent) return
  openJumpAction(clickEvent, {
    linkUrl,
    appId,
    params: content.params,
    title,
    linkTitle: title,
  })
}

function goInvite(detail?: Record<string, unknown> | null) {
  const id = detail?.id
  const q = id != null ? `?id=${encodeURIComponent(String(id))}` : ''
  navigate('to', `/pages-sub/account/invite/invite${q}`)
}

function viewAllCards() {
  navigate('to', '/pages-sub/account/card-shop/shop')
}

async function openCardDetail(riding: Record<string, unknown>) {
  const ok = await checkPayCertification()
  if (!ok) return
  selectedCard.value = riding
  showDetail.value = true
}

function openRule() {
  if (!ruleUrl.value) {
    uni.showToast({ title: t('common.empty'), icon: 'none' })
    return
  }
  navigate('to', `/pages/webview/webview?url=${encodeURIComponent(ruleUrl.value)}`)
}

async function onBuyCard() {
  const item = selectedCard.value
  if (!item || !Object.keys(item).length) return
  if (paying.value) return
  paying.value = true
  try {
    const id = item.cardId || item.card_id || item.id
    const cost = cardCurCost(item)
    const res = await createChannelPay({
      saleType: 'RIDING_CARD',
      totalFee: cost,
      saleInfo: {
        total_fee: cost,
        riding_card_id: id,
      },
    })
    if (res.paid || res.success) {
      showDetail.value = false
      uni.showToast({ title: t('common.submitSuccess'), icon: 'success' })
      setTimeout(() => navigate('to', '/pages-sub/account/cards/cards'), 500)
    } else if (String(res.code) === '24005') {
      uni.showToast({ title: t('account.cardBuyLimit'), icon: 'none' })
    } else {
      uni.showToast({ title: res.msg || t('pay.payFail'), icon: 'none' })
    }
  } finally {
    paying.value = false
  }
}

async function onCompleteTask(act: Record<string, unknown>) {
  if (!canComplete(act)) return
  const type = Number(act.fixedActivityType ?? act.fixed_activity_type)
  switch (type) {
    case 1:
      await claimRegisterReward()
      break
    case 2:
      await claimVerifyReward()
      break
    case 3:
      navigate('to', '/pages-sub/pay/recharge/recharge')
      break
    case 4:
    case 5:
      navigate('to', '/pages-sub/account/qualification/qualification')
      break
    case 6:
      navigate('to', '/pages-sub/account/card-shop/shop')
      break
    case 7:
      navigate('to', '/pages/map/map')
      break
    case 8:
      uni.showToast({ title: t('account.activityAdTodo'), icon: 'none' })
      break
    case 9:
      goCareer(act.careerTag)
      break
    default:
      uni.showToast({ title: t('common.networkError'), icon: 'none' })
  }
}

function goCareer(careerTag: unknown) {
  let tag: Record<string, unknown> | null = null
  if (typeof careerTag === 'string') {
    try {
      tag = JSON.parse(careerTag) as Record<string, unknown>
    } catch {
      tag = null
    }
  } else if (careerTag && typeof careerTag === 'object') {
    tag = careerTag as Record<string, unknown>
  }
  if (tag?.tagName || tag?.name) {
    const name = encodeURIComponent(String(tag.tagName || tag.name))
    const id = tag.id != null ? `&tagId=${encodeURIComponent(String(tag.id))}` : ''
    navigate('to', `/pages-sub/account/career/career?tagName=${name}${id}`)
    return
  }
  navigate('to', '/pages-sub/account/career/career')
}

async function claimRegisterReward() {
  uni.showLoading({ title: t('common.loading'), mask: true })
  try {
    const params = serviceId.value ? { serviceId: serviceId.value } : {}
    const res = await regularGetRegisterReward(params)
    if (res.success) {
      uni.showToast({ title: t('account.activityClaimOk'), icon: 'success' })
      await loadList()
    } else {
      uni.showToast({ title: res.msg || t('common.networkError'), icon: 'none' })
    }
  } catch (e) {
    logger.warn('register reward fail', e)
    uni.showToast({ title: t('common.networkError'), icon: 'none' })
  } finally {
    uni.hideLoading()
  }
}

async function claimVerifyReward() {
  uni.showLoading({ title: t('common.loading'), mask: true })
  try {
    const params = serviceId.value ? { serviceId: serviceId.value } : {}
    const res = await regularGetVerifyReward(params)
    if (res.success) {
      const data = (res.data || {}) as { needVerify?: boolean }
      if (data.needVerify) {
        navigate('to', '/pages/auth/verified')
      } else {
        uni.showToast({ title: t('account.activityClaimOk'), icon: 'success' })
        await loadList()
      }
    } else {
      uni.showToast({ title: res.msg || t('common.networkError'), icon: 'none' })
    }
  } catch (e) {
    logger.warn('verify reward fail', e)
    uni.showToast({ title: t('common.networkError'), icon: 'none' })
  } finally {
    uni.hideLoading()
  }
}

async function loadRule() {
  try {
    const sid = serviceId.value || storage.get<string>('serviceId', '') || ''
    const rule = await ridingConfigGetRule(sid ? { serviceId: sid } : {})
    if (rule.success && rule.data) {
      ruleUrl.value = String((rule.data as { url?: string }).url || '')
    }
  } catch (e) {
    logger.warn('activity card rule soft fail', e)
  }
}

async function loadList() {
  loading.value = true
  try {
    const sid = serviceId.value || storage.get<string>('serviceId', '') || ''
    const res = await getActivityList(sid ? { serviceId: sid } : {})
    if (res.success) {
      const data = res.data as ActivityBlock[] | { records?: ActivityBlock[] }
      activityList.value = Array.isArray(data) ? data : data?.records || []
    } else {
      activityList.value = []
      uni.showToast({ title: res.msg || t('common.networkError'), icon: 'none' })
    }
  } catch (e) {
    logger.warn('activity list fail', e)
    activityList.value = []
    uni.showToast({ title: t('common.networkError'), icon: 'none' })
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
.state {
  padding: 80rpx 32rpx;
  text-align: center;
  color: #999;
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 160rpx 48rpx;
  color: #999;
  font-size: 28rpx;
  text-align: center;
}

.empty__img {
  width: 280rpx;
  height: 280rpx;
  margin-bottom: 24rpx;
}

.scroll {
  box-sizing: border-box;
  padding-bottom: 48rpx;
}

.banner {
  margin: 24rpx 32rpx 0;
}

.banner__swiper {
  width: 100%;
  height: 344rpx;
  border-radius: 16rpx;
  overflow: hidden;
}

.banner__img {
  width: 100%;
  height: 100%;
}

.invitation {
  position: relative;
  margin: 24rpx 32rpx 0;
}

.invitation__bg {
  width: 100%;
  display: block;
  border-radius: 16rpx;
}

.invitation__content {
  position: absolute;
  top: 72rpx;
  left: 64rpx;
  right: 64rpx;
  display: flex;
  flex-direction: column;
  color: #fff;
  font-size: 28rpx;
  font-weight: 600;
}

.invitation__content--plain {
  position: relative;
  top: 0;
  left: 0;
  right: 0;
  padding: 40rpx 32rpx;
  border-radius: 16rpx;
  background: linear-gradient(135deg, #ff9a4d, #ff6b2b);
}

.invitation__text-img {
  width: 166rpx;
  margin-bottom: 12rpx;
}

.invitation__title {
  font-size: 36rpx;
  font-weight: 700;
  margin-bottom: 12rpx;
}

.invitation__desc {
  line-height: 1.4;
}

.invitation__btn {
  display: inline-flex;
  align-items: center;
  align-self: flex-start;
  margin-top: 16rpx;
  padding: 0 24rpx;
  height: 46rpx;
  background: #fff;
  border-radius: 24rpx;
  font-size: 24rpx;
  font-weight: 600;
  color: #ff922b;
}

.invitation__arrow {
  width: 26rpx;
  height: 24rpx;
  margin-left: 4rpx;
}

.section__head {
  display: flex;
  align-items: center;
  margin-bottom: 8rpx;
}

.section__title {
  flex: 1;
  font-size: 32rpx;
  font-weight: 700;
  color: #333;
}

.section__more {
  display: flex;
  align-items: center;
  color: #999;
  font-size: 28rpx;
}

.section__arrow {
  width: 26rpx;
  height: 24rpx;
  margin-left: 4rpx;
}

.ride-row {
  display: flex;
  align-items: center;
  margin-top: 24rpx;
  padding: 16rpx 12rpx;
  border-radius: 16rpx;
  background: rgba(58, 160, 232, 0.1);
}

.ride-row__img {
  width: 100rpx;
  height: 105rpx;
  border-radius: 12rpx;
  background: #e8f4fc;
  flex-shrink: 0;
}

.ride-row__des {
  flex: 1;
  margin-left: 24rpx;
  min-width: 0;
}

.ride-row__name {
  font-size: 28rpx;
  font-weight: 600;
  color: #333;
}

.ride-row__tags {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  margin-top: 6rpx;
  font-size: 24rpx;
  color: #666;
}

.ride-row__line {
  width: 2rpx;
  height: 24rpx;
  background: #666;
  margin: 0 10rpx;
}

.ride-row__price {
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #333;
}

.ride-row__cur {
  font-size: 36rpx;
  font-weight: 700;
}

.ride-row__old {
  margin-left: 8rpx;
  font-size: 20rpx;
  color: #999;
  text-decoration: line-through;
}

.ride-row__buy {
  width: 104rpx;
  height: 64rpx;
  line-height: 64rpx;
  text-align: center;
  background: var(--brand-color, #3aa0e8);
  border-radius: 16rpx;
  font-size: 28rpx;
  font-weight: 600;
  color: #fff;
  flex-shrink: 0;
}

.task-row {
  display: flex;
  align-items: center;
  margin-top: 28rpx;
}

.task-row__img {
  width: 112rpx;
  height: 112rpx;
  flex-shrink: 0;
  border-radius: 16rpx;
  background: #fff5eb;
}

.task-row__body {
  flex: 1;
  margin-left: 16rpx;
  min-width: 0;
}

.task-row__top {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 12rpx;
}

.task-row__name {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
  flex: 1;
}

.task-row__remain {
  font-size: 24rpx;
  color: #999;
  flex-shrink: 0;
}

.task-row__bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16rpx;
  gap: 16rpx;
}

.task-row__meta {
  flex: 1;
  font-size: 24rpx;
  color: #999;
  line-height: 1.4;
  min-width: 0;
}

.task-row__cd {
  display: block;
  margin-top: 4rpx;
  color: #ff922b;
}

.task-row__btn {
  min-width: 132rpx;
  height: 52rpx;
  padding: 0 20rpx;
  line-height: 52rpx;
  text-align: center;
  background: #ffe9d5;
  border-radius: 26rpx;
  font-size: 26rpx;
  color: #ff922b;
  flex-shrink: 0;
}

.task-row__btn.disabled {
  opacity: 0.55;
}

.explain {
  margin-top: 16rpx;
  font-size: 28rpx;
  color: #666;
  line-height: 1.5;
}

.explain__body {
  margin-top: 16rpx;
  color: #999;
  white-space: pre-wrap;
  line-height: 1.4;
}

.mask {
  position: fixed;
  inset: 0;
  z-index: 1200;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: flex-end;
}

.sheet {
  width: 100%;
  background: #fff;
  border-radius: 32rpx 32rpx 0 0;
  padding: 32rpx 32rpx calc(32rpx + env(safe-area-inset-bottom));
  max-height: 85vh;
  overflow-y: auto;
}

.sheet__head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24rpx;
}

.sheet__name {
  font-size: 34rpx;
  font-weight: 700;
}

.sheet__price {
  margin-top: 8rpx;
  color: #ff5936;
  font-size: 36rpx;
  font-weight: 700;
}

.sheet__origin {
  margin-left: 12rpx;
  color: #999;
  font-size: 24rpx;
  font-weight: 400;
  text-decoration: line-through;
}

.sheet__close {
  font-size: 44rpx;
  color: #999;
  line-height: 1;
  padding: 0 8rpx;
}

.sheet__section {
  margin-bottom: 28rpx;
}

.sheet__row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12rpx;
}

.sheet__label {
  font-weight: 600;
  color: #333;
}

.sheet__link {
  color: var(--brand-color, #3aa0e8);
  font-size: 24rpx;
}

.sheet__html {
  color: #666;
  font-size: 24rpx;
  line-height: 1.5;
}

.sheet__empty {
  color: #999;
  font-size: 24rpx;
}

.sheet__input {
  margin-top: 12rpx;
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 24rpx;
}

.sheet__actions {
  display: flex;
  gap: 16rpx;
}

.sheet__actions > view {
  flex: 1;
}

.sheet__save {
  display: block;
  font-size: 20rpx;
  font-weight: 400;
  opacity: 0.9;
}
</style>
