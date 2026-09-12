<script setup lang="ts">
import { ref } from 'vue'
import { numFormat, numFormatInt } from '@/utils/numFormat'
import { theme } from '@/utils/theme'
import type { StatGroup, StatItem } from './useRevenueData'

defineProps<{ group: StatGroup }>()

const detailOf = ref<StatItem | null>(null)
</script>

<template>
  <div class="stat">
    <div class="stat__head">
      <p class="stat__total" :style="{ color: theme.themeColor }">
        <CountUpValue :value="group.total" :color="theme.themeColor" />
        <span class="stat__unit">元<span class="stat__tip">{{ group.tip }}</span></span>
      </p>
      <p class="stat__formula">{{ group.formula }}</p>
    </div>

    <p class="stat__symbol" :style="{ color: theme.themeColor }">=</p>

    <div class="stat__items">
      <template v-for="(item, index) in group.items" :key="item.label">
        <p v-if="index > 0" class="stat__symbol" :style="{ color: theme.themeColor }">
          {{ item.symbol }}
        </p>
        <div class="stat__item">
          <p class="stat__value">{{ numFormat(item.value) }}</p>
          <span class="stat__label">{{ item.label }}</span>
          <span
            v-if="item.detail.length > 0"
            class="stat__detail"
            :style="{ color: theme.themeColor }"
            @click="detailOf = item"
          >
            查看详情
          </span>
        </div>
      </template>
    </div>

    <van-popup :show="detailOf !== null" position="bottom" round @update:show="detailOf = null">
      <div class="detail">
        <p class="detail__title">{{ detailOf?.label }}</p>
        <div v-for="row in detailOf?.detail" :key="row.label" class="detail__row">
          <span>{{ row.label }}</span>
          <span class="detail__value">
            {{ row.integer ? numFormatInt(row.value) : numFormat(row.value) }}
          </span>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<style scoped>
.stat__total {
  height: 37px;
  font-size: 26px;
  font-weight: 600;
  line-height: 37px;
  margin-left: 14px;
}

.stat__unit {
  font-size: 15px;
  font-weight: 600;
  line-height: 21px;
}

.stat__tip {
  font-size: 10px;
  font-weight: 400;
  color: #999999;
  line-height: 14px;
}

.stat__formula {
  font-size: 12px;
  font-weight: 400;
  color: #646464;
  line-height: 17px;
  margin: 0 16px 4px;
}

.stat__symbol {
  height: 30px;
  margin-left: 33px;
  font-size: 30px;
  line-height: 30px;
}

.stat__items {
  display: flex;
  flex-direction: column;
}

.stat__item {
  height: 73px;
  background: #ffffff;
  box-shadow: 0 2px 4px 0 rgba(0, 0, 0, 0.04);
  border-radius: 4px;
  margin: 11px 10px 0;
  position: relative;
}

.stat__value {
  font-size: 20px;
  font-weight: 600;
  color: #505050;
  line-height: 28px;
  margin: 12px 0 0 16px;
}

.stat__label {
  font-size: 12px;
  font-weight: 400;
  color: #646464;
  line-height: 17px;
  margin-left: 16px;
}

.stat__detail {
  font-size: 12px;
  font-weight: 400;
  line-height: 17px;
  position: absolute;
  right: 16px;
  bottom: 12px;
}

.detail {
  padding: 16px;
}

.detail__title {
  font-size: 15px;
  font-weight: 600;
  color: #282828;
  margin-bottom: 8px;
}

.detail__row {
  display: flex;
  justify-content: space-between;
  padding: 10px 0;
  font-size: 13px;
  color: #646464;
  border-bottom: 1px solid #f2f2f2;
}

.detail__row:last-child {
  border-bottom: none;
}

.detail__value {
  font-size: 15px;
  font-weight: 600;
  color: #282828;
}
</style>
