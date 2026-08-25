package persistence

import (
	"ebike-fence-go/internal/infrastructure/persistence/repo"
	"ebike-fence-go/internal/pkg/mysql"
)

var (
	Fence           *repo.FenceRepository
	Parking         *repo.ParkingRepository
	FenceRfid       *repo.FenceRfidRepository
	FenceCustomType *repo.FenceCustomTypeRepository
	AreaEmployee    *repo.AreaEmployeeRepository
	SiteApplication *repo.SiteApplicationRepository
	Config          *repo.ConfigRepository
)

// Init wires GORM repositories when MySQL is available.
func Init() {
	if mysql.DB == nil {
		return
	}
	Fence = repo.NewFenceRepository(mysql.DB)
	Parking = repo.NewParkingRepository(mysql.DB)
	FenceRfid = repo.NewFenceRfidRepository(mysql.DB)
	FenceCustomType = repo.NewFenceCustomTypeRepository(mysql.DB)
	AreaEmployee = repo.NewAreaEmployeeRepository(mysql.DB)
	SiteApplication = repo.NewSiteApplicationRepository(mysql.DB)
	Config = repo.NewConfigRepository(mysql.DB)
}
