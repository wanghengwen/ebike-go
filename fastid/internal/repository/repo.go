package repository

import (
	"luopingtech-fastid-register/internal/model"

	"gorm.io/gorm"
)

// FastIdRepo handles all database operations for the FastId registration service.
type FastIdRepo struct {
	DB *gorm.DB
}

// NewFastIdRepo creates a new repository instance.
func NewFastIdRepo(db *gorm.DB) *FastIdRepo {
	return &FastIdRepo{DB: db}
}

// FindAppByName queries the t_app table by namespace, group_id, and name.
// Returns nil, nil if no matching app is found.
func (r *FastIdRepo) FindAppByName(tx *gorm.DB, namespace, groupID, name string) (*model.App, error) {
	var app model.App
	result := tx.Where("namespace = ? AND group_id = ? AND name = ?", namespace, groupID, name).First(&app)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &app, nil
}

// CreateApp inserts a new App record.
func (r *FastIdRepo) CreateApp(tx *gorm.DB, app *model.App) error {
	return tx.Create(app).Error
}

// FindMachine queries the t_machine table by app_id and machine_uuid.
// Returns nil, nil if no matching machine is found.
func (r *FastIdRepo) FindMachine(tx *gorm.DB, appID uint, machineUUID string) (*model.Machine, error) {
	var machine model.Machine
	result := tx.Where("app_id = ? AND machine_uuid = ?", appID, machineUUID).First(&machine)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &machine, nil
}

// GetMaxMachineID returns IFNULL(MAX(machine_id), 0) for the given app, with a
// FOR UPDATE lock to prevent concurrent duplicate allocation.
// This matches the Java logic: SELECT IFNULL(max(machine_id),0) as machine_id ...
func (r *FastIdRepo) GetMaxMachineID(tx *gorm.DB, appID uint) (int, error) {
	var maxID int
	result := tx.Raw("SELECT IFNULL(MAX(machine_id), 0) FROM t_machine WHERE app_id = ? FOR UPDATE", appID).Scan(&maxID)
	if result.Error != nil {
		return 0, result.Error
	}
	return maxID, nil
}

// CreateMachine inserts a new Machine record.
func (r *FastIdRepo) CreateMachine(tx *gorm.DB, machine *model.Machine) error {
	return tx.Create(machine).Error
}

// ListApps returns all registered apps.
func (r *FastIdRepo) ListApps() ([]model.App, error) {
	var apps []model.App
	result := r.DB.Find(&apps)
	return apps, result.Error
}

// ListMachines returns all machines for the given app.
func (r *FastIdRepo) ListMachines(appID uint) ([]model.Machine, error) {
	var machines []model.Machine
	result := r.DB.Where("app_id = ?", appID).Find(&machines)
	return machines, result.Error
}
