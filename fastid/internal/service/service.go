package service

import (
	"errors"
	"strings"

	"luopingtech-fastid-register/internal/model"
	"luopingtech-fastid-register/internal/repository"

	"gorm.io/gorm"
)

const (
	defaultNamespace = "PUBLIC"
	defaultGroupID   = "PUBLIC"
)

// IFastIdService defines the business operations for the FastId registration service.
// This interface allows mocking in unit tests.
type IFastIdService interface {
	GetMachineID(namespace, groupID, appName, machineUUID string) (int, error)
	ListApps() ([]model.App, error)
	ListMachines(appID uint) ([]model.Machine, error)
}

// FastIdService implements the machine ID registration logic.
// The GetMachineID method replicates the exact behavior of Java's FastidService.getMachineId.
type FastIdService struct {
	repo *repository.FastIdRepo
}

// NewFastIdService creates a new service instance.
func NewFastIdService(repo *repository.FastIdRepo) *FastIdService {
	return &FastIdService{repo: repo}
}

// GetMachineID registers a machine and returns its assigned machine ID.
//
// This method faithfully replicates the Java @Transactional getMachineId logic:
//  1. Validate appName and machineUUID are not blank.
//  2. Default namespace to "PUBLIC", groupID to "PUBLIC" if empty.
//  3. Find or create the App record.
//  4. Find existing machine by (app_id, machine_uuid); if found, return its machine_id.
//  5. Query IFNULL(MAX(machine_id), 0) with FOR UPDATE lock.
//  6. New machine_id = max + 1.
//  7. Insert the new Machine record.
func (s *FastIdService) GetMachineID(namespace, groupID, appName, machineUUID string) (int, error) {
	// Validate required parameters (matches Java: StringUtils.isBlank)
	if strings.TrimSpace(appName) == "" || strings.TrimSpace(machineUUID) == "" {
		return 0, errors.New("appName 和 machineUuid 都不能为空")
	}

	// Apply defaults (matches Java: StringUtils.isEmpty)
	if namespace == "" {
		namespace = defaultNamespace
	}
	if groupID == "" {
		groupID = defaultGroupID
	}

	var machineID int

	// Wrap everything in a database transaction (matches Java's @Transactional)
	err := s.repo.DB.Transaction(func(tx *gorm.DB) error {
		// Step 1: Find or create the App
		app, err := s.repo.FindAppByName(tx, namespace, groupID, appName)
		if err != nil {
			return err
		}
		if app == nil {
			app = &model.App{
				Namespace:  namespace,
				GroupID:    groupID,
				Name:       appName,
				CreateTime: model.Now(),
			}
			if err := s.repo.CreateApp(tx, app); err != nil {
				return err
			}
		}

		// Step 2: Check if this machine is already registered
		machine, err := s.repo.FindMachine(tx, app.ID, machineUUID)
		if err != nil {
			return err
		}
		if machine != nil {
			machineID = machine.MachineID
			return nil
		}

		// Step 3: Get the current max machine_id for this app (with lock)
		maxID, err := s.repo.GetMaxMachineID(tx, app.ID)
		if err != nil {
			return err
		}
		// Java logic: machineId = maxMachineId.getMachineId() + 1
		// where IFNULL(MAX(machine_id), 0) returns 0 for empty table,
		// so first machine gets id = 1.
		machineID = maxID + 1

		// Step 4: Insert the new machine record
		newMachine := &model.Machine{
			AppID:       app.ID,
			MachineID:   machineID,
			MachineUUID: machineUUID,
			CreateTime:  model.Now(),
			BeatTime:    model.Now(),
		}
		return s.repo.CreateMachine(tx, newMachine)
	})

	return machineID, err
}

// ListApps returns all registered applications.
func (s *FastIdService) ListApps() ([]model.App, error) {
	return s.repo.ListApps()
}

// ListMachines returns all machines registered under the given app ID.
func (s *FastIdService) ListMachines(appID uint) ([]model.Machine, error) {
	return s.repo.ListMachines(appID)
}
