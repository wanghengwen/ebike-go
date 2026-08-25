package geo

import (
	"encoding/json"
	"testing"
)

func TestMeter2DegreeJava100m(t *testing.T) {
	// Java: BigDecimal(100*180).divide(6371000-PI, 5, HALF_UP)
	got := Meter2DegreeJava(100)
	want := 0.00283
	if got != want {
		t.Fatalf("Meter2DegreeJava(100)=%v want %v", got, want)
	}
}

func TestCreatePolygonBufferAreaLianlianBuilding(t *testing.T) {
	// Raw parking polygon for 连连大厦 from SHADOW DIFF (Go side).
	raw := "[[120.184977,30.186938],[120.184977,30.187738],[120.185777,30.187738],[120.185777,30.186938]]"
	// Java createPolygonBufferArea result for bufferDistance=100, coefficient=0.
	javaWant := "[[120.184977,30.184108000000002],[120.18297588780925,30.184936887809243],[120.182147,30.186938],[120.182147,30.187738],[120.18297588780925,30.189739112190757],[120.184977,30.190568],[120.185777,30.190568],[120.18777811219076,30.189739112190757],[120.188607,30.187738],[120.188607,30.186938],[120.18777811219076,30.184936887809243],[120.185777,30.184108000000002]]"
	assertBufferMatchesJava(t, raw, 100, 0, javaWant)
}

func TestCreatePolygonBufferAreaYiJianHuanChe(t *testing.T) {
	raw := "[[120.184634,30.18793],[120.183937,30.190716],[120.18671,30.191085],[120.187279,30.188153]]"
	javaWant := "[[120.18487175385128,30.18511000476841],[120.18298790015871,30.18562798885483],[120.18188861269583,30.187243160462664],[120.18119161269583,30.19002916046266],[120.18159600560864,30.19230617145604],[120.18356370559421,30.193521272052156],[120.18633670559421,30.19389027205216],[120.18836217617022,30.193382653999752],[120.18948816862361,30.191624146639437],[120.19005716862361,30.188692146639436],[120.18953618976187,30.18644595741736],[120.18751675385128,30.18533300476841]]"
	assertBufferMatchesJava(t, raw, 100, 0, javaWant)
}

func assertBufferMatchesJava(t *testing.T, raw string, bufferDistance, coef float64, javaWant string) {
	t.Helper()
	got, err := CreatePolygonBufferArea(raw, bufferDistance, coef)
	if err != nil {
		t.Fatalf("CreatePolygonBufferArea: %v", err)
	}
	var gotPts, wantPts [][]float64
	if err := json.Unmarshal([]byte(got), &gotPts); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if err := json.Unmarshal([]byte(javaWant), &wantPts); err != nil {
		t.Fatalf("unmarshal want: %v", err)
	}
	if len(gotPts) != len(wantPts) {
		t.Fatalf("point count got=%d want=%d\ngot=%s", len(gotPts), len(wantPts), got)
	}
	const eps = 1e-9
	for i := range wantPts {
		if absf(gotPts[i][0]-wantPts[i][0]) > eps || absf(gotPts[i][1]-wantPts[i][1]) > eps {
			t.Fatalf("point[%d] got=%v want=%v\nfull got=%s", i, gotPts[i], wantPts[i], got)
		}
	}
}

func absf(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
