package com.luopingtech.ebike.ops.feature.admin

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.admin.AdminRepository
import com.luopingtech.ebike.ops.domain.admin.ObjectionDealRequest
import com.luopingtech.ebike.ops.domain.admin.ObjectionOrderDetail
import com.luopingtech.ebike.ops.domain.admin.ObjectionOrderItem
import com.luopingtech.ebike.ops.domain.admin.ObjectionPageQuery
import com.luopingtech.ebike.ops.domain.admin.ObjectionSendMode
import com.luopingtech.ebike.ops.domain.admin.ObjectionStates
import com.luopingtech.ebike.ops.domain.admin.filterObjectionAmountYuan
import com.luopingtech.ebike.ops.domain.admin.filterObjectionCardTimes
import com.luopingtech.ebike.ops.domain.admin.objectionYuanToFen
import com.luopingtech.ebike.ops.domain.admin.processAmounts
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTimeRanges
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

sealed class ObjectionNav {
    data object List : ObjectionNav()
    data object Detail : ObjectionNav()
    data class Process(val feeReasonable: Boolean) : ObjectionNav()
}

data class ObjectionOrderUiState(
    val nav: ObjectionNav = ObjectionNav.List,
    val loading: Boolean = false,
    val loadingMore: Boolean = false,
    val items: List<ObjectionOrderItem> = emptyList(),
    val pageNum: Int = 1,
    val finished: Boolean = false,
    val stateFilter: Int? = null,
    val keyword: String = "",
    val createdTimeStart: String = "",
    val createdTimeEnd: String = "",
    val detail: ObjectionOrderDetail? = null,
    val processReason: String = "",
    val modifyPayYuan: String = "",
    val modifyDispatchYuan: String = "",
    val modifyHelmetYuan: String = "",
    val refundPayYuan: String = "",
    val refundDispatchYuan: String = "",
    val refundHelmetYuan: String = "",
    val refundCardTimes: String = "",
    val sendMode: ObjectionSendMode = ObjectionSendMode(),
    val noticeSys: Boolean = false,
    val noticeSms: Boolean = false,
    val noticeApp: Boolean = false,
    val noticeMessage: String? = null,
    val submitting: Boolean = false,
    val message: String? = null,
    val errorMessage: String? = null,
)

