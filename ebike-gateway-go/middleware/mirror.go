package middleware

import (
	"bytes"
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"ebike-gateway-go/config"
	"ebike-gateway-go/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// readOnlyEndpoints is an explicit list of paths that are safe to mirror (read-only operations).
// This prevents unintended side effects caused by matching broad keywords.
var readOnlyEndpoints = map[string]bool{
	// Business Endpoints
	"/business/ebike-operation/fix/task/analyze_data": true,
	"/business/ebike-management/component/record/pagelist": true,
	"/business/fence/parking/parkingorderstatisticpage": true,
	"/business/visual/riding_card_order/page": true,
	"/business/order/carstatistics/page": true,
	"/business/returnbikeaudit/listreturnbikeaudits": true,
	"/business/blacklist/export": true,
	"/business/ebike-marketing/ridingconfig/getrule": true,
	"/business/wealth/merchant/confirminfo": true,
	"/business/helpconfig/getfaqbyid": true,
	"/business/user/changebind/detail": true,
	"/business/baofu/account/balance_query": true,
	"/business/ebike-management/carpackage/getbyid": true,
	"/business/fence/tags/getall": true,
	"/business/systemconfig/getalarmcontact": true,
	"/business/fence/servicearea/getbytoken": true,
	"/business/wealth/billdetails/details": true,
	"/business/ebike-management/component/record/detaillist": true,
	"/business/exportusertickets": true,
	"/business/ebike-operation/repair/export": true,
	"/business/rent/gettempunlockconfig": true,
	"/business/ebike-user/creditscore/list": true,
	"/business/fence/banriding/getbyid": true,
	"/business/userticketdetail": true,
	"/business/helpconfig/gethomescrollermsgbyserviceid": true,
	"/business/returnbikeaudit/getreturnbikeauditizon": true,
	"/business/fence/bigscreen/getconfigbyserviceid": true,
	"/business/ebike-management/google/getaddress": true,
	"/business/ebike-management/billrecord/list": true,
	"/business/eibke-marketing/recharge/get_recharge_scope": true,
	"/business/fence/noparking/getbyid": true,
	"/business/order/export": true,
	"/business/blacklist/page": true,
	"/business/ebike-account/riding_card/detail": true,
	"/business/ebike-account/riding_card/use_detail": true,
	"/business/ebike-management/carinfo/offlinebycarlist": true,
	"/business/ebike-management/component/pagelist": true,
	"/business/ebike-operation/fix/page": true,
	"/business/user/tag/info": true,
	"/business/operatingbigscreen/querydata/getcarnum": true,
	"/business/ebike-management/carinfo/list": true,
	"/business/ebike-management/role/list": true,
	"/business/ebike-operation/tools/unlock_car_list": true,
	"/business/user/user/search": true,
	"/business/paas/device/detail": true,
	"/business/ebike-management/component/configuration/queryquantity": true,
	"/business/eibke-marketing/preferentialconfig/detail": true,
	"/business/fence/servicearea/getserviceareabyroleids": true,
	"/business/paas/device/ble/report/bledeviceinforeport": true,
	"/business/fence/parking/page": true,
	"/business/helpconfig/sortguidepage": true,
	"/business/paas/device/querydevicescreen": true,
	"/business/returnbikeaudit/getreturnbikeauditdetails": true,
	"/business/config/getad": true,
	"/business/ebike-management/msgtemplate/list": true,
	"/business/ebike-management/user/forget-password": true,
	"/business/ebike-operation/alarm/task/analyze_data": true,
	"/business/fence/maintain/area/personnelmanagementlist": true,
	"/business/ebike-management/carpackage/page": true,
	"/business/ebike-operation/fix/fix/getnowmonthsum": true,
	"/business/eibke-marketing/recharge/list": true,
	"/business/fence/parking/getparkingmonitordetail": true,
	"/business/ebike-management/user/getuserbypin": true,
	"/business/creditscore/v2/getconfig": true,
	"/business/returnbikeaudit/getreturnbikeauditdetailsbyid": true,
	"/business/ebike-marketing/ridingconfig/export": true,
	"/business/ebike-management/dictdata/query": true,
	"/business/ebike-management/msgtemplate/getdetail": true,
	"/business/ebike-management/repairconfig/listbycarmodel": true,
	"/business/fence/parking/getparkingmonitorlist": true,
	"/business/test/get": true,
	"/business/wealth/business-configuration/getconfigbyagentid": true,
	"/business/wealth/charge-note/getavailablewithdrawchargenotebybusinessid": true,
	"/business/ebike-management/dicttype/page": true,
	"/business/ebike-operation/alarm/task/detail": true,
	"/business/eibke-marketing/radpacketactivity/page": true,
	"/business/fence/servicearea/getserviceareabyroleidsandsubtenantid": true,
	"/business/paas/device/querysaddleoverloadcontact": true,
	"/business/systemconfig/getconfigbackcar": true,
	"/business/user/getagestatistic": true,
	"/business/ebike-management/user/getuserbyphone": true,
	"/business/eibke-marketing/voucher/detail": true,
	"/business/fence/parking/site/page": true,
	"/business/pay/transaction/page": true,
	"/business/ebike-operation/change_battery/task/export": true,
	"/business/ebike-management/dictdata/page": true,
	"/business/ebike-user/creditscore/page": true,
	"/business/fence/rfid/queryrfidsbyfenceidstr": true,
	"/business/user/tag/list": true,
	"/business/ebike-management/component/configuration/pagelist": true,
	"/business/systemconfig/getconfigusecar": true,
	"/business/ebike-management/app-data-role/get": true,
	"/business/ebike-management/deviceupgrade/record/pagelist": true,
	"/business/ebike-operation/fix/task/detail": true,
	"/business/eibke-marketing/radpacketactivity/detail": true,
	"/business/fence/noparking/getbyserviceid": true,
	"/business/fence/parking/getparkinginandoutflow": true,
	"/business/ebike-operation/move_car/task/export": true,
	"/business/ebike-account/riding_card_record/page": true,
	"/business/ebike-management/dicttype/list": true,
	"/business/ebike-management/user/getuserlistbymenuperms": true,
	"/business/ebike-operation/performance/change_battery/query": true,
	"/business/ebike-operation/repair/page": true,
	"/business/order/list": true,
	"/business/helpconfig/editguidepage": true,
	"/business/paas/device/car_count": true,
	"/business/baofu/account/balance_deal_log_query": true,
	"/business/ebike-operation/performance/move_car/query": true,
	"/business/eibke-marketing/regularactivity/regular_list": true,
	"/business/user/userstatistichour": true,
	"/business/paas/device/maplist": true,
	"/business/invoice/export": true,
	"/business/wealth/billdetails/export": true,
	"/business/ebike-management/msgtemplate/getsendmode": true,
	"/business/fence/servicearea/getserviceareabyid": true,
	"/business/ebike-management/carinfo/onlinebytaglist": true,
	"/business/ebike-management/userdevice/getbyuserid": true,
	"/business/eibke-marketing/recharge/get_recharge_config": true,
	"/business/fence/resource/management/list": true,
	"/business/paas/device/querylistbyimeimac": true,
	"/business/paas/device/querydeviceopemap": true,
	"/business/wealth/customer/accountlist": true,
	"/business/ebike-account/wallet/detail": true,
	"/business/ebike-fence/creditscore/getconfigbyserviceid": true,
	"/business/fence/parking/getbylocations": true,
	"/business/fence/config/protocol/list": true,
	"/business/paas/device/querydevicemapfake": true,
	"/business/ebike-management/employeetrack/gettrack": true,
	"/business/config/ridingpermission/get": true,
	"/business/user/career/detail": true,
	"/business/wealth/assets-details/paydetail": true,
	"/business/ebike-operation/alarm/checkresultdetail": true,
	"/business/fence/banriding/getbyserviceid": true,
	"/business/fence/servicearea/getbylocation": true,
	"/business/helpconfig/getfaqbyserviceid": true,
	"/business/paas/device/filterlist": true,
	"/business/ebike-marketing/ridingconfig/detail": true,
	"/business/ebike-account/wallet/statistics_all": true,
	"/business/ebike-management/carinfo/addcarinforole": true,
	"/business/ebike-management/menu/getmenutree": true,
	"/business/ebike-marketing/user_rewards/page": true,
	"/business/ebike-operation/alarm/tasklist": true,
	"/business/fence/parking/site/analyzeone": true,
	"/business/tenant/querylist": true,
	"/business/ebike-management/msgtemplate/gettemplatebytype": true,
	"/business/paas/device/getbluetoothtoken": true,
	"/business/ebike-management/carinfo/page": true,
	"/business/ebike-user/creditscore/noridding/info": true,
	"/business/fence/parking/getparkingmonitorpage": true,
	"/business/fence/resource/management/page": true,
	"/business/helpconfig/getmainpushconfig": true,
	"/business/tenant/pagelist": true,
	"/business/invoice/pageinvoice": true,
	"/business/wealth/assets-details/listexport": true,
	"/business/ebike-management/batterytype/list": true,
	"/business/ebike-operation/change_battery/task/page": true,
	"/business/ebike-operation/move_car/task/man_made_move_list": true,
	"/business/eibke-marketing/invite/detail": true,
	"/business/fence/noparking/getbyserviceidfast": true,
	"/business/helpconfig/addguidepageconfig": true,
	"/business/wealth/customer/detail": true,
	"/business/ebike-management/cartag/record/page": true,
	"/business/ebike-management/cartag/type/page": true,
	"/business/ebike-management/permission/detail": true,
	"/business/ebike-operation/move_car/task/analyze_data": true,
	"/business/user/user/memberstatistic": true,
	"/business/pay/withdraw/page": true,
	"/business/wealth/bill-server-area/export": true,
	"/business/ebike-management/carinfo/offlinebytaglist": true,
	"/business/ebike-management/user/getbyid": true,
	"/business/ebike-management/user/listbyserviceids": true,
	"/business/ebike-operation/move_car/check_list": true,
	"/business/fence/servicearea/getserviceareabyids": true,
	"/business/user/pay/score/getpermissionrecord": true,
	"/business/analyze-worker/hot/point/end/list": true,
	"/business/ebike-management/log/list": true,
	"/business/ebike-management/user/getmyshortcut": true,
	"/business/fence/maintain/area/personneldetail": true,
	"/business/baofu/split_config/query": true,
	"/business/ebike-operation/move_car/batch_list": true,
	"/business/helpconfig/delguidepage": true,
	"/business/ebike-management/voiceupgrade/config/pagelist": true,
	"/business/ebike-management/billrecord/getbyid": true,
	"/business/ebike-operation/change_battery/car_searching": true,
	"/business/order/blist": true,
	"/business/visual/wallet_order/export": true,
	"/business/invoice/pageinvoicerecord": true,
	"/business/ebike-management/user/getsourcebytoken": true,
	"/business/ebike-management/user/list": true,
	"/business/paas/device/info": true,
	"/business/order/orderdetail": true,
	"/business/fence/fence/custom/type/list": true,
	"/business/creditscore/v2/getuse": true,
	"/business/ebike-operation/fix/task/export": true,
	"/business/ebike-management/component/configuration/queryallcomponentnamemaplist": true,
	"/business/ebike-management/dictdata/getmapbytypes": true,
	"/business/fence/parking/getbyserviceidfast": true,
	"/business/user/user/export": true,
	"/business/paas/device/pagebus": true,
	"/business/wealth/billdetails/areaexport": true,
	"/business/ebike-operation/alarm/page": true,
	"/business/fence/parking/getbyserviceid": true,
	"/business/fence/resource/management/applist": true,
	"/business/multi/query": true,
	"/business/paas/device/export": true,
	"/business/ebike-management/voiceupgrade/record/pagelist": true,
	"/business/ebike-operation/change_battery/task/analyze_data": true,
	"/business/ebike-operation/fix/task/page": true,
	"/business/ebike-operation/move_car/task/daily_analyze": true,
	"/business/eibke-marketing/regularactivity/regular_detail": true,
	"/business/fence/maintain/area/listbyserviceid": true,
	"/business/creditscore/v2/page": true,
	"/business/ebike-management/component/failurecomponentnamelist/export": true,
	"/business/ebike-management/repairconfig/list": true,
	"/business/fence/parking/getlist": true,
	"/business/fence/maintain/area/pagelistbyserviceid": true,
	"/business/ebike-management/gaode/getaddress": true,
	"/business/paas/device/getcarliststatistical": true,
	"/business/ebike-account/riding_card/query_by_sysno": true,
	"/business/ebike-management/carinfo/getbind": true,
	"/business/ebike-management/component/querybycomponentno": true,
	"/business/tenant/query": true,
	"/business/user/user/detail": true,
	"/business/wealth/customer/getaccountinfo": true,
	"/business/ebike-management/role/getbyid": true,
	"/business/ebike-operation/fix/taskdetail": true,
	"/business/notice/detail": true,
	"/business/ebike-management/user/getlistbyroleid": true,
	"/business/order/carstatistics/export": true,
	"/business/ebike-operation/alarm/task/export": true,
	"/business/ebike-marketing/user_rewards/target_user_reward": true,
	"/business/ebike-management/home/page/role/getbyroleid": true,
	"/business/fence/parking/parkingorderstatistic/export": true,
	"/business/ebike-management/carinfo/onlinebycarlist": true,
	"/business/ebike-management/deviceupgrade/exportimei": true,
	"/business/pageusertickets": true,
	"/business/fence/parking/site/analyze": true,
	"/business/ebike-management/carinfo/checkcarlistinserviceid": true,
	"/business/ebike-management/voltageplan/list": true,
	"/business/ebike-operation/task_rules/pagelist": true,
	"/business/eibke-marketing/voucher/list": true,
	"/business/user/addauthinfo": true,
	"/business/ebike-marketing/user_rewards/export": true,
	"/business/ebike-marketing/ridingconfig/list": true,
	"/business/wealth/business-assets/getinfo": true,
	"/business/wealth/customer/getaccountbytenantname": true,
	"/business/baofu/account/balance_deal_query": true,
	"/business/ebike-management/component/configuration/querybrandlist": true,
	"/business/visual/wallet_order/page": true,
	"/business/wealth/bill/exportbyid": true,
	"/business/applicationsite/pagesiteapplication": true,
	"/business/ebike-operation/move_car/task/detail": true,
	"/business/ebike-management/deviceupgrade/config/pagelist": true,
	"/business/user/user/getlastlogininfo": true,
	"/business/ebike-management/menu/getmenutreebyroleid": true,
	"/business/eibke-marketing/preferentialconfig/pagelist": true,
	"/business/eibke-marketing/regularactivity/activation_regular_list": true,
	"/business/systemconfig/getconfigbaseitem": true,
	"/business/paas/device/page": true,
	"/business/ebike-operation/performance/export": true,
	"/business/ebike-operation/fix/check_result_detail": true,
	"/business/fence/servicearea/getfencebyserviceidfast": true,
	"/business/helpconfig/getcustomerservicebyserviceid": true,
	"/business/paas/device/querycarimeibind": true,
	"/business/user/changebind/page": true,
	"/business/user/user/detailbyphone": true,
	"/business/ebike-marketing/ridingconfig/page": true,
	"/business/eibke-marketing/invite/pagelist": true,
	"/business/fence/servicearea/getall": true,
	"/actuator/health": true,
	"/business/eibke-marketing/radpacketactivity/export": true,
	"/business/wealth/customer/getbalancenew": true,
	"/business/ebike-account/user_account": true,
	"/business/ebike-management/carinfo/detail": true,
	"/business/ebike-management/voltageplan/page": true,
	"/business/fence/parking/getbyid": true,
	"/business/getcount": true,
	"/business/returnbikeaudit/pagereturnbikeaudits": true,
	"/business/wealth/billdetails/arealist": true,
	"/business/ebike-management/employeetrack/getalloperation": true,
	"/business/ebike-operation/change_battery/task/list": true,
	"/business/helpconfig/gethomeactivityentrancebyserviceid": true,
	"/business/visual/riding_card_order/export": true,
	"/business/ebike-operation/move_car/task/page": true,
	"/business/eibke-marketing/voucher/export": true,
	"/business/wealth/customer/list": true,
	"/business/baofu/account/detail": true,
	"/business/helpconfig/gethomenav": true,
	"/business/orderanalyze/getorderanalyze": true,
	"/business/eibke-marketing/radpacketactivity/getactivityinfo": true,
	"/business/pay/transaction/export": true,
	"/business/applicationsite/getconfig": true,
	"/business/order/riding/export": true,
	"/business/ebike-management/user/getpersonalizeconfig": true,
	"/business/ebike-management/user/stationpage": true,
	"/business/ebike-operation/config/config/query": true,
	"/business/ebike-operation/task_rules/detail": true,
	"/business/notice/page": true,
	"/business/order/config/get": true,
	"/business/systemconfig/getconfigpay": true,
	"/business/fence/servicearea/getnearfence": true,
	"/business/wealth/customer/openaccount": true,
	"/business/wealth/assets-details/getbydealnum": true,
	"/business/ebike-operation/change_battery/task/daily_analyze": true,
	"/business/ridingpermission/get": true,
	"/business/ebike-management/gaode/getlocation": true,
	"/business/ebike-management/component/configuration/queryallcomponentnamelist": true,
	"/business/fence/fence/custom/list": true,
	"/business/order/detaillast": true,
	"/business/fence/servicearea/getnearfencebylocations": true,
	"/business/user/auth/page": true,
	"/business/ebike-operation/alarm/task/page": true,
	"/business/ebike-operation/fix/list": true,
	"/business/wealth/employer/confirminfo": true,
	"/business/baofu/account/list": true,
	"/business/helpconfig/getguidepageconfigbyserviceid": true,
	"/business/helpconfig/getspecialtipsbyid": true,
	"/business/ebike-management/user/pagelist": true,
	"/business/ebike-operation/move_car/page": true,
	"/business/ebike-operation/repair/detail": true,
	"/business/fence/servicearea/getfencebyserviceid": true,
	"/business/helpconfig/gethomescrollermsgbyid": true,
	"/business/ebike-management/operation/visual/op_count": true,
	"/business/wealth/bill/export": true,
	"/business/ebike-management/carinfo/export": true,
	"/business/ebike-marketing/user_rewards/user_reward_detail": true,
	"/business/fence/parking/getpagebyserviceid": true,
	"/business/visual/deposit_order/page": true,
	"/business/wealth/billdetails/page": true,
	"/business/fence/parkingsite/carlist": true,
	"/business/helpconfig/getspecialtipsbyserviceid": true,
	"/business/user/user/page": true,
	"/business/wealth/customer/addsubaccount": true,
	"/business/tenant/carpackage/details": true,
	"/business/fence/rfid/queryrfidsbyfenceid": true,
	"/business/order/config/getall": true,
	"/business/ebike-management/user/getuserbytoken": true,
	"/business/tenant/authinfo": true,
	"/business/user/getalluserstatistic": true,
	"/business/wealth/bill-server-area/page": true,
	"/business/wealth/assets-details/listbypage": true,
	"/business/ebike-management/cartag/type/list": true,
	"/business/ebike-management/role/applist": true,
	"/business/operatingbigscreen/polyline/carinfo": true,
	"/business/paas/device/list": true,
	"/business/user/career/page": true,
	"/business/wealth/bill/page": true,
	"/business/analyze-worker/hot/point/start/list": true,
	"/business/ebike-operation/change_battery/task/detail": true,
	"/business/ebike-management/role/pagelist": true,
	"/business/ebike-management/google/getlocation": true,
	"/business/wealth/billdetailsfree/list": true,
	
	// Operation
	"/client/operation/repair/page": true,
	
	// User
	"/client/user/user/personinfo": true,
	"/client/user/user/thirdbindinfo": true,
	"/client/user/user/qualificationlist": true,
	"/client/user/user/canride": true,
	"/client/user/blacklist/info": true,
	"/client/gray/queryorder": true,
	"/client/gray/userdata": true,
	"/client/user/auth/izneed": true,
	"/client/user/auth/state": true,
	"/client/user/config/enable": true,
	"/client/ebike-user/creditscore/noridding/info": true,
	"/client/ebike-user/creditscore/list": true,
	"/client/ebike-user/creditscore/page": true,
	"/client/tenant/config": true,
	
	// Order
	"/client/order/calculatecost": true,
	"/client/order/list": true,
	"/client/order/detaillast": true,
	"/client/order/detail": true,
	"/client/order/queryfrozen": true,
	"/client/order/listinvoicedorders": true,
	"/client/order/config/get": true,
	"/client/invoice/page": true,
	"/client/returnbikeaudit/izcancameraaudit": true,
	"/client/returnbikeaudit/izcapable": true,
	"/client/userticket/getuserticket": true,
	"/client/paas/device/ebikelocation": true,
	"/client/paas/device/getbluetoothtoken": true,
	
	// Pay — getOpenIdByJsCode consumes WeChat js_code (single-use); do not mirror
	"/client/ebike-pay/pay/withdraw/page": true,
	// permission opens WeChat pay-score authorization — write operation, do not mirror
	"/client/pay/score/getpermissionrecord": true,
	"/client/pay/score/getconfig": true,
	"/client/zhima/payafteruse/signquery": true,
	"/client/zhima/payafteruse/orderquery": true,
	
	// Marketing
	"/client/ebike-marketing/activity_center/default_activity_center_list": true,
	"/client/ebike-marketing/card/riding_config_list": true,
	"/client/ebike-marketing/card/riding_config_get_rule": true,
	"/client/ebike-marketing/invite/detail": true,
	"/client/ebike-marketing/invite/record": true,
	"/client/ebike-marketing/invite/rule": true,
	"/client/ebike-marketing/recharge_config/list": true,
	"/client/ebike-marketing/recharge_config/recharge_scope": true,
	"/client/ebike-marketing/recharge_config/get_recharge_config": true,
	"/client/ebike-marketing/redpaketcar/getrule": true,
	// user_watch_ad_judgement triggers reward judgement — write operation, do not mirror
	"/client/ebike-account/riding_card/get_service_riding_card": true,
	"/client/ebike-account/user_account": true,
	"/client/ebike-account/riding_card/get_riding_card": true,
	"/client/ebike-account/favorable_card/get_user_favorable_card": true,
	"/client/ebike-account/free_order/get_user_all_free_order": true,
	"/client/ebike-account/discount/get_user_all_discount": true,
	"/client/ebike-account/deposit_card/get_user_deposit_card": true,
	"/client/ebike-account/wallet/get_wallet_info": true,
	
	// Fence
	"/client/fence/servicearea/getall": true,
	"/client/fence/servicearea/getbylocation": true,
	"/client/fence/servicearea/getfencebyserviceid": true,
	"/client/fence/servicearea/getserviceareabyid": true,
	"/client/fence/servicearea/getnearfence": true,
	// returncar/ridingcar omitted — not registered in ebike-service-client-go
	"/client/fence/parking/getbyid": true,
	"/client/fence/parking/getbyserviceid": true,
	"/client/fence/parking/nearparkingnum": true,
	"/client/fence/noparking/getbyid": true,
	"/client/fence/noparking/getbyserviceid": true,
	"/client/fence/config/protocol/bytype": true,
	"/client/fence/config/protocol/default": true,
	"/client/fence/resource/management/applist": true,
	"/client/ebike-fence/creditscore/getconfig": true,
	
	// Management
	// code/verify consumes the SMS code — write operation, do not mirror
	"/client/messagecenter/page/list": true,
	// getById marks the message read (MsgRecordGatewayImpl.updateById) — do not mirror
	"/client/messagecenter/judgeishavenoreadmsg": true,
	"/client/management/gaode/getaddress": true,
	"/client/management/gaode/getlocation": true,
	"/client/management/gaode/navigate": true,
	"/client/management/gaode/v2/navigate": true,
	"/client/management/gaode/v3/navigate": true,
	"/client/management/google/getaddress": true,
	"/client/management/google/getlocation": true,
	"/client/management/repairconfig/list": true,
	"/client/management/repairconfig/listbycar": true,
	
	// Clientconfig
	"/client/helpconfig/getadconfig": true,
	"/client/config/getad": true,
	"/client/ad_config/detail": true,
	"/client/helpconfig/getfaqbyserviceid": true,
	"/client/helpconfig/getfaqbyid": true,
	"/client/helpconfig/gethomescrollermsgbyserviceid": true,
	"/client/helpconfig/gethomescrollermsgbyserviceid/v2": true,
	"/client/helpconfig/gethomescrollermsgbyid": true,
	"/client/helpconfig/getguidepageconfigbyserviceid": true,
	"/client/helpconfig/getspecialtipsbyserviceid": true,
	"/client/helpconfig/getcustomerservicebyserviceid": true,
	"/client/helpconfig/gethomeactivityentrancebyserviceid": true,
	"/client/helpconfig/gethomeactivitybyid": true,
	"/client/helpconfig/gethomenavbyserviceid": true,
	"/client/helpconfig/gethomenavbyid": true,
	"/client/helpconfig/getizmainpush": true,
	"/client/system/getusecarconfig": true,
	"/client/system/getbackcarconfig": true,
	"/client/system/getbackcarconfigbycarid": true,
	"/client/systemconfig/getconfigpay": true,
	"/client/systemconfig/getconfigbaseitem": true,
	"/client/ridingpermission/get": true,
	"/client/applicationsite/getconfig": true,
	
	// Misc
	"/business/test/postform": true,
	"/business/test/postjson": true,
}

// mirrorBodyCapture intercepts gin's ResponseWriter to capture the response body bytes
// while still letting the original response flow through to the real client.
type mirrorBodyCapture struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (m *mirrorBodyCapture) Write(b []byte) (int, error) {
	m.buf.Write(b)                   // capture
	return m.ResponseWriter.Write(b) // still forward to real client
}

// isMirrorableRequest returns true if the request is safe to mirror (read-only).
func isMirrorableRequest(method, path string) bool {
	// Never mirror destructive HTTP methods
	switch strings.ToUpper(method) {
	case http.MethodPut, http.MethodPatch, http.MethodDelete:
		return false
	}

	// GET is always safe
	if strings.ToUpper(method) == http.MethodGet {
		return true
	}

	// For POST: check exact whitelist
	lowerPath := strings.ToLower(path)
	return readOnlyEndpoints[lowerPath]
}

// matchRoutePattern checks if path matches an Ant-style pattern.
// Empty pattern matches all paths.
func matchRoutePattern(pattern, path string) bool {
	if pattern == "" {
		return true
	}
	p := strings.ReplaceAll(pattern, "**", "___DS___")
	p = strings.ReplaceAll(p, "*", "[^/]*")
	p = strings.ReplaceAll(p, "___DS___", ".*")
	matched, _ := regexp.MatchString("^"+p+"$", path)
	return matched
}

// mirrorClient is a package-level HTTP client reused across all mirror requests.
// Using a shared client ensures TCP connection pooling and avoids TIME_WAIT storms.
var (
	mirrorClient *http.Client
	mirrorOnce   sync.Once
)

// initMirrorClient creates the shared HTTP client with separate connect and read
// timeouts. The connect timeout (net.Dialer.Timeout) controls how long to wait
// for the TCP handshake — when the mirror target is down, the request fails fast
// within ConnectTimeoutMs instead of blocking the goroutine for the full read
// timeout. The read timeout (http.Client.Timeout) is the overall request deadline.
func initMirrorClient(cfg config.MirrorConfig) {
	mirrorOnce.Do(func() {
		transport := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   time.Duration(cfg.ConnectTimeoutMs) * time.Millisecond,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		}
		mirrorClient = &http.Client{
			Timeout:   time.Duration(cfg.ConnectTimeoutMs+cfg.ReadTimeoutMs) * time.Millisecond,
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		logger.Log.Info("[Mirror] HTTP client initialized",
			zap.Int("connectTimeoutMs", cfg.ConnectTimeoutMs),
			zap.Int("readTimeoutMs", cfg.ReadTimeoutMs),
		)
	})
}

// MirrorMiddleware captures the Java upstream response and asynchronously sends
// a mirror request to the Go shadow service (ebike-service-client-go).
//
// Design:
//  1. If mirroring is disabled or the request is write-type → pass through normally
//  2. Buffer the request body (needed to replay it to the shadow service)
//  3. Capture the Java upstream response body via mirrorBodyCapture
//  4. After the upstream responds, spawn a goroutine to send the mirror request
//     with X-Shadow-Java-Result = Base64(Java response body)
//  5. The Go shadow service's ShadowDiffMiddleware compares results and logs diffs
func MirrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.AppConfig.Mirror

		// Fast path: mirror disabled
		if !cfg.Enabled || cfg.Target == "" {
			c.Next()
			return
		}

		// Lazily initialize the shared mirror HTTP client
		initMirrorClient(cfg)

		path := c.Request.URL.Path

		// Check route pattern filter
		if !matchRoutePattern(cfg.RoutePattern, path) {
			c.Next()
			return
		}

		// Check if this request is safe to mirror (read-only)
		if !isMirrorableRequest(c.Request.Method, path) {
			c.Next()
			return
		}

		// Buffer the request body so it can be forwarded twice:
		// once to Java (already done by upstream proxy), once to Go shadow
		var bodyBytes []byte
		if c.Request.Body != nil {
			// Check if body was already buffered by ReadBodyMiddleware
			if cached, exists := c.Get("bodyBytes"); exists {
				bodyBytes = cached.([]byte)
			} else {
				var err error
				bodyBytes, err = io.ReadAll(c.Request.Body)
				if err != nil {
					logger.Log.Warn("[Mirror] Failed to read request body, skipping mirror",
						zap.String("path", path), zap.Error(err))
					c.Next()
					return
				}
				// Restore body for the upstream proxy
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				c.Set("bodyBytes", bodyBytes)
			}
		}

		// Install response body capture wrapper
		capture := &mirrorBodyCapture{
			ResponseWriter: c.Writer,
			buf:            &bytes.Buffer{},
		}
		c.Writer = capture

		// Process request normally (goes to Java service)
		c.Next()

		// After Java upstream has responded, fire mirror request asynchronously
		javaResponseBody := capture.buf.Bytes()
		if len(javaResponseBody) == 0 {
			return // Nothing to compare, skip
		}

		// Snapshot immutable values before handing off to goroutine
		mirrorTarget := cfg.Target
		mirrorPath := path
		mirrorMethod := c.Request.Method
		mirrorQuery := c.Request.URL.RawQuery
		mirrorHeaders := cloneHeaders(c.Request.Header)

		javaResultBase64 := base64.StdEncoding.EncodeToString(javaResponseBody)

		go sendMirrorRequest(
			mirrorTarget,
			mirrorMethod,
			mirrorPath,
			mirrorQuery,
			mirrorHeaders,
			bodyBytes,
			javaResultBase64,
		)
	}
}

