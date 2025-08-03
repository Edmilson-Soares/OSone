package repos

import (
	"osone/utils"
)

func (driver *Driver) AddVirtual(data utils.Virtual) (string, error) {

	input := Virtual{
		Name:         data.Name,
		ID:           data.ID,
		Icon:         data.Icon,
		Desc:         data.Desc,
		Auth:         data.Auth,
		EnterpriseId: data.EnterpriseId,
	}

	result := driver.db.Create(&input)
	if result.Error != nil {
		return "", result.Error
	}
	return input.ID, nil
}

func (driver *Driver) GetVirtual(id string) (Virtual, error) {
	space := Virtual{}
	err := driver.db.First(&space).Where("id = ?", id).Error
	if err != nil {
		return space, err
	}
	return space, nil
}
func (driver *Driver) GetVirtualView(id string) (Virtual, error) {
	space := Virtual{}
	err := driver.db.Preload("Devices").First(&space).Where("id = ?", id).Error
	if err != nil {
		return space, err
	}
	return space, nil
}
func (driver *Driver) UpdateVirtual(virtual Virtual) error {
	return driver.db.Save(&virtual).Error

}

func (driver *Driver) DeleteVirtual(id string) error {
	return driver.db.Unscoped().Delete(&Virtual{}, "id = ?", id).Error
}