class ObjectionOrderFeature(
    private val repository: AdminRepository,
) {
    private val _state = MutableStateFlow(ObjectionOrderUiState())
    val state: StateFlow<ObjectionOrderUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = ObjectionOrderUiState()
    }

    fun consumeMessage() {
        _state.value = _state.value.copy(message = null, errorMessage = null, noticeMessage = null)
    }

    fun setKeyword(value: String) {
        _state.value = _state.value.copy(keyword = value)
    }

    fun setStateFilter(state: Int?) {
        _state.value = _state.value.copy(stateFilter = state)
    }

    /**
     * 设置创建日期（`yyyy-MM-dd`）。开始晚于结束时返回 false 并写入错误文案，不改动区间。
     */
    fun setCreatedDates(startDate: String, endDate: String): Boolean {
        val startMs = OfflineOpsTimeRanges.parseDateStart(startDate)
        val endMs = OfflineOpsTimeRanges.parseDateStart(endDate)
        if (startMs == null || endMs == null) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.ObjectionDateInvalid))
            return false
        }
        if (startMs > endMs) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.ObjectionDateInvalid))
            return false
        }
        _state.value = _state.value.copy(
            createdTimeStart = OfflineOpsTimeRanges.formatDate(startMs) + " 00:00:00",
            createdTimeEnd = OfflineOpsTimeRanges.formatDate(endMs) + " 23:59:59",
            errorMessage = null,
        )
        return true
    }

    fun setProcessReason(value: String) {
        _state.value = _state.value.copy(processReason = value.take(25))
    }

    fun setModifyPayYuan(value: String) {
        val amounts = _state.value.detail?.processAmounts() ?: return
        val cur = _state.value.modifyPayYuan
        _state.value = _state.value.copy(
            modifyPayYuan = filterObjectionAmountYuan(value, amounts.originCostYuanMax(), cur),
        )
    }

    fun setModifyDispatchYuan(value: String) {
        val amounts = _state.value.detail?.processAmounts() ?: return
        val cur = _state.value.modifyDispatchYuan
        _state.value = _state.value.copy(
            modifyDispatchYuan = filterObjectionAmountYuan(value, amounts.dispatchCostYuanMax(), cur),
        )
    }

    fun setModifyHelmetYuan(value: String) {
        val amounts = _state.value.detail?.processAmounts() ?: return
        val cur = _state.value.modifyHelmetYuan
        _state.value = _state.value.copy(
            modifyHelmetYuan = filterObjectionAmountYuan(value, amounts.helmetYuanMax(), cur),
        )
    }

    fun setRefundPayYuan(value: String) {
        val amounts = _state.value.detail?.processAmounts() ?: return
        val cur = _state.value.refundPayYuan
        // 原版意图：退还骑行上限为行程费用 payCost
        _state.value = _state.value.copy(
            refundPayYuan = filterObjectionAmountYuan(value, amounts.payCostYuanMax(), cur),
        )
    }

    fun setRefundDispatchYuan(value: String) {
        val amounts = _state.value.detail?.processAmounts() ?: return
        val cur = _state.value.refundDispatchYuan
        _state.value = _state.value.copy(
            refundDispatchYuan = filterObjectionAmountYuan(value, amounts.dispatchCostYuanMax(), cur),
        )
    }

    fun setRefundHelmetYuan(value: String) {
        val amounts = _state.value.detail?.processAmounts() ?: return
        val cur = _state.value.refundHelmetYuan
        _state.value = _state.value.copy(
            refundHelmetYuan = filterObjectionAmountYuan(value, amounts.helmetYuanMax(), cur),
        )
    }

    fun setRefundCardTimes(value: String) {
        val cur = _state.value.refundCardTimes
        _state.value = _state.value.copy(
            refundCardTimes = filterObjectionCardTimes(value, cur),
        )
    }

    fun toggleNoticeSys() {
        val s = _state.value
        if (!s.sendMode.izSys) {
            _state.value = s.copy(noticeSys = false, noticeMessage = Strings.t(Str.ObjectionNoticeNoSysTpl))
            return
        }
        _state.value = s.copy(noticeSys = !s.noticeSys, noticeMessage = null)
    }

    fun toggleNoticeSms() {
        val s = _state.value
        if (!s.sendMode.izSms) {
            _state.value = s.copy(noticeSms = false, noticeMessage = Strings.t(Str.ObjectionNoticeNoSmsTpl))
            return
        }
        _state.value = s.copy(noticeSms = !s.noticeSms, noticeMessage = null)
    }

    fun toggleNoticeApp() {
        val s = _state.value
        if (!s.sendMode.izApp) {
            _state.value = s.copy(noticeApp = false, noticeMessage = Strings.t(Str.ObjectionNoticeNoAppTpl))
            return
        }
        _state.value = s.copy(noticeApp = !s.noticeApp, noticeMessage = null)
    }

    fun navigateBack() {
        when (_state.value.nav) {
            is ObjectionNav.Process -> {
                _state.value = _state.value.copy(
                    nav = ObjectionNav.Detail,
                    processReason = "",
                    modifyPayYuan = "",
                    modifyDispatchYuan = "",
                    modifyHelmetYuan = "",
                    refundPayYuan = "",
                    refundDispatchYuan = "",
                    refundHelmetYuan = "",
                    refundCardTimes = "",
                    noticeSys = false,
                    noticeSms = false,
                    noticeApp = false,
                    noticeMessage = null,
                    errorMessage = null,
                )
            }
            ObjectionNav.Detail -> {
                _state.value = _state.value.copy(nav = ObjectionNav.List, detail = null)
            }
            ObjectionNav.List -> Unit
        }
    }

    fun openProcess(feeReasonable: Boolean) {
        val detail = _state.value.detail ?: return
        if (detail.ticket.state != ObjectionStates.Pending) return
        _state.value = _state.value.copy(
            nav = ObjectionNav.Process(feeReasonable),
            processReason = "",
            modifyPayYuan = "",
            modifyDispatchYuan = "",
            modifyHelmetYuan = "",
            refundPayYuan = "",
            refundDispatchYuan = "",
            refundHelmetYuan = "",
            refundCardTimes = "",
            noticeSys = false,
            noticeSms = false,
            noticeApp = false,
            noticeMessage = null,
            errorMessage = null,
        )
    }

    suspend fun loadSendMode() {
        when (val result = repository.objectionSendMode()) {
            is OpsResult.Ok -> _state.value = _state.value.copy(sendMode = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(sendMode = ObjectionSendMode())
        }
    }

    suspend fun load(area: ServiceArea?) {
        if (area == null) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        val cur = _state.value
        val range = if (cur.createdTimeStart.isBlank() || cur.createdTimeEnd.isBlank()) {
            OfflineOpsTimeRanges.objectionDefaultRange()
        } else {
            null
        }
        _state.value = cur.copy(
            loading = true,
            errorMessage = null,
            pageNum = 1,
            finished = false,
            createdTimeStart = range?.startText ?: cur.createdTimeStart,
            createdTimeEnd = range?.endText ?: cur.createdTimeEnd,
            nav = ObjectionNav.List,
            detail = null,
        )
        fetchPage(area, page = 1, append = false)
    }

    suspend fun search(area: ServiceArea?) {
        if (area == null) return
        _state.value = _state.value.copy(loading = true, errorMessage = null, pageNum = 1, finished = false)
        fetchPage(area, page = 1, append = false)
    }

    suspend fun loadMore(area: ServiceArea?) {
        val s = _state.value
        if (area == null || s.loading || s.loadingMore || s.finished) return
        _state.value = s.copy(loadingMore = true, errorMessage = null)
        fetchPage(area, page = s.pageNum + 1, append = true)
    }

    suspend fun openDetail(area: ServiceArea?, item: ObjectionOrderItem) {
        if (area == null) return
        _state.value = _state.value.copy(loading = true, errorMessage = null, nav = ObjectionNav.Detail)
        when (val result = repository.objectionDetail(item.id, item.orderId)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(loading = false, detail = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                detail = ObjectionOrderDetail(ticket = item),
                errorMessage = result.error.message,
            )
        }
    }

    fun canSubmitProcess(): Boolean {
        val s = _state.value
        val detail = s.detail ?: return false
        val nav = s.nav as? ObjectionNav.Process ?: return false
        if (s.processReason.trim().isEmpty()) return false
        if (nav.feeReasonable) return true
        val amounts = detail.processAmounts()
        return if (amounts.isPaid()) {
            s.refundDispatchYuan.isNotBlank() ||
                s.refundCardTimes.isNotBlank() ||
                s.refundPayYuan.isNotBlank() ||
                s.refundHelmetYuan.isNotBlank() ||
                amounts.helmetPenaltyFen <= 0L
        } else {
            s.modifyPayYuan.isNotBlank() && s.modifyDispatchYuan.isNotBlank()
        }
    }

    suspend fun submitDeal(area: ServiceArea?) {
        val s = _state.value
        val detail = s.detail ?: return
        val nav = s.nav as? ObjectionNav.Process ?: return
        if (area == null) return
        val reason = s.processReason.trim()
        if (reason.isEmpty()) {
            _state.value = s.copy(errorMessage = Strings.t(Str.ObjectionReasonRequired))
            return
        }
        val amounts = detail.processAmounts()
        val feeReasonable = nav.feeReasonable
        if (!feeReasonable) {
            val ok = if (amounts.isPaid()) {
                s.refundDispatchYuan.isNotBlank() ||
                    s.refundCardTimes.isNotBlank() ||
                    s.refundPayYuan.isNotBlank() ||
                    s.refundHelmetYuan.isNotBlank() ||
                    amounts.helmetPenaltyFen <= 0L
            } else {
                s.modifyPayYuan.isNotBlank() && s.modifyDispatchYuan.isNotBlank()
            }
            if (!ok) {
                _state.value = s.copy(
                    errorMessage = if (amounts.isUnpaid()) {
                        Strings.t(Str.ObjectionUnpaidModifyRequired)
                    } else {
                        Strings.t(Str.ObjectionAmountRequired)
                    },
                )
                return
            }
        }
        val remindWay = buildList {
            if (s.noticeSys) add(1)
            if (s.noticeSms) add(2)
            if (s.noticeApp) add(3)
        }
        val paidRefund = !feeReasonable && amounts.isPaid()
        val helmetFen = objectionYuanToFen(s.refundHelmetYuan.ifBlank { s.modifyHelmetYuan })
        val req = ObjectionDealRequest(
            id = detail.ticket.id,
            orderId = detail.ticket.orderId,
            initiator = detail.ticket.initiator,
            feeReasonable = feeReasonable,
            isPaidRefund = paidRefund,
            opReason = reason,
            refundCostFen = if (paidRefund) objectionYuanToFen(s.refundPayYuan) else null,
            refundDispatchCostFen = if (paidRefund) objectionYuanToFen(s.refundDispatchYuan) else null,
            refundHelmetPenaltyFen = if (paidRefund) helmetFen else null,
            refundCardTimes = s.refundCardTimes.trim().toIntOrNull(),
            modifyPayCostFen = if (!feeReasonable && !paidRefund) objectionYuanToFen(s.modifyPayYuan) else null,
            modifyDispatchCostFen = if (!feeReasonable && !paidRefund) objectionYuanToFen(s.modifyDispatchYuan) else null,
            modifyHelmetPenaltyFen = if (!feeReasonable && !paidRefund) helmetFen else null,
            remindWay = remindWay,
        )
        _state.value = s.copy(submitting = true, errorMessage = null)
        when (val result = repository.dealObjection(req)) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    submitting = false,
                    message = Strings.t(Str.ObjectionDealOk),
                    nav = ObjectionNav.List,
                    detail = null,
                )
                search(area)
            }
            is OpsResult.Err -> _state.value = _state.value.copy(
                submitting = false,
                errorMessage = result.error.message,
            )
        }
    }

    private suspend fun fetchPage(area: ServiceArea, page: Int, append: Boolean) {
        val s = _state.value
        val query = ObjectionPageQuery(
            serviceId = area.id,
            pageNum = page,
            pageSize = 10,
            state = s.stateFilter,
            keyword = s.keyword,
            createdTimeStart = s.createdTimeStart,
            createdTimeEnd = s.createdTimeEnd,
        )
        when (val result = repository.objectionPage(query)) {
            is OpsResult.Ok -> {
                val merged = if (append) s.items + result.value else result.value
                _state.value = _state.value.copy(
                    loading = false,
                    loadingMore = false,
                    items = merged.distinctBy { it.id },
                    pageNum = page,
                    finished = result.value.size < query.pageSize,
                )
            }
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                loadingMore = false,
                errorMessage = result.error.message,
            )
        }
    }
}
