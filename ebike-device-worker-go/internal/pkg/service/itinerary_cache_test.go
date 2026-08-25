package service

import (
	"sync"
	"testing"
	"time"

	"ebike-device-worker-go/internal/api/dto"
	"ebike-device-worker-go/internal/config"
	"ebike-device-worker-go/internal/pkg/web"
)

func TestOrderTrajectoryCacheKey(t *testing.T) {
	config.GlobalConfig = &config.Config{}
	if got := orderTrajectoryCacheKey("123", 1, ""); got != "order_traj:123:g0:t1:i_" {
		t.Fatalf("empty imei key=%q", got)
	}
	if got := orderTrajectoryCacheKey("123", 0, "862551059811223"); got != "order_traj:123:g0:t0:i862551059811223" {
		t.Fatalf("imei key=%q", got)
	}
	itineraryCacheBumpGen("123")
	if got := orderTrajectoryCacheKey("123", 1, ""); got != "order_traj:123:g1:t1:i_" {
		t.Fatalf("bumped gen key=%q", got)
	}
}

func TestOrderTrajectoryL1NegativeCache(t *testing.T) {
	config.GlobalConfig = &config.Config{}
	orderID := "neg-" + t.Name()
	key := orderTrajectoryCacheKey(orderID, 1, "")
	storeOrderTrajectoryNegative(key)

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, ok := orderTrajNegL1.Get(key, nil); !ok {
				t.Error("expected L1 negative entry")
			}
		}()
	}
	wg.Wait()
}

func TestFetchOrderTrajectoryCoalescesInflight(t *testing.T) {
	config.GlobalConfig = &config.Config{}
	orderID := "coalesce-order"
	key := orderTrajectoryCacheKey(orderID, 1, "")

	inflight := &orderTrajInflight{done: make(chan struct{})}
	orderTrajLoads.Store(key, inflight)

	errCh := make(chan error, 1)
	go func() {
		_, err := fetchOrderTrajectory(orderID, 1, "")
		errCh <- err
	}()

	time.Sleep(20 * time.Millisecond)
	inflight.err = orderTrajectoryNotExistErr()
	close(inflight.done)
	orderTrajLoads.Delete(key)

	if err := <-errCh; err == nil {
		t.Fatal("expected not exist error")
	} else if biz, ok := err.(*web.BizError); !ok || biz.Code != dto.CodeOrderTrajectoryNotExist {
		t.Fatalf("unexpected err: %v", err)
	}
}
