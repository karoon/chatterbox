package mqtt

type Storage interface {
	StorageSave() (MqttMessage, error)
	StorageGet() (MqttMessage, error)
}
