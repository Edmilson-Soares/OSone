package repos

import (
	"osone/utils"
)

func (driver *Driver) AddDevice(data utils.Device) (string, error) {

	input := Device{
		Name:      data.Name,
		ID:        data.ID,
		Icon:      data.Icon,
		Desc:      data.Desc,
		VirtualId: data.VirtualId,
		Location:  data.Location,
		Code:      data.Code,
		Config:    data.Config,
		Network:   data.Network,
		Auth:      data.Auth,
	}

	result := driver.db.Create(&input)
	if result.Error != nil {
		return "", result.Error
	}
	return input.ID, nil
}

func (driver *Driver) GetDevice(id string) (Device, error) {
	device := Device{}
	err := driver.db.First(&device).Where("id = ?", id).Error
	if err != nil {
		return device, err
	}
	return device, nil
}
func (driver *Driver) GetCodeDevice(code string) (Device, error) {
	device := Device{}
	err := driver.db.First(&device).Where("code = ?", code).Error
	if err != nil {
		return device, err
	}
	return device, nil
}
func (driver *Driver) UpdateDevice(device Device) error {
	return driver.db.Save(&device).Error

}

func (driver *Driver) DeleteDevice(id string) error {
	return driver.db.Unscoped().Delete(&Device{}, "id = ?", id).Error
}
