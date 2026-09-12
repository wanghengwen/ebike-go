<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { showToast } from 'vant'
import { createUpdateCostTicket, type OrderItem } from '@/api'
import { yuan } from './format'

const props = defineProps<{ show: boolean; order: OrderItem }>()
const emit = defineEmits<{ 'update:show': [boolean]; done: [] }>()

const rideCost = ref('')
const dispatchCost = ref('')
const helmetPenalty = ref('')
const submitting = ref(false)

/** 各项上限就是原值。界面单位是元，接口单位是分，换算只发生在提交那一步。 */
const caps = computed(() => ({
  ride: (props.order.originCost ?? 0) / 100,
  dispatch: (props.order.dispatchCost ?? 0) / 100,
  helmet: (props.order.helmetPenalty ?? 0) / 100,
}))

watch(
  () => props.show,
  (open) => {
    if (!open) return
    rideCost.value = caps.value.ride.toFixed(2)
    dispatchCost.value = caps.value.dispatch.toFixed(2)
    helmetPenalty.value = caps.value.helmet.toFixed(2)
  },
)

/** 空、非数、负数、超过两位小数、超过原值都不算数。 */
function parse(input: string, cap: number): number | null {
  const text = input.trim()
  if (!text || !/^\d+(\.\d{1,2})?$/.test(text)) return null
  const value = Number(text)
  return value > cap ? null : value
}

const parsed = computed(() => ({
  ride: parse(rideCost.value, caps.value.ride),
  dispatch: parse(dispatchCost.value, caps.value.dispatch),
  helmet: parse(helmetPenalty.value, caps.value.helmet),
}))

const valid = computed(() => parsed.value.ride !== null && parsed.value.dispatch !== null)

/** 后端会拒绝改后合计低于已实付的单，这里先算给用户看，别等提交完才报错。 */
const belowPaid = computed(() => {
  const { ride, dispatch, helmet } = parsed.value
  if (ride === null || dispatch === null) return false
  const total = Math.floor(ride * 100) + Math.floor(dispatch * 100) + Math.floor((helmet ?? 0) * 100)
  return total < (props.order.hasPaid ?? 0)
})

async function submit(): Promise<void> {
  const { ride, dispatch, helmet } = parsed.value
  if (ride === null || dispatch === null) {
    showToast('金额不能为空、超过原值或超过两位小数')
    return
  }
  if (belowPaid.value) {
    showToast(`改后合计不能低于已实付 ${yuan(props.order.hasPaid)} 元`)
    return
  }
  if (!props.order.id) return

  submitting.value = true
  try {
    // 遗留实现用 Math.floor 取整到分，这里保持一致，避免 0.1+0.2 之类的浮点尾巴多出 1 分。
    const { success, msg } = await createUpdateCostTicket({
      orderId: props.order.id,
      modifyPayCost: Math.floor(ride * 100),
      modifyDispatchCost: Math.floor(dispatch * 100),
      ...(helmet === null ? {} : { modifyHelmetPenalty: Math.floor(helmet * 100) }),
    })
    if (!success) {
      showToast(msg || '改价失败')
      return
    }
    showToast('已提交改价工单')
    emit('update:show', false)
    emit('done')
  } catch {
    showToast('改价失败，请重试')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <van-popup
    :show="show"
    position="bottom"
    round
    @update:show="emit('update:show', $event)"
  >
    <div class="modify">
      <p class="modify__title">修改订单金额</p>

      <van-field
        v-model="rideCost"
        type="number"
        label="骑行费用"
        :placeholder="`不超过 ${caps.ride.toFixed(2)}`"
      >
        <template #extra>元</template>
      </van-field>
      <van-field
        v-model="dispatchCost"
        type="number"
        label="调度费用"
        :placeholder="`不超过 ${caps.dispatch.toFixed(2)}`"
      >
        <template #extra>元</template>
      </van-field>
      <van-field
        v-model="helmetPenalty"
        type="number"
        label="头盔罚金"
        :placeholder="`不超过 ${caps.helmet.toFixed(2)}`"
      >
        <template #extra>元</template>
      </van-field>

      <p v-if="belowPaid" class="modify__warn">
        改后合计低于已实付 {{ yuan(order.hasPaid) }} 元，后端会拒绝
      </p>

      <div class="modify__actions">
        <van-button block @click="emit('update:show', false)">取消</van-button>
        <van-button
          block
          type="primary"
          :disabled="!valid || belowPaid"
          :loading="submitting"
          @click="submit"
        >
          提交
        </van-button>
      </div>
    </div>
  </van-popup>
</template>

<style scoped>
.modify {
  padding: 16px 0 20px;
}

.modify__title {
  font-size: 16px;
  font-weight: 600;
  color: #282828;
  padding: 0 16px 12px;
}

.modify__warn {
  padding: 10px 16px 0;
  font-size: 12px;
  color: #ee0a24;
}

.modify__actions {
  display: flex;
  gap: 12px;
  padding: 16px 16px 0;
}
</style>
