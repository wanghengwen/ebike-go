package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"map-service-go/internal/cache"
	"map-service-go/internal/config"
	"map-service-go/internal/gateway"
	"map-service-go/internal/model"
	"map-service-go/internal/pkg/areacode"
	"map-service-go/internal/pkg/geohash"

	"github.com/sirupsen/logrus"
)

const regeoCacheTTL = 10 * 24 * time.Hour

type MapService struct {
	aMapGateway *gateway.AMapGatewayImpl
	tMapGateway *gateway.TMapGatewayImpl
}

func NewMapService() *MapService {
	return &MapService{
		aMapGateway: gateway.NewAMapGateway(),
		tMapGateway: gateway.NewTMapGateway(),
	}
}

func (s *MapService) getGateway(api string) gateway.MapApiGateway {
	if strings.EqualFold(api, "tMap") {
		return s.tMapGateway
	}
	return s.aMapGateway
}

func (s *MapService) GetAddressByLocation(ctx context.Context, cmd *model.MapCmd) (*model.AddressCo, error) {
	longitude, err := strconv.ParseFloat(cmd.Longitude, 64)
	if err != nil {
		return nil, err
	}
	latitude, err := strconv.ParseFloat(cmd.Latitude, 64)
	if err != nil {
		return nil, err
	}

	if longitude <= 0 || longitude >= 180 || latitude <= 0 || latitude >= 90 {
		return &model.AddressCo{
			Name:   "",
			Adcode: "0",
			Area:   "0",
		}, nil
	}

	return s.getAddressByLocationCached(ctx, cmd.Api, cmd.TenantId, cmd.Longitude, cmd.Latitude)
}

// GetBatchAddress mirrors Java AmapServiceImpl.getBatchAddress: call gateway directly
// with Redis cache (AOP equivalent) but without coordinate validation.
func (s *MapService) GetBatchAddress(ctx context.Context, cmd *model.MapCmd) ([]model.AddressCo, error) {
	addressCos := make([]model.AddressCo, 0)
	if cmd.Locations == nil {
		return addressCos, nil
	}
	for _, loc := range *cmd.Locations {
		addr, err := s.getAddressByLocationCached(ctx, cmd.Api, cmd.TenantId, loc.Longitude, loc.Latitude)
		if err != nil {
			logrus.WithContext(ctx).Errorf("GetBatchAddress failed: %v", err)
			return nil, err
		}
		if addr != nil {
			addressCos = append(addressCos, *addr)
		}
	}
	return addressCos, nil
}

func (s *MapService) getAddressByLocationCached(ctx context.Context, api, tenantId, longitude, latitude string) (*model.AddressCo, error) {
	date := time.Now().Format("20060102")
	redisCtx := ctx

	if cache.Rdb != nil {
		cache.Rdb.Incr(redisCtx, "addressCount_"+date)
		if addr := s.queryAddressRedis(redisCtx, longitude, latitude); addr != nil {
			cache.Rdb.Incr(redisCtx, "addressCacheCount_"+date)
			return addr, nil
		}
	}

	gw := s.getGateway(api)
	addr, err := gw.GetAddressByLocation(ctx, longitude, latitude, tenantId)
	if err != nil {
		logrus.WithContext(ctx).Errorf("GetAddressByLocation failed: %v", err)
		return nil, err
	}

	if addr != nil && cache.Rdb != nil {
		s.setAddressRedis(redisCtx, longitude, latitude, addr)
	}

	return addr, nil
}

func (s *MapService) GetLocationByAddress(ctx context.Context, cmd *model.MapCmd) (*model.Location, error) {
	gw := s.getGateway(cmd.Api)
	loc, err := gw.GetLocationByAddress(ctx, cmd.Address, cmd.City, cmd.TenantId)
	if err != nil {
		logrus.WithContext(ctx).Errorf("GetLocationByAddress failed: %v", err)
		return nil, err
	}
	return loc, nil
}

func (s *MapService) Navigate(ctx context.Context, cmd *model.MapCmd) (*model.Navigate, error) {
	gw := s.getGateway(cmd.Api)
	nav, err := gw.GetNavigate(ctx, cmd.Origin, cmd.Destination, cmd.Type, cmd.TenantId)
	if err != nil {
		logrus.WithContext(ctx).Errorf("Navigate failed: %v", err)
		return nil, err
	}
	return nav, nil
}

func (s *MapService) queryAddressRedis(ctx context.Context, lngStr, latStr string) *model.AddressCo {
	lat, _ := strconv.ParseFloat(latStr, 64)
	lng, _ := strconv.ParseFloat(lngStr, 64)

	geoLength := 8
	if config.GetConfig() != nil {
		geoLength = config.GetConfig().Xyy.GeoLength
	}

	gh := geohash.NewGeoHash(lat, lng)
	gh.SetHashLength(geoLength)

	hashBase32 := gh.GetGeoHashBase32()
	key := fmt.Sprintf("re_%s", hashBase32)

	val, err := cache.Rdb.Get(ctx, key).Result()
	if err == nil && val != "" {
		return s.transToAddressCo(val)
	}
	return nil
}

func (s *MapService) transToAddressCo(value string) *model.AddressCo {
	parts := strings.Split(value, "+")
	if len(parts) >= 2 {
		adCode := parts[0]
		area := areacode.GetFullAddress(adCode)
		name := area + parts[1]
		return &model.AddressCo{
			Name:   name,
			Area:   area,
			Adcode: adCode,
		}
	}
	return nil
}

func (s *MapService) setAddressRedis(ctx context.Context, lngStr, latStr string, addr *model.AddressCo) {
	lat, _ := strconv.ParseFloat(latStr, 64)
	lng, _ := strconv.ParseFloat(lngStr, 64)

	geoLength := 8
	if config.GetConfig() != nil {
		geoLength = config.GetConfig().Xyy.GeoLength
	}

	gh := geohash.NewGeoHash(lat, lng)
	gh.SetHashLength(geoLength)

	hashBase32 := gh.GetGeoHashBase32()
	key := fmt.Sprintf("re_%s", hashBase32)

	adCode := addr.Adcode
	name := strings.Replace(addr.Name, addr.Area, "", 1)
	value := adCode + "+" + name

	logrus.WithContext(ctx).Infof("geo redis key : %s, value :%s", key, value)
	cache.Rdb.Set(ctx, key, value, regeoCacheTTL)
}
