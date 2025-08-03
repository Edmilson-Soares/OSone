package mqtt

import (
	"encoding/json"
	"fmt"
	"time"

	"go.etcd.io/bbolt"
)

type KV struct {
	db *bbolt.DB
}

type AuthPermission struct {
	Subscribers []string `json:"subscribers"`
	Publichers  []string `json:"publishers"`
}

type Device struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Virtual     string         `json:"virtual"`
	Code        string         `json:"code"`
	Password    string         `json:"password"`
	Permissions AuthPermission `json:"permissions"`
}

func (bk Broker) authKV() *KV {
	db, err := bbolt.Open("db/auth.db", 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil
	}
	db.Update(func(tx *bbolt.Tx) error {
		// Criar o bucket se não existir
		_, err := tx.CreateBucketIfNotExists([]byte("devices"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte("apps"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	//defer db.Close()
	return &KV{db: db}
}

func (kv KV) getDevice(id string) (Device, error) {
	var device Device
	err := kv.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("devices"))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}
		data := bucket.Get([]byte(id))
		if data == nil {
			return fmt.Errorf("auth not found")
		}
		return json.Unmarshal(data, &device)
	})
	if err != nil {
		return Device{}, err
	}

	return device, nil
}

func (kv KV) addDevice(device Device) error {
	return kv.db.Update(func(tx *bbolt.Tx) error {
		// Criar o bucket se não existir
		bucket := tx.Bucket([]byte("devices"))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}
		encoded, err := json.Marshal(device)
		if err != nil {
			return fmt.Errorf("marshal auth: %s", err)
		}
		return bucket.Put([]byte(device.ID), encoded)
	})

}

func (kv KV) delDevice(deviceId string) error {
	return kv.db.Update(func(tx *bbolt.Tx) error {
		// Criar o bucket se não existir
		bucket := tx.Bucket([]byte("devices"))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}
		return bucket.Delete([]byte(deviceId))
	})

}