// cloneHeaders copies the headers map (shallow copy of values) for safe goroutine use.
func cloneHeaders(src http.Header) http.Header {
	dst := make(http.Header, len(src))
	for k, vv := range src {
		dst[k] = append([]string(nil), vv...)
	}
	return dst
}

// sendMirrorRequest sends the mirror (shadow) request to the Go service.
// It attaches the Java response as X-Shadow-Java-Result so the Go service
// can compare results in its ShadowDiffMiddleware.
func sendMirrorRequest(
	target, method, path, query string,
	headers http.Header,
	bodyBytes []byte,
	javaResultBase64 string,
) {
	fullURL := target + path
	if query != "" {
		fullURL += "?" + query
	}

	var bodyReader io.Reader
	if len(bodyBytes) > 0 {
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		logger.Log.Warn("[Mirror] Failed to create mirror request",
			zap.String("url", fullURL), zap.Error(err))
		return
	}

	// Copy original headers (preserves authorities, Content-Type, Authorization, etc.)
	for k, vv := range headers {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	// Attach Java response for diff comparison in the Go service
	req.Header.Set("X-Shadow-Java-Result", javaResultBase64)

	resp, err := mirrorClient.Do(req)
	if err != nil {
		logger.Log.Warn("[Mirror] Mirror request failed",
			zap.String("url", fullURL), zap.Error(err))
		return
	}
	defer resp.Body.Close()
	// Drain body to allow connection reuse
	_, _ = io.Copy(io.Discard, resp.Body)

	logger.Log.Debug("[Mirror] Mirror request completed",
		zap.String("path", path),
		zap.Int("mirrorStatus", resp.StatusCode),
	)
}
