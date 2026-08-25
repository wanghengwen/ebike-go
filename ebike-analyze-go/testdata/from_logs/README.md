# 从 logs 提取的测试数据（ebike-analyze-go）

- 生成时间: 2026-07-10T11:35:31
- 日志文件: 4 个
- 解析请求数: 176
- 解析失败: 0
- 唯一路径: 7
- 跳过路径: 7

## 已跳过（无需检查）

- `/analyze/member`
- `/carServiceStatistics/getList`
- `/orderQuery/selectOrderCount`
- `/parking/orderStatistics/list`
- `/user/ageStatistic/v2`
- `/user/allUserStatistic`
- `/user/userStatisticHour`

## 已提取路径

| 次数 | 形状数 | 路径 |
|------|--------|------|
| 130 | 57 | `/userQuery/selectUserCount` |
| 15 | 10 | `/car-statistics/queryByCarList` |
| 9 | 4 | `/riding_card_order/page` |
| 8 | 6 | `/orderQuery/selectOrderAnalyze` |
| 8 | 8 | `/user/ageStatistic` |
| 5 | 3 | `/parking/statistics` |
| 1 | 1 | `/parking/getParkingInAndOutflow` |
