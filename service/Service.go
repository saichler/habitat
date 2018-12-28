package service

import . "github.com/saichler/habitat"

type Service interface {
	HID() *HID
}
