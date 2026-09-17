package com.luopingtech.ebike.ops.feature.admin

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.result.OpsResult
import com.luopingtech.ebike.ops.data.admin.AdminRepository
import com.luopingtech.ebike.ops.domain.admin.CareerAuditDetail
import com.luopingtech.ebike.ops.domain.admin.CareerAuditItem
import com.luopingtech.ebike.ops.domain.admin.CertificationAuditSubmit
import com.luopingtech.ebike.ops.domain.admin.CertificationPageQuery
import com.luopingtech.ebike.ops.domain.admin.ObjectionSendMode
import com.luopingtech.ebike.ops.domain.analysis.OfflineOpsTimeRanges
import com.luopingtech.ebike.ops.domain.model.ServiceArea
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

enum class CertificationNav {
    List,
    Detail,
    Submit,
}

data class ProfessionAuditUiState(
    val nav: CertificationNav = CertificationNav.List,
    val loading: Boolean = false,
    val loadingMore: Boolean = false,
    val finished: Boolean = false,
    val pageNum: Int = 1,
    val items: List<CareerAuditItem> = emptyList(),
    val detail: CareerAuditDetail? = null,
    val keyword: String = "",
    /** null = 全部 */
    val auditStateFilter: Int? = null,
    val startDate: String = "",
    val endDate: String = "",
    val submitPass: Boolean = true,
    val rejectReason: String = "",
    val noticeSys: Boolean = false,
    val noticeSms: Boolean = false,
    val noticeApp: Boolean = false,
    val sendMode: ObjectionSendMode = ObjectionSendMode(),
    val message: String? = null,
    val errorMessage: String? = null,
)

