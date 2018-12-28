package service

import . "github.com/saichler/habitat"

type Service interface {
	SID() *SID
}
