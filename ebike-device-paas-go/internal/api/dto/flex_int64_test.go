package dto

import (
	"encoding/json"
	"testing"
)

func TestInt64Slice_StringElements(t *testing.T) {
	var q DeviceListQry
	if err := json.Unmarshal([]byte(`{"serviceIdList":["239659900966803395"]}`), &q); err != nil {
		t.Fatal(err)
	}
	if len(q.ServiceIdList) != 1 || q.ServiceIdList[0] != 239659900966803395 {
		t.Fatalf("ServiceIdList = %v", q.ServiceIdList)
	}
}

func TestInt64Slice_NumberElements(t *testing.T) {
	var q DeviceListQry
	if err := json.Unmarshal([]byte(`{"serviceIdList":[239659900966803395]}`), &q); err != nil {
		t.Fatal(err)
	}
	if len(q.ServiceIdList) != 1 || q.ServiceIdList[0] != 239659900966803395 {
		t.Fatalf("ServiceIdList = %v", q.ServiceIdList)
	}
}

func TestFlexInt64_StringReportTime(t *testing.T) {
	var q DeviceListQry
	if err := json.Unmarshal([]byte(`{"reportTime":"1783658875716"}`), &q); err != nil {
		t.Fatal(err)
	}
	if q.ReportTime.Int64() != 1783658875716 {
		t.Fatalf("reportTime = %d", q.ReportTime.Int64())
	}
}

func TestTrajectoryHistoryQry_StringFields(t *testing.T) {
	var q TrajectoryHistoryQry
	raw := `{"orderId":"374148001472975036","startTime":"1783656640000","endTime":"1783665490000"}`
	if err := json.Unmarshal([]byte(raw), &q); err != nil {
		t.Fatal(err)
	}
	if q.OrderId.Int64() != 374148001472975036 {
		t.Fatalf("orderId = %d", q.OrderId.Int64())
	}
	if q.StartTime.Int64() != 1783656640000 || q.EndTime.Int64() != 1783665490000 {
		t.Fatalf("time range = %d-%d", q.StartTime.Int64(), q.EndTime.Int64())
	}
}

func TestDevicePageQry_StringServiceId(t *testing.T) {
	var q DevicePageQry
	if err := json.Unmarshal([]byte(`{"serviceId":"227587649840878077"}`), &q); err != nil {
		t.Fatal(err)
	}
	if q.ServiceId.Int64() != 227587649840878077 {
		t.Fatalf("serviceId = %d", q.ServiceId.Int64())
	}
}

func TestMetricQry_StringTimes(t *testing.T) {
	var q MetricQry
	raw := `{"imei":"869364082362526","startTime":"1783612800000","endTime":"1783665538654"}`
	if err := json.Unmarshal([]byte(raw), &q); err != nil {
		t.Fatal(err)
	}
	if q.StartTime.Int64() != 1783612800000 || q.EndTime.Int64() != 1783665538654 {
		t.Fatalf("time range = %d-%d", q.StartTime.Int64(), q.EndTime.Int64())
	}
}