class ProfessionAuditFeature(
    private val repository: AdminRepository,
) {
    private val _state = MutableStateFlow(ProfessionAuditUiState())
    val state: StateFlow<ProfessionAuditUiState> = _state.asStateFlow()

    fun clear() {
        _state.value = ProfessionAuditUiState()
    }

    fun setKeyword(value: String) {
        _state.value = _state.value.copy(keyword = value)
    }

    fun setAuditStateFilter(state: Int?) {
        _state.value = _state.value.copy(auditStateFilter = state)
    }

    fun setDates(startDate: String, endDate: String): Boolean {
        val startMs = OfflineOpsTimeRanges.parseDateStart(startDate) ?: return false
        val endMs = OfflineOpsTimeRanges.parseDateStart(endDate) ?: return false
        if (endMs < startMs) return false
        _state.value = _state.value.copy(
            startDate = OfflineOpsTimeRanges.formatDate(startMs),
            endDate = OfflineOpsTimeRanges.formatDate(endMs),
        )
        return true
    }

    fun setRejectReason(value: String) {
        _state.value = _state.value.copy(rejectReason = value.take(20))
    }

    fun toggleNoticeSys() {
        _state.value = _state.value.copy(noticeSys = !_state.value.noticeSys)
    }

    fun toggleNoticeSms() {
        _state.value = _state.value.copy(noticeSms = !_state.value.noticeSms)
    }

    fun toggleNoticeApp() {
        _state.value = _state.value.copy(noticeApp = !_state.value.noticeApp)
    }

    fun navigateBack() {
        when (_state.value.nav) {
            CertificationNav.Submit -> _state.value = _state.value.copy(
                nav = CertificationNav.Detail,
                rejectReason = "",
                noticeSys = false,
                noticeSms = false,
                noticeApp = false,
                errorMessage = null,
            )
            CertificationNav.Detail -> _state.value = _state.value.copy(
                nav = CertificationNav.List,
                detail = null,
                errorMessage = null,
                message = null,
            )
            CertificationNav.List -> Unit
        }
    }

    suspend fun load(area: ServiceArea?) {
        ensureDefaultDates()
        search(area)
    }

    suspend fun search(area: ServiceArea?) {
        if (area == null) {
            _state.value = _state.value.copy(errorMessage = Strings.t(Str.SelectServiceAreaFirst))
            return
        }
        ensureDefaultDates()
        _state.value = _state.value.copy(
            loading = true,
            errorMessage = null,
            message = null,
            pageNum = 1,
            finished = false,
        )
        when (val result = repository.careerPage(buildQuery(area, pageNum = 1))) {
            is OpsResult.Ok -> _state.value = _state.value.copy(
                loading = false,
                items = result.value,
                finished = result.value.size < PAGE_SIZE,
                pageNum = 1,
            )
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    suspend fun loadMore(area: ServiceArea?) {
        val s = _state.value
        if (area == null || s.loading || s.loadingMore || s.finished) return
        val next = s.pageNum + 1
        _state.value = s.copy(loadingMore = true, errorMessage = null)
        when (val result = repository.careerPage(buildQuery(area, pageNum = next))) {
            is OpsResult.Ok -> _state.value = _state.value.copy(
                loadingMore = false,
                items = _state.value.items + result.value,
                pageNum = next,
                finished = result.value.size < PAGE_SIZE,
            )
            is OpsResult.Err -> _state.value = _state.value.copy(
                loadingMore = false,
                errorMessage = result.error.message,
            )
        }
    }

    suspend fun openDetail(id: String) {
        _state.value = _state.value.copy(loading = true, errorMessage = null, message = null)
        when (val result = repository.careerDetail(id)) {
            is OpsResult.Ok -> _state.value = _state.value.copy(
                loading = false,
                detail = result.value,
                nav = CertificationNav.Detail,
            )
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    suspend fun openSubmit(pass: Boolean) {
        _state.value = _state.value.copy(
            submitPass = pass,
            rejectReason = "",
            noticeSys = false,
            noticeSms = false,
            noticeApp = false,
            nav = CertificationNav.Submit,
            errorMessage = null,
        )
        when (val result = repository.careerSendMode()) {
            is OpsResult.Ok -> _state.value = _state.value.copy(sendMode = result.value)
            is OpsResult.Err -> _state.value = _state.value.copy(sendMode = ObjectionSendMode())
        }
    }

    fun canSubmit(): Boolean {
        val s = _state.value
        if (s.submitPass) return true
        return s.rejectReason.trim().isNotEmpty()
    }

    suspend fun submit(area: ServiceArea?) {
        val s = _state.value
        val detail = s.detail ?: return
        if (!s.submitPass && s.rejectReason.trim().isEmpty()) {
            _state.value = s.copy(errorMessage = Strings.t(Str.CertRejectReasonRequired))
            return
        }
        val remind = buildList {
            if (s.noticeSys) add(0)
            if (s.noticeSms) add(1)
            if (s.noticeApp) add(2)
        }
        _state.value = s.copy(loading = true, errorMessage = null, message = null)
        when (
            val result = repository.careerAudit(
                CertificationAuditSubmit(
                    id = detail.id,
                    pass = s.submitPass,
                    reason = s.rejectReason.trim(),
                    remindTypes = remind,
                ),
            )
        ) {
            is OpsResult.Ok -> {
                _state.value = _state.value.copy(
                    loading = false,
                    message = Strings.t(Str.OperationOk),
                    nav = CertificationNav.List,
                    detail = null,
                )
                delay(200)
                search(area)
            }
            is OpsResult.Err -> _state.value = _state.value.copy(
                loading = false,
                errorMessage = result.error.message,
            )
        }
    }

    private fun ensureDefaultDates() {
        val s = _state.value
        if (s.startDate.isNotBlank() && s.endDate.isNotBlank()) return
        val range = OfflineOpsTimeRanges.objectionDefaultRange()
        _state.value = s.copy(
            startDate = OfflineOpsTimeRanges.formatDate(range.startMs),
            endDate = OfflineOpsTimeRanges.formatDate(range.endMs),
        )
    }

    private fun buildQuery(area: ServiceArea, pageNum: Int): CertificationPageQuery {
        val s = _state.value
        return CertificationPageQuery(
            serviceId = area.id,
            pageNum = pageNum,
            pageSize = PAGE_SIZE,
            auditState = s.auditStateFilter,
            keyword = s.keyword,
            startTime = s.startDate.takeIf { it.isNotBlank() }?.let { "$it 00:00:00" }.orEmpty(),
            endTime = s.endDate.takeIf { it.isNotBlank() }?.let { "$it 23:59:59" }.orEmpty(),
        )
    }

    companion object {
        private const val PAGE_SIZE = 20
    }
}
